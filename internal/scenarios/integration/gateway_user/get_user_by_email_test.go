package gateway_user

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetUserByEmailTest(t *testing.T, tc *TestContext) (string, string, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get user by email test", "test", t.Name(), "username", registerReq.Username, "email", registerReq.Email)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, registerReq.Email, func() {
		log.Info("Get user by email test complete, local cleanup", "test", t.Name())
	}
}

func TestGetUserByEmailSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, userEmail, teardown := setupGetUserByEmailTest(t, tc)
	defer teardown()

	// Get user by email
	user, err := tc.UserClient.GetUserByEmail(userEmail)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Verify the response contains correct user data
	assert.Equal(t, userEmail, user.Email)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.ID)
}

func TestGetUserByEmailNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, teardown := setupGetUserByEmailTest(t, tc)
	defer teardown()

	// Try to get a non-existent user
	nonExistentEmail := "non.existent.user@example.com"
	_, err := tc.UserClient.GetUserByEmail(nonExistentEmail)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetUserByEmailValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, teardown := setupGetUserByEmailTest(t, tc)
	defer teardown()

	testCases := []struct {
		name        string
		email       string
		expectedErr string
	}{
		{
			name:        "EmptyEmail",
			email:       " ",
			expectedErr: "validation failed",
		},
		{
			name:        "InvalidEmail",
			email:       "not-an-email",
			expectedErr: "validation failed",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			_, _, teardown := setupGetUserByEmailTest(t, ctx)
			defer teardown()

			_, err := ctx.UserClient.GetUserByEmail(tc.email)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}
