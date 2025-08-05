package gateway_auth

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogoutTest(t *testing.T, tc *TestContext) (*fixtures.LogoutRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up logout test", "test", t.Name(), "username", registerReq.Username)

	registerResp, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	tc.APIClient.SetToken(registerResp.AccessToken)

	user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		tc.TrackUserForCleanup(user.ID, user.Username, registerResp.AccessToken)
	}

	logoutReq := fixtures.GenerateLogoutRequest(registerResp.RefreshToken)

	tc.APIClient.SetToken("")

	return logoutReq, func() {
		log.Info("Logout test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestLogoutSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	logoutReq, teardown := setupLogoutTest(t, tc)
	defer teardown()

	err := tc.AuthClient.Logout(*logoutReq)

	require.NoError(t, err)
}

func TestLogoutInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, teardown := setupLogoutTest(t, tc)
	defer teardown()

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateLogoutRequest("invalid_refresh_token")

		err := tc.AuthClient.Logout(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrInvalidRefreshToken.Error())
	})

	t.Run("EmptyRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateLogoutRequest("")
		invalidReq.RefreshToken = ""

		err := tc.AuthClient.Logout(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), custom_errors.ErrValidationFailed.Error())
	})
}

func TestLogoutTwice(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	logoutReq, teardown := setupLogoutTest(t, tc)
	defer teardown()

	err := tc.AuthClient.Logout(*logoutReq)
	require.NoError(t, err)

	err = tc.AuthClient.Logout(*logoutReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidRefreshToken.Error())
}
