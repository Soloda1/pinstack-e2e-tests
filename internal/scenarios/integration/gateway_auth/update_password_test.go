package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUpdatePasswordTest(t *testing.T, tc *TestContext) (*fixtures.RegisterRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up update password test", "test", t.Name(), "username", registerReq.Username)

	registerResp, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")
	require.NotEmpty(t, registerResp.AccessToken, "Expected valid access token")

	tc.APIClient.SetToken(registerResp.AccessToken)

	user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		tc.TrackUserForCleanup(user.ID, user.Username, registerResp.AccessToken)
	}

	return registerReq, func() {
		log.Info("Update password test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestUpdatePasswordSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	registerReq, teardown := setupUpdatePasswordTest(t, tc)
	defer teardown()

	updateReq := fixtures.UpdatePasswordRequest{
		OldPassword: registerReq.Password,
		NewPassword: "NewPassword123!",
	}

	resp, err := tc.AuthClient.UpdatePassword(updateReq)

	require.NoError(t, err, "Should update password without error")
	require.NotNil(t, resp, "Expected non-nil response")
	assert.NotEmpty(t, resp.Message, "Expected success message")

	loginReq := fixtures.GenerateLoginRequest(registerReq.Username, "NewPassword123!")
	loginResp, err := tc.AuthClient.Login(*loginReq)

	require.NoError(t, err, "Should login with new password")
	require.NotNil(t, loginResp, "Expected non-nil login response")
	assert.NotEmpty(t, loginResp.AccessToken, "Expected valid access token with new password")
}

func TestUpdatePasswordValidation(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	registerReq, teardown := setupUpdatePasswordTest(t, tc)
	defer teardown()

	t.Run("EmptyOldPassword", func(t *testing.T) {
		updateReq := fixtures.UpdatePasswordRequest{
			OldPassword: "",
			NewPassword: "NewPassword123!",
		}

		_, err := tc.AuthClient.UpdatePassword(updateReq)
		assert.Error(t, err, "Should fail with empty old password")
		assert.Contains(t, err.Error(), "validation", "Expected validation error")
	})

	t.Run("EmptyNewPassword", func(t *testing.T) {
		updateReq := fixtures.UpdatePasswordRequest{
			OldPassword: registerReq.Password,
			NewPassword: "",
		}

		_, err := tc.AuthClient.UpdatePassword(updateReq)
		assert.Error(t, err, "Should fail with empty new password")
		assert.Contains(t, err.Error(), "validation", "Expected validation error")
	})

	t.Run("WeakNewPassword", func(t *testing.T) {
		updateReq := fixtures.UpdatePasswordRequest{
			OldPassword: registerReq.Password,
			NewPassword: "weak",
		}

		_, err := tc.AuthClient.UpdatePassword(updateReq)
		assert.Error(t, err, "Should fail with weak password")
		assert.Contains(t, err.Error(), "validation", "Expected validation error")
	})
}

func TestUpdatePasswordUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, teardown := setupUpdatePasswordTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	updateReq := fixtures.UpdatePasswordRequest{
		OldPassword: "AnyPassword123!",
		NewPassword: "NewPassword123!",
	}

	_, err := tc.AuthClient.UpdatePassword(updateReq)
	assert.Error(t, err, "Should fail without authorization")
	assert.Contains(t, err.Error(), "unauth", "Expected unauthorized error")
}

func TestUpdatePasswordWrongOldPassword(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, teardown := setupUpdatePasswordTest(t, tc)
	defer teardown()

	updateReq := fixtures.UpdatePasswordRequest{
		OldPassword: "WrongPassword123!",
		NewPassword: "NewPassword123!",
	}

	_, err := tc.AuthClient.UpdatePassword(updateReq)
	assert.Error(t, err, "Should fail with wrong old password")
	assert.Contains(t, err.Error(), "invalid", "Expected invalid old password error")
}
