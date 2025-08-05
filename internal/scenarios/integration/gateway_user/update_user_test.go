package gateway_user

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUpdateUserTest(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up update user test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for update test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Update user test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupUpdateUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	updateReq := fixtures.GenerateUpdateUserRequest(userID, "", "", "", "")

	response, err := tc.UserClient.UpdateUser(*updateReq)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, userID, response.ID)
	assert.Equal(t, updateReq.Username, response.Username)
	assert.Equal(t, updateReq.Email, response.Email)
	assert.Equal(t, updateReq.FullName, response.FullName)
	assert.Equal(t, updateReq.Bio, response.Bio)

	updatedUser, err := tc.UserClient.GetUserByID(userID)
	require.NoError(t, err)
	assert.Equal(t, updateReq.Username, updatedUser.Username)
	assert.Equal(t, updateReq.Email, updatedUser.Email)
	assert.Equal(t, updateReq.FullName, updatedUser.FullName)
	assert.Equal(t, updateReq.Bio, updatedUser.Bio)
}

func TestUpdateUserUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, userID, teardown := setupUpdateUserTest(t, tc)
	defer teardown()

	updateReq := fixtures.GenerateUpdateUserRequest(userID, "new_username", "", "", "")

	t.Run("NoToken", func(t *testing.T) {
		tc.APIClient.SetToken("")

		_, err := tc.UserClient.UpdateUser(*updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		tc.APIClient.SetToken("invalid_token")

		_, err := tc.UserClient.UpdateUser(*updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error())
	})
}

func TestUpdateUserValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupUpdateUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		id          int64
		username    string
		email       string
		fullName    string
		bio         string
		expectedErr string
	}{
		{
			name:        "InvalidUserID",
			id:          -1,
			username:    "valid_username",
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name:        "ZeroUserID",
			id:          0,
			username:    "valid_username",
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name:        "InvalidEmail",
			id:          userID,
			email:       "invalid-email",
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, teardown := setupUpdateUserTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			updateReq := fixtures.GenerateUpdateUserRequest(tc.id, tc.username, tc.email, tc.fullName, tc.bio)
			_, err := ctx.UserClient.UpdateUser(*updateReq)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestUpdateUserConflictErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken1, userID1, teardown1 := setupUpdateUserTest(t, tc)
	defer teardown1()

	registerReq2 := fixtures.GenerateRegisterRequest()
	tokens2, err := tc.AuthClient.Register(*registerReq2)
	require.NoError(t, err, "Failed to register second test user")

	user2, err := tc.UserClient.GetUserByUsername(registerReq2.Username)
	require.NoError(t, err, "Failed to get second user info")

	tc.TrackUserForCleanup(user2.ID, user2.Username, tokens2.AccessToken)

	t.Run("UsernameAlreadyExists", func(t *testing.T) {
		tc.APIClient.SetToken(accessToken1)

		updateReq := fixtures.GenerateUpdateUserRequest(userID1, registerReq2.Username, "", "", "")

		_, err := tc.UserClient.UpdateUser(*updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUsernameExists.Error())
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		tc.APIClient.SetToken(accessToken1)

		updateReq := fixtures.GenerateUpdateUserRequest(userID1, "", registerReq2.Email, "", "")

		_, err := tc.UserClient.UpdateUser(*updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrEmailExists.Error())
	})
}

func TestUpdateUserPermissionErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken1, userID1, teardown1 := setupUpdateUserTest(t, tc)
	defer teardown1()

	_, userID2, teardown2 := setupUpdateUserTest(t, tc)
	defer teardown2()

	t.Run("UpdateOtherUser", func(t *testing.T) {
		tc.APIClient.SetToken(accessToken1)

		updateReq := fixtures.GenerateUpdateUserRequest(userID2, "", "", "", "")

		_, err := tc.UserClient.UpdateUser(*updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrForbidden.Error())
	})

	t.Run("UpdateSelfUser", func(t *testing.T) {
		tc.APIClient.SetToken(accessToken1)

		updateReq := fixtures.GenerateUpdateUserRequest(userID1, "", "", "", "")
		response, err := tc.UserClient.UpdateUser(*updateReq)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, updateReq.Username, response.Username)
	})
}
