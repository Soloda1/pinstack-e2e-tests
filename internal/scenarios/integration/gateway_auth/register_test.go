package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRegisterTest(t *testing.T) (*fixtures.RegisterRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()

	log.Info("Setting up test", "test", t.Name(), "username", registerReq.Username)

	return registerReq, func() {
		log.Info("Test complete, local cleanup", "test", t.Name())
		apiClient.SetToken("")
	}
}

func TestRegisterSuccess(t *testing.T) {
	registerReq, teardown := setupRegisterTest(t)
	defer teardown()

	resp, err := authClient.Register(*registerReq)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	apiClient.SetToken(resp.AccessToken)
	userClient := client.NewUserClient(apiClient)

	user, err := userClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		trackUserForCleanup(user.ID, user.Username, resp.AccessToken)
	}
}

func TestRegisterInvalidInput(t *testing.T) {
	_, teardown := setupRegisterTest(t)
	defer teardown()

	t.Run("EmptyUsername", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Username = ""
		_, err := authClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("ShortUsername", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Username = "ab"
		_, err := authClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("LongUsername", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Username = "abcdefghijklmnopqrstuvwxyz1234567890"
		_, err := authClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Email = "invalid_email"
		_, err := authClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("ShortPassword", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Password = "12345"
		_, err := authClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}

func TestRegisterConflict(t *testing.T) {
	registerReq, teardown := setupRegisterTest(t)
	defer teardown()

	resp, err := authClient.Register(*registerReq)
	require.NoError(t, err)
	require.NotNil(t, resp)

	apiClient.SetToken(resp.AccessToken)
	userClient := client.NewUserClient(apiClient)

	user, err := userClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		trackUserForCleanup(user.ID, user.Username, resp.AccessToken)
	}

	apiClient.SetToken("")

	t.Run("DuplicateUsername", func(t *testing.T) {
		conflictReq := fixtures.GenerateRegisterRequest()
		conflictReq.Username = registerReq.Username
		conflictReq.Email = fixtures.GenerateRegisterRequest().Email
		_, err := authClient.Register(*conflictReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exists")
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		conflictReq := fixtures.GenerateRegisterRequest()
		conflictReq.Email = registerReq.Email
		conflictReq.Username = fixtures.GenerateRegisterRequest().Username
		_, err := authClient.Register(*conflictReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exists")
	})
}
