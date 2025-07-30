package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRefreshTokenTest(t *testing.T, tc *TestContext) (*fixtures.RefreshTokenRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up refresh token test", "test", t.Name(), "username", registerReq.Username)

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

	refreshReq := fixtures.GenerateRefreshTokenRequest(registerResp.RefreshToken)

	return refreshReq, func() {
		log.Info("Refresh token test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestRefreshTokenSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	refreshReq, teardown := setupRefreshTokenTest(t, tc)
	defer teardown()

	resp, err := tc.AuthClient.RefreshToken(*refreshReq)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestRefreshTokenInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, teardown := setupRefreshTokenTest(t, tc)
	defer teardown()

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateRefreshTokenRequest("invalid_refresh_token")

		_, err := tc.AuthClient.RefreshToken(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("EmptyRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateRefreshTokenRequest("")
		invalidReq.RefreshToken = ""

		_, err := tc.AuthClient.RefreshToken(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}

func TestRefreshTokenAfterLogout(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	refreshReq, teardown := setupRefreshTokenTest(t, tc)
	defer teardown()

	logoutReq := fixtures.GenerateLogoutRequest(refreshReq.RefreshToken)
	err := tc.AuthClient.Logout(*logoutReq)
	require.NoError(t, err)

	_, err = tc.AuthClient.RefreshToken(*refreshReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestRefreshTokenReuse(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	refreshReq, teardown := setupRefreshTokenTest(t, tc)
	defer teardown()

	resp, err := tc.AuthClient.RefreshToken(*refreshReq)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, err = tc.AuthClient.RefreshToken(*refreshReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
