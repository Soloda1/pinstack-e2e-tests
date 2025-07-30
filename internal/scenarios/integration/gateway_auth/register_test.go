package gateway_auth

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRegisterTest(t *testing.T, tc *TestContext) (*fixtures.RegisterRequest, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()

	log.Info("Setting up test", "test", t.Name(), "username", registerReq.Username)

	return registerReq, func() {
		log.Info("Test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestRegisterSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	registerReq, teardown := setupRegisterTest(t, tc)
	defer teardown()

	resp, err := tc.AuthClient.Register(*registerReq)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	tc.APIClient.SetToken(resp.AccessToken)

	user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		tc.TrackUserForCleanup(user.ID, user.Username, resp.AccessToken)
	}
}

func TestRegisterInvalidInput(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, teardown := setupRegisterTest(t, tc)
	defer teardown()

	t.Run("EmptyUsername", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Username = ""
		_, err := tc.AuthClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("ShortUsername", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Username = "ab"
		_, err := tc.AuthClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("LongUsername", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Username = "abcdefghijklmnopqrstuvwxyz1234567890"
		_, err := tc.AuthClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Email = "invalid_email"
		_, err := tc.AuthClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("ShortPassword", func(t *testing.T) {
		invalidReq := fixtures.GenerateRegisterRequest()
		invalidReq.Password = "12345"
		_, err := tc.AuthClient.Register(*invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation")
	})
}

func TestRegisterConflict(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	registerReq, teardown := setupRegisterTest(t, tc)
	defer teardown()

	resp, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err)
	require.NotNil(t, resp)

	tc.APIClient.SetToken(resp.AccessToken)

	user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		log.Warn("Failed to get user info for cleanup tracking", "username", registerReq.Username, "error", err.Error())
	} else {
		tc.TrackUserForCleanup(user.ID, user.Username, resp.AccessToken)
	}

	tc.APIClient.SetToken("")

	t.Run("DuplicateUsername", func(t *testing.T) {
		conflictReq := fixtures.GenerateRegisterRequest()
		conflictReq.Username = registerReq.Username
		conflictReq.Email = fixtures.GenerateRegisterRequest().Email
		_, err := tc.AuthClient.Register(*conflictReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exists")
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		conflictReq := fixtures.GenerateRegisterRequest()
		conflictReq.Email = registerReq.Email
		conflictReq.Username = fixtures.GenerateRegisterRequest().Username
		_, err := tc.AuthClient.Register(*conflictReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exists")
	})
}
