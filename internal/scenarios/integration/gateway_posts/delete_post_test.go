package gateway_posts

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDeletePostTest(t *testing.T, tc *TestContext) (string, int64, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up delete post test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for delete post test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	tc.APIClient.SetToken(tokens.AccessToken)

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

	err := tc.PostClient.DeletePost(postID)
	require.NoError(t, err)

	_, err = tc.PostClient.GetPostByID(postID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrPostNotFound.Error())
}

func TestDeletePostUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, postID, teardown := setupDeletePostTest(t, tc)
	defer teardown()

	tc.TrackPostForCleanup(postID, userID, accessToken)

	t.Run("NoToken", func(t *testing.T) {
		tc.APIClient.SetToken("")

		err := tc.PostClient.DeletePost(postID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		tc.APIClient.SetToken("invalid_token")

		err := tc.PostClient.DeletePost(postID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error())
	})
}

func TestDeletePostNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, postID, teardown := setupDeletePostTest(t, tc)
	defer teardown()

	tc.TrackPostForCleanup(postID, userID, accessToken)

	tc.APIClient.SetToken(accessToken)

	nonExistentPostID := int64(999999)
	err := tc.PostClient.DeletePost(nonExistentPostID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrPostNotFound.Error())
}

func TestDeletePostForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken1, userID1, postID, teardown1 := setupDeletePostTest(t, tc)
	defer teardown1()

	tc.TrackPostForCleanup(postID, userID1, accessToken1)

	registerReq2 := fixtures.GenerateRegisterRequest()
	tokens2, err := tc.AuthClient.Register(*registerReq2)
	require.NoError(t, err)

	userByUsername2, err := tc.UserClient.GetUserByUsername(registerReq2.Username)
	require.NoError(t, err)
	userID2 := userByUsername2.ID

	tc.TrackUserForCleanup(userID2, userByUsername2.Username, tokens2.AccessToken)

	tc.APIClient.SetToken(tokens2.AccessToken)

	err = tc.PostClient.DeletePost(postID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrForbidden.Error())
}
