package gateway_posts

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDeletePostTest(t *testing.T, tc *TestContext) (string, int64, int64, func()) {
	t.Helper()

	// Register a new user
	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up delete post test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for delete post test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	tc.APIClient.SetToken(tokens.AccessToken)

	// Create a post to delete
	postReq := fixtures.GenerateCreatePostRequest()
	createdPost, err := tc.PostClient.CreatePost(*postReq)
	require.NoError(t, err, "Failed to create test post")

	log.Info("Created test post for deletion", "post_id", createdPost.ID, "title", createdPost.Title)

	return tokens.AccessToken, userByUsername.ID, createdPost.ID, func() {
		log.Info("Delete post test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestDeletePostSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, postID, teardown := setupDeletePostTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	// Delete the post
	err := tc.PostClient.DeletePost(postID)
	require.NoError(t, err)

	// Verify post is deleted by trying to get it
	_, err = tc.PostClient.GetPostByID(postID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeletePostUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, postID, teardown := setupDeletePostTest(t, tc)
	defer teardown()

	// Add post to cleanup list since we won't delete it in the test
	tc.TrackPostForCleanup(postID, userID, accessToken)

	t.Run("NoToken", func(t *testing.T) {
		// Clear token
		tc.APIClient.SetToken("")

		// Try to delete post without authentication
		err := tc.PostClient.DeletePost(postID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthenticated")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		tc.APIClient.SetToken("invalid_token")

		// Try to delete post with invalid token
		err := tc.PostClient.DeletePost(postID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})
}

func TestDeletePostNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, postID, teardown := setupDeletePostTest(t, tc)
	defer teardown()

	// Add post to cleanup list since we won't delete it in the test
	tc.TrackPostForCleanup(postID, userID, accessToken)

	tc.APIClient.SetToken(accessToken)

	// Try to delete a non-existent post
	nonExistentPostID := int64(999999)
	err := tc.PostClient.DeletePost(nonExistentPostID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeletePostForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Create first user with a post
	accessToken1, userID1, postID, teardown1 := setupDeletePostTest(t, tc)
	defer teardown1()

	// Add post to cleanup since first user will fail to delete it later
	tc.TrackPostForCleanup(postID, userID1, accessToken1)

	// Create second user
	registerReq2 := fixtures.GenerateRegisterRequest()
	tokens2, err := tc.AuthClient.Register(*registerReq2)
	require.NoError(t, err)

	userByUsername2, err := tc.UserClient.GetUserByUsername(registerReq2.Username)
	require.NoError(t, err)
	userID2 := userByUsername2.ID

	tc.TrackUserForCleanup(userID2, userByUsername2.Username, tokens2.AccessToken)

	// Try to delete first user's post with second user's token
	tc.APIClient.SetToken(tokens2.AccessToken)

	err = tc.PostClient.DeletePost(postID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}
