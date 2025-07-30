package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRefreshTokenTest(t *testing.T) (*fixtures.RefreshTokenRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up refresh token test", "test", t.Name(), "username", registerReq.Username)

	registerResp, err := authClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	apiClient.SetToken(registerResp.AccessToken)
	userClient := client.NewUserClient(apiClient)

	user, err := userClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		trackUserForCleanup(user.ID, user.Username, registerResp.AccessToken)
	}

	apiClient.SetToken("")

	refreshReq := fixtures.GenerateRefreshTokenRequest(registerResp.RefreshToken)

	// Возвращаем запрос на обновление токена и функцию очистки
	return refreshReq, func() {
		log.Info("Refresh token test complete, local cleanup", "test", t.Name())
		apiClient.SetToken("")
	}
}

func TestRefreshTokenSuccess(t *testing.T) {
	refreshReq, teardown := setupRefreshTokenTest(t)
	defer teardown()

	resp, err := authClient.RefreshToken(*refreshReq)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestRefreshTokenInvalidToken(t *testing.T) {
	_, teardown := setupRefreshTokenTest(t)
	defer teardown()

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateRefreshTokenRequest("invalid_refresh_token")

		_, err := authClient.RefreshToken(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("EmptyRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateRefreshTokenRequest("")
		invalidReq.RefreshToken = ""

		_, err := authClient.RefreshToken(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}

func TestRefreshTokenAfterLogout(t *testing.T) {
	refreshReq, teardown := setupRefreshTokenTest(t)
	defer teardown()

	logoutReq := fixtures.GenerateLogoutRequest(refreshReq.RefreshToken)
	err := authClient.Logout(*logoutReq)
	require.NoError(t, err)

	_, err = authClient.RefreshToken(*refreshReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestRefreshTokenReuse(t *testing.T) {
	refreshReq, teardown := setupRefreshTokenTest(t)
	defer teardown()

	resp, err := authClient.RefreshToken(*refreshReq)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, err = authClient.RefreshToken(*refreshReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
