package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoginTest(t *testing.T, tc *TestContext) (*fixtures.LoginRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up login test", "test", t.Name(), "username", registerReq.Username)

	registerResp, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	tc.APIClient.SetToken(registerResp.AccessToken)

	user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		tc.TrackUserForCleanup(user.ID, user.Username, registerResp.AccessToken)
	}

	tc.APIClient.SetToken("")

	loginReq := fixtures.GenerateLoginRequest(registerReq.Username, registerReq.Password)

	return loginReq, func() {
		log.Info("Login test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestLoginSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	loginReq, teardown := setupLoginTest(t, tc)
	defer teardown()

	resp, err := tc.AuthClient.Login(*loginReq)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestLoginInvalidCredentials(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	loginReq, teardown := setupLoginTest(t, tc)
	defer teardown()

	t.Run("WrongPassword", func(t *testing.T) {
		invalidReq := *loginReq
		invalidReq.Password = "wrong_password"

		_, err := tc.AuthClient.Login(invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		invalidReq := fixtures.GenerateLoginRequest("nonexistent_user_"+fixtures.GenerateRegisterRequest().Username, "password123")

		_, err := tc.AuthClient.Login(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestLoginValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, teardown := setupLoginTest(t, tc)
	defer teardown()

	t.Run("EmptyLogin", func(t *testing.T) {
		invalidReq := fixtures.GenerateLoginRequest("", "password123")
		invalidReq.Login = ""

		_, err := tc.AuthClient.Login(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		invalidReq := fixtures.GenerateLoginRequest("usernameemptypassword", "")
		invalidReq.Password = ""

		_, err := tc.AuthClient.Login(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("ShortPassword", func(t *testing.T) {
		invalidReq := fixtures.GenerateLoginRequest("usernameshortpassword", "12345")

		_, err := tc.AuthClient.Login(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}
