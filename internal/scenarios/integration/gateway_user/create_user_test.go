package gateway_user

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCreateUserTest(t *testing.T, tc *TestContext) (string, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up create user test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	}

	return tokens.AccessToken, func() {
		log.Info("Create user test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestCreateUserSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, teardown := setupCreateUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	createReq := fixtures.GenerateCreateUserRequest()

	createdUser, err := tc.UserClient.CreateUser(*createReq)
	require.NoError(t, err)
	require.NotNil(t, createdUser)

	tc.TrackUserForCleanup(createdUser.ID, createdUser.Username, accessToken)

	assert.Equal(t, createReq.Username, createdUser.Username)
	assert.Equal(t, createReq.Email, createdUser.Email)
	assert.Equal(t, createReq.FullName, createdUser.FullName)
	assert.Equal(t, createReq.Bio, createdUser.Bio)
	assert.NotZero(t, createdUser.ID)
	assert.NotEmpty(t, createdUser.CreatedAt)
	assert.NotEmpty(t, createdUser.UpdatedAt)
}

func TestCreateUserUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	t.Run("NoToken", func(t *testing.T) {
		createReq := fixtures.GenerateCreateUserRequest()

		_, err := tc.UserClient.CreateUser(*createReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		tc.APIClient.SetToken("invalid_token")

		createReq := fixtures.GenerateCreateUserRequest()

		_, err := tc.UserClient.CreateUser(*createReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error())
	})
}

func TestCreateUserValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, teardown := setupCreateUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		modifyReq   func(*fixtures.CreateUserRequest)
		expectedErr string
	}{
		{
			name: "EmptyUsername",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Username = ""
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "ShortUsername",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Username = "ab"
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "LongUsername",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Username = "abcdefghijklmnopqrstuvwxyz1234567890"
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "EmptyEmail",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Email = ""
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "InvalidEmail",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Email = "invalid_email"
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "EmptyPassword",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Password = ""
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "ShortPassword",
			modifyReq: func(req *fixtures.CreateUserRequest) {
				req.Password = "123"
			},
			expectedErr: custom_errors.ErrValidationFailed.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, teardown := setupCreateUserTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			createReq := fixtures.GenerateCreateUserRequest()
			tc.modifyReq(createReq)

			_, err := ctx.UserClient.CreateUser(*createReq)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestCreateUserConflictErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, teardown := setupCreateUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	createReq := fixtures.GenerateCreateUserRequest()

	createdUser, err := tc.UserClient.CreateUser(*createReq)
	require.NoError(t, err)
	tc.TrackUserForCleanup(createdUser.ID, createdUser.Username, accessToken)

	t.Run("DuplicateUsername", func(t *testing.T) {
		duplicateReq := fixtures.GenerateCreateUserRequest()
		duplicateReq.Username = createReq.Username

		_, err := tc.UserClient.CreateUser(*duplicateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrUsernameExists.Error())
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		duplicateReq := fixtures.GenerateCreateUserRequest()
		duplicateReq.Email = createReq.Email

		_, err := tc.UserClient.CreateUser(*duplicateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrEmailExists.Error())
	})
}
