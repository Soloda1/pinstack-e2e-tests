package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogoutTest(t *testing.T) (*fixtures.LogoutRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up logout test", "test", t.Name(), "username", registerReq.Username)

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

	logoutReq := fixtures.GenerateLogoutRequest(registerResp.RefreshToken)

	apiClient.SetToken("")

	// Возвращаем запрос на выход и функцию очистки
	return logoutReq, func() {
		log.Info("Logout test complete, local cleanup", "test", t.Name())
		apiClient.SetToken("")
	}
}

func TestLogoutSuccess(t *testing.T) {
	logoutReq, teardown := setupLogoutTest(t)
	defer teardown()

	err := authClient.Logout(*logoutReq)

	require.NoError(t, err)
}

func TestLogoutInvalidToken(t *testing.T) {
	_, teardown := setupLogoutTest(t)
	defer teardown()

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateLogoutRequest("invalid_refresh_token")

		err := authClient.Logout(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("EmptyRefreshToken", func(t *testing.T) {
		invalidReq := fixtures.GenerateLogoutRequest("")
		invalidReq.RefreshToken = ""

		err := authClient.Logout(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}

func TestLogoutTwice(t *testing.T) {
	logoutReq, teardown := setupLogoutTest(t)
	defer teardown()

	err := authClient.Logout(*logoutReq)
	require.NoError(t, err)

	err = authClient.Logout(*logoutReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
