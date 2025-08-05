package gateway_user

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetUserByIDTest(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get user by ID test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Get user by ID test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetUserByIDSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupGetUserByIDTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	user, err := tc.UserClient.GetUserByID(userID)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, userID, user.ID)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
}

func TestGetUserByIDNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupGetUserByIDTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	nonExistentUserID := int64(999999)
	_, err := tc.UserClient.GetUserByID(nonExistentUserID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error())
}

func TestGetUserByIDValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupGetUserByIDTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		id          int64
		expectedErr string
	}{
		{
			name:        "InvalidUserID",
			id:          -1,
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name:        "ZeroUserID",
			id:          0,
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, teardown := setupGetUserByIDTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			_, err := ctx.UserClient.GetUserByID(tc.id)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}
