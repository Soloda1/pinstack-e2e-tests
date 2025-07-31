package gateway_user

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUpdateAvatarTest(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up update avatar test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for update avatar test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Update avatar test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestUpdateAvatarSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupUpdateAvatarTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	avatarReq := fixtures.GenerateUpdateAvatarRequest()

	err := tc.UserClient.UpdateAvatar(*avatarReq)
	require.NoError(t, err)

	updatedUser, err := tc.UserClient.GetUserByID(userID)
	require.NoError(t, err)
	assert.Equal(t, avatarReq.AvatarURL, updatedUser.AvatarURL)
}

func TestUpdateAvatarUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, teardown := setupUpdateAvatarTest(t, tc)
	defer teardown()

	avatarReq := fixtures.GenerateUpdateAvatarRequest()

	t.Run("NoToken", func(t *testing.T) {
		tc.APIClient.SetToken("")

		err := tc.UserClient.UpdateAvatar(*avatarReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthenticated")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		tc.APIClient.SetToken("invalid_token")

		err := tc.UserClient.UpdateAvatar(*avatarReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})
}

func TestUpdateAvatarValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupUpdateAvatarTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		avatarURL   string
		expectedErr string
	}{
		{
			name:        "EmptyAvatarURL",
			avatarURL:   "",
			expectedErr: "validation failed",
		},
		{
			name:        "InvalidAvatarURL",
			avatarURL:   "not-a-url",
			expectedErr: "validation failed",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, teardown := setupUpdateAvatarTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			avatarReq := &fixtures.UpdateAvatarRequest{
				AvatarURL: tc.avatarURL,
			}

			err := ctx.UserClient.UpdateAvatar(*avatarReq)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestUpdateAvatarPermissionErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Create first user
	accessToken1, userID1, teardown1 := setupUpdateAvatarTest(t, tc)
	defer teardown1()

	// Create second user
	_, _, teardown2 := setupUpdateAvatarTest(t, tc)
	defer teardown2()

	t.Run("UpdateSelfUserAvatar", func(t *testing.T) {
		// User should be able to update their own avatar
		tc.APIClient.SetToken(accessToken1)

		avatarReq := fixtures.GenerateUpdateAvatarRequest()
		err := tc.UserClient.UpdateAvatar(*avatarReq)
		require.NoError(t, err)

		// Verify the avatar URL is updated
		updatedUser, err := tc.UserClient.GetUserByID(userID1)
		require.NoError(t, err)
		assert.Equal(t, avatarReq.AvatarURL, updatedUser.AvatarURL)
	})
}
