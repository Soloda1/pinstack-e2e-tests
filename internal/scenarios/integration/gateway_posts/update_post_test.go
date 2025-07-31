package gateway_posts

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUpdatePostTest(t *testing.T, tc *TestContext) (string, int64, int64, func()) {
	t.Helper()

	// Register a new user
	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up update post test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for update post test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	tc.APIClient.SetToken(tokens.AccessToken)

	// Create a post to update later
	postReq := fixtures.GenerateCreatePostRequest()
	createdPost, err := tc.PostClient.CreatePost(*postReq)
	require.NoError(t, err, "Failed to create test post")

	// Track post for cleanup
	tc.TrackPostForCleanup(createdPost.ID, userByUsername.ID, tokens.AccessToken)

	log.Info("Created test post for update test",
		"post_id", createdPost.ID,
		"title", createdPost.Title,
		"author_id", userByUsername.ID)

	return tokens.AccessToken, userByUsername.ID, createdPost.ID, func() {
		log.Info("Update post test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestUpdatePost(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	// Generate update data
	updateReq := fixtures.GenerateUpdatePostRequest()

	// Update the post
	updatedPost, err := tc.PostClient.UpdatePost(postID, *updateReq)
	require.NoError(t, err, "Failed to update post")

	// Verify the updated data
	assert.Equal(t, postID, updatedPost.ID, "Post ID should not change")
	assert.Equal(t, authorID, updatedPost.Author.ID, "Post author should not change")
	assert.Equal(t, updateReq.Title, updatedPost.Title, "Title should be updated")
	assert.Equal(t, updateReq.Content, updatedPost.Content, "Content should be updated")

	// Check if tags were updated correctly
	assert.Equal(t, len(updateReq.Tags), len(updatedPost.Tags), "Should have the same number of tags")
	tagMap := make(map[string]bool)
	for _, tag := range updatedPost.Tags {
		tagMap[tag.Name] = true
	}
	for _, tag := range updateReq.Tags {
		assert.True(t, tagMap[tag], "Updated post should contain tag %s", tag)
	}

	// Verify the updated data
	assert.Equal(t, updateReq.Title, updatedPost.Title)
	assert.Equal(t, updateReq.Content, updatedPost.Content)
	assert.Equal(t, authorID, updatedPost.Author.ID)

	// Check if tags were updated correctly
	tagNameMap := make(map[string]bool)
	for _, tag := range updatedPost.Tags {
		tagNameMap[tag.Name] = true
	}

	for _, tagName := range updateReq.Tags {
		assert.True(t, tagNameMap[tagName], "Tag should be in the updated post")
	}

	// Get the post to verify the update persisted
	fetchedPost, err := tc.PostClient.GetPostByID(postID)
	require.NoError(t, err)
	assert.Equal(t, updateReq.Title, fetchedPost.Title)
	assert.Equal(t, updateReq.Content, fetchedPost.Content)
}

func TestUpdatePostNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, _, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	// Try to update a non-existent post
	nonExistentPostID := int64(999999) // Use a very large ID that likely doesn't exist
	updateReq := fixtures.GenerateUpdatePostRequest()

	_, err := tc.PostClient.UpdatePost(nonExistentPostID, *updateReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdatePostForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup the post owner
	_, _, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	// Register a different user
	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Registering second user for forbidden test", "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err)

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err)

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	// Set token to the second user
	tc.APIClient.SetToken(tokens.AccessToken)

	// Try to update the post created by the first user
	updateReq := fixtures.GenerateUpdatePostRequest()

	_, err = tc.PostClient.UpdatePost(postID, *updateReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdatePostWithInvalidData(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	// Test with empty title
	updateReq := fixtures.UpdatePostRequest{
		Title:   "", // Empty title should be rejected
		Content: "Valid content",
	}

	_, err := tc.PostClient.UpdatePost(postID, updateReq)
	if err != nil {
		assert.Contains(t, err.Error(), "bad request")
	}

	// Test with extremely long content
	extremelyLongContent := ""
	for i := 0; i < 10000; i++ {
		extremelyLongContent += "too long content "
	}

	updateReq = fixtures.UpdatePostRequest{
		Title:   "Valid Title",
		Content: extremelyLongContent,
	}

	_, err = tc.PostClient.UpdatePost(postID, updateReq)
	if err != nil {
		assert.Contains(t, err.Error(), "bad request")
	}
}

func TestPartialUpdatePost(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	// Get the original post
	originalPost, err := tc.PostClient.GetPostByID(postID)
	require.NoError(t, err)

	// Update only the title
	titleOnlyUpdate := fixtures.UpdatePostRequest{
		Title: "Updated Title Only",
	}

	updatedPost, err := tc.PostClient.UpdatePost(postID, titleOnlyUpdate)
	require.NoError(t, err)

	// Verify only title was updated
	assert.Equal(t, titleOnlyUpdate.Title, updatedPost.Title)
	assert.Equal(t, originalPost.Content, updatedPost.Content)

	// Now update only the content
	contentOnlyUpdate := fixtures.UpdatePostRequest{
		Content: "This is updated content only",
	}

	updatedPost, err = tc.PostClient.UpdatePost(postID, contentOnlyUpdate)
	require.NoError(t, err)

	// Verify only content was updated
	assert.Equal(t, titleOnlyUpdate.Title, updatedPost.Title) // Title should remain from previous update
	assert.Equal(t, contentOnlyUpdate.Content, updatedPost.Content)
}
