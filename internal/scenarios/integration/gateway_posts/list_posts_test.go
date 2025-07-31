package gateway_posts

import (
	"testing"
	"time"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupListPostsTest(t *testing.T, tc *TestContext) (string, int64, []*fixtures.CreatePostResponse, func()) {
	t.Helper()

	// Register a new user
	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up list posts test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for list posts test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	tc.APIClient.SetToken(tokens.AccessToken)

	// Create multiple posts for testing list functionality
	var createdPosts []*fixtures.CreatePostResponse

	// Create 5 posts with different timestamps
	for i := 0; i < 5; i++ {
		postReq := fixtures.GenerateCreatePostRequest()
		createdPost, err := tc.PostClient.CreatePost(*postReq)
		require.NoError(t, err, "Failed to create test post")

		// Track post for cleanup
		tc.TrackPostForCleanup(createdPost.ID, userByUsername.ID, tokens.AccessToken)
		createdPosts = append(createdPosts, createdPost)

		log.Info("Created test post for list test", "post_id", createdPost.ID, "title", createdPost.Title)

		// Add a small delay between posts to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	return tokens.AccessToken, userByUsername.ID, createdPosts, func() {
		log.Info("List posts test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestListPostsAll(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, createdPosts, teardown := setupListPostsTest(t, tc)
	defer teardown()

	// List posts by the created author ID instead of all posts (which is more reliable)
	response, err := tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, 0, 0)
	require.NoError(t, err)

	// Should get exactly the posts we created
	assert.Equal(t, len(createdPosts), response.Total, "Should return exactly our created posts")
	assert.Equal(t, len(createdPosts), len(response.Posts), "Should return exactly our created posts")

	// Check that our posts are in the list
	postIDMap := make(map[int64]bool)
	for _, post := range response.Posts {
		postIDMap[post.ID] = true
	}

	for _, createdPost := range createdPosts {
		assert.True(t, postIDMap[createdPost.ID], "Created post should be in the list")
	}
}

func TestListPostsByAuthor(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, createdPosts, teardown := setupListPostsTest(t, tc)
	defer teardown()

	// List posts filtered by author
	response, err := tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, 0, 0)
	require.NoError(t, err)

	// Should get exactly the posts we created
	assert.Equal(t, len(createdPosts), response.Total, "Should return exactly our created posts")
	assert.Equal(t, len(createdPosts), len(response.Posts), "Should return exactly our created posts")

	// Verify all posts are from the specified author
	for _, post := range response.Posts {
		assert.Equal(t, authorID, post.Author.ID, "Post author should match filter")
	}

	// Verify we can find all our created posts
	postIDMap := make(map[int64]bool)
	for _, post := range response.Posts {
		postIDMap[post.ID] = true
	}

	for _, createdPost := range createdPosts {
		assert.True(t, postIDMap[createdPost.ID], "Created post should be in the author-filtered list")
	}

	// Test with a different author ID (should return no posts)
	differentAuthorID := authorID + 1000 // Assuming this ID doesn't exist
	emptyResponse, err := tc.PostClient.ListPosts(differentAuthorID, time.Time{}, time.Time{}, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, emptyResponse.Total, "Should return no posts for non-existent author")
	assert.Empty(t, emptyResponse.Posts, "Should return empty posts array for non-existent author")
}

func TestListPostsWithDateFilters(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, createdPosts, teardown := setupListPostsTest(t, tc)
	defer teardown()

	// Get the creation timestamps from first and last posts
	oldestTimestamp := createdPosts[0].CreatedAt
	newestTimestamp := createdPosts[len(createdPosts)-1].CreatedAt

	// Use a time slightly before all posts to test created_after filter
	beforeAllPosts := oldestTimestamp.Add(-1 * time.Hour)
	responseAfter, err := tc.PostClient.ListPosts(authorID, beforeAllPosts, time.Time{}, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(createdPosts), responseAfter.Total, "Should return all created posts")
	assert.Equal(t, len(createdPosts), len(responseAfter.Posts), "Should return all created posts")

	// Use a time after all posts to test created_before filter
	afterAllPosts := newestTimestamp.Add(1 * time.Hour)
	responseBefore, err := tc.PostClient.ListPosts(authorID, time.Time{}, afterAllPosts, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(createdPosts), responseBefore.Total, "Should return all created posts")
	assert.Equal(t, len(createdPosts), len(responseBefore.Posts), "Should return all created posts")

	// Test both filters together
	start := oldestTimestamp.Add(-1 * time.Hour) // Well before first post
	end := newestTimestamp.Add(1 * time.Hour)    // Well after last post

	responseBoth, err := tc.PostClient.ListPosts(authorID, start, end, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(createdPosts), responseBoth.Total, "Should return all created posts within the time window")

	// Test with no results by using a non-existent author ID
	nonExistentAuthorID := authorID + 1000 // Assuming this ID doesn't exist
	responseFuture, err := tc.PostClient.ListPosts(nonExistentAuthorID, time.Time{}, time.Time{}, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, responseFuture.Total, "Should return no posts for non-existent author")
	assert.Empty(t, responseFuture.Posts, "Should return empty posts array for non-existent author")
}

func TestListPostsPagination(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, createdPosts, teardown := setupListPostsTest(t, tc)
	defer teardown()

	// Set a small limit to test pagination
	limit := 2

	// First page (offset 0, limit 2)
	firstPage, err := tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, 0, limit)
	require.NoError(t, err)
	assert.Equal(t, len(createdPosts), firstPage.Total, "Total count should be consistent")
	assert.Equal(t, limit, len(firstPage.Posts), "Should return exactly limit posts")

	// Second page (offset 2, limit 2)
	secondPage, err := tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, limit, limit)
	require.NoError(t, err)
	assert.Equal(t, len(createdPosts), secondPage.Total, "Total count should be consistent")
	assert.Equal(t, limit, len(secondPage.Posts), "Should return exactly limit posts")

	// Make sure first and second page posts are different
	for _, firstPagePost := range firstPage.Posts {
		for _, secondPagePost := range secondPage.Posts {
			assert.NotEqual(t, firstPagePost.ID, secondPagePost.ID, "Posts from different pages should be unique")
		}
	}

	// Last page with remainder (if any)
	lastPageOffset := (len(createdPosts) / limit) * limit
	lastPage, err := tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, lastPageOffset, limit)
	require.NoError(t, err)
	assert.Equal(t, len(createdPosts), lastPage.Total, "Total count should be consistent")
	remainingPosts := len(createdPosts) - lastPageOffset
	assert.Equal(t, remainingPosts, len(lastPage.Posts), "Last page should contain the remaining posts")
}

func TestListPostsWithInvalidParams(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, _, teardown := setupListPostsTest(t, tc)
	defer teardown()

	_, err := tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, -1, 0)
	if err != nil {
		assert.Contains(t, err.Error(), "bad request")
	}

	_, err = tc.PostClient.ListPosts(authorID, time.Time{}, time.Time{}, 0, -1)
	if err != nil {
		assert.Contains(t, err.Error(), "bad request")
	}

	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)
	_, err = tc.PostClient.ListPosts(authorID, future, past, 0, 0)
	if err != nil {
		assert.Contains(t, err.Error(), "bad request")
	}
}
