package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUpdatePasswordTest(t *testing.T) (*fixtures.RegisterRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up update password test", "test", t.Name(), "username", registerReq.Username)

	registerResp, err := authClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")
	require.NotEmpty(t, registerResp.AccessToken, "Expected valid access token")

	apiClient.SetToken(registerResp.AccessToken)
	userClient := client.NewUserClient(apiClient)

	user, err := userClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		trackUserForCleanup(user.ID, user.Username, registerResp.AccessToken)
	}

	return registerReq, func() {
		log.Info("Update password test complete, local cleanup", "test", t.Name())
		apiClient.SetToken("")
	}
}

func TestUpdatePasswordSuccess(t *testing.T) {
	registerReq, teardown := setupUpdatePasswordTest(t)
	defer teardown()

	updateReq := fixtures.UpdatePasswordRequest{
		OldPassword: registerReq.Password,
		NewPassword: "NewPassword123!",
	}

	resp, err := authClient.UpdatePassword(updateReq)

	require.NoError(t, err, "Should update password without error")
	require.NotNil(t, resp, "Expected non-nil response")
	assert.NotEmpty(t, resp.Message, "Expected success message")

	loginReq := fixtures.GenerateLoginRequest(registerReq.Username, "NewPassword123!")
	loginResp, err := authClient.Login(*loginReq)

	require.NoError(t, err, "Should login with new password")
	require.NotNil(t, loginResp, "Expected non-nil login response")
	assert.NotEmpty(t, loginResp.AccessToken, "Expected valid access token with new password")
}

func TestUpdatePasswordValidation(t *testing.T) {
	registerReq, teardown := setupUpdatePasswordTest(t)
	defer teardown()

	t.Run("EmptyOldPassword", func(t *testing.T) {
		updateReq := fixtures.UpdatePasswordRequest{
			OldPassword: "",
			NewPassword: "NewPassword123!",
		}

		_, err := authClient.UpdatePassword(updateReq)
		assert.Error(t, err, "Should fail with empty old password")
		assert.Contains(t, err.Error(), "validation", "Expected validation error")
	})

	t.Run("EmptyNewPassword", func(t *testing.T) {
		updateReq := fixtures.UpdatePasswordRequest{
			OldPassword: registerReq.Password,
			NewPassword: "",
		}

		_, err := authClient.UpdatePassword(updateReq)
		assert.Error(t, err, "Should fail with empty new password")
		assert.Contains(t, err.Error(), "validation", "Expected validation error")
	})

	t.Run("WeakNewPassword", func(t *testing.T) {
		updateReq := fixtures.UpdatePasswordRequest{
			OldPassword: registerReq.Password,
			NewPassword: "weak",
		}

		_, err := authClient.UpdatePassword(updateReq)
		assert.Error(t, err, "Should fail with weak password")
		assert.Contains(t, err.Error(), "validation", "Expected validation error")
	})
}

func TestUpdatePasswordUnauthorized(t *testing.T) {
	_, teardown := setupUpdatePasswordTest(t)
	defer teardown()

	apiClient.SetToken("")

	updateReq := fixtures.UpdatePasswordRequest{
		OldPassword: "AnyPassword123!",
		NewPassword: "NewPassword123!",
	}

	_, err := authClient.UpdatePassword(updateReq)
	assert.Error(t, err, "Should fail without authorization")
	assert.Contains(t, err.Error(), "unauth", "Expected unauthorized error")
}

func TestUpdatePasswordWrongOldPassword(t *testing.T) {
	_, teardown := setupUpdatePasswordTest(t)
	defer teardown()

	updateReq := fixtures.UpdatePasswordRequest{
		OldPassword: "WrongPassword123!",
		NewPassword: "NewPassword123!",
	}

	_, err := authClient.UpdatePassword(updateReq)
	assert.Error(t, err, "Should fail with wrong old password")
	assert.Contains(t, err.Error(), "invalid", "Expected invalid old password error")
}
