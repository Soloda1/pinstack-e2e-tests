package gateway_posts

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetPostTest(t *testing.T, tc *TestContext) (string, int64, int64, *fixtures.CreatePostRequest, *fixtures.CreatePostResponse, func()) {
	t.Helper()

	// Register a new user
	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get post test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for get post test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	tc.APIClient.SetToken(tokens.AccessToken)

	// Create a post to retrieve
	postReq := fixtures.GenerateCreatePostRequest()
	createdPost, err := tc.PostClient.CreatePost(*postReq)
	require.NoError(t, err, "Failed to create test post")

	// Track post for cleanup
	tc.TrackPostForCleanup(createdPost.ID, userByUsername.ID, tokens.AccessToken)

	log.Info("Created test post for get test", "post_id", createdPost.ID, "title", createdPost.Title)

	return tokens.AccessToken, userByUsername.ID, createdPost.ID, postReq, createdPost, func() {
		log.Info("Get post test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetPostByIDSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, postID, postReq, createdPost, teardown := setupGetPostTest(t, tc)
	defer teardown()

	// Get the post by ID
	retrievedPost, err := tc.PostClient.GetPostByID(postID)
	require.NoError(t, err)

	// Validate retrieved post data
	assert.Equal(t, postID, retrievedPost.ID, "Retrieved post ID should match")
	assert.Equal(t, postReq.Title, retrievedPost.Title, "Retrieved post title should match")
	assert.Equal(t, postReq.Content, retrievedPost.Content, "Retrieved post content should match")
	assert.Equal(t, createdPost.AuthorID, retrievedPost.Author.ID, "Retrieved post author ID should match")
	assert.Equal(t, createdPost.AuthorUsername, retrievedPost.Author.Username, "Retrieved post author username should match")

	// Validate media items
	assert.Equal(t, len(postReq.MediaItems), len(retrievedPost.Media), "Number of media items should match")
	if len(postReq.MediaItems) > 0 {
		mediaMap := make(map[string]bool)
		for _, item := range retrievedPost.Media {
			mediaMap[item.URL] = true
		}
		for _, requestItem := range postReq.MediaItems {
			assert.True(t, mediaMap[requestItem.URL], "Media URL should be present in response")
		}
	}

	// Validate tags
	assert.Equal(t, len(postReq.Tags), len(retrievedPost.Tags), "Number of tags should match")
	if len(postReq.Tags) > 0 {
		tagMap := make(map[string]bool)
		for _, tag := range retrievedPost.Tags {
			tagMap[tag.Name] = true
		}
		for _, requestTag := range postReq.Tags {
			assert.True(t, tagMap[requestTag], "Tag name should be present in response")
		}
	}
}

func TestGetPostByIDNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Try to get a non-existent post
	nonExistentPostID := int64(999999)
	_, err := tc.PostClient.GetPostByID(nonExistentPostID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetPostByIDDeletedPost(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, postID, _, _, teardown := setupGetPostTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	// Delete the post
	err := tc.PostClient.DeletePost(postID)
	require.NoError(t, err, "Failed to delete post for test")

	// Try to get the deleted post
	_, err = tc.PostClient.GetPostByID(postID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
