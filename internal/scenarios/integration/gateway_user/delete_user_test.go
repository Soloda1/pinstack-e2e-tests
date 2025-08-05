package gateway_user

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDeleteUserTest(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up delete user test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for delete test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Delete user test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestDeleteUserSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupDeleteUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	err := tc.UserClient.DeleteUser(userID)
	require.NoError(t, err)

	_, err = tc.UserClient.GetUserByID(userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error())
}

func TestDeleteUserUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, userID, teardown := setupDeleteUserTest(t, tc)
	defer teardown()

	t.Run("NoToken", func(t *testing.T) {
		tc.APIClient.SetToken("")
		err := tc.UserClient.DeleteUser(userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		tc.APIClient.SetToken("invalid_token")

		err := tc.UserClient.DeleteUser(userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error())
	})
}

func TestDeleteUserValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupDeleteUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		userID      int64
		expectedErr string
	}{
		{
			name:        "InvalidUserID",
			userID:      -1,
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name:        "ZeroUserID",
			userID:      0,
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, teardown := setupDeleteUserTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			err := ctx.UserClient.DeleteUser(tc.userID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestDeleteUserPermissionErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken1, userID1, teardown1 := setupDeleteUserTest(t, tc)
	defer teardown1()

	_, userID2, teardown2 := setupDeleteUserTest(t, tc)
	defer teardown2()

	t.Run("DeleteOtherUser", func(t *testing.T) {
		tc.APIClient.SetToken(accessToken1)

		err := tc.UserClient.DeleteUser(userID2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrForbidden.Error())
	})

	t.Run("DeleteSelfUser", func(t *testing.T) {
		tc.APIClient.SetToken(accessToken1)

		err := tc.UserClient.DeleteUser(userID1)
		require.NoError(t, err)

		_, err = tc.UserClient.GetUserByID(userID1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error())
	})
}
