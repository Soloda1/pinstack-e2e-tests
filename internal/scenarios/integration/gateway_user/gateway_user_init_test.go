package gateway_user

import (
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
)

var (
	cfg *config.Config
	log *logger.Logger
)

type UserCleanupInfo struct {
	ID          int64
	Username    string
	AccessToken string
}

type TestContext struct {
	APIClient    *client.Client
	AuthClient   *client.AuthClient
	UserClient   *client.UserClient
	CreatedUsers []UserCleanupInfo
	mu           sync.Mutex
}

func NewTestContext() *TestContext {
	apiClient := client.NewClient(cfg, log)
	return &TestContext{
		APIClient:    apiClient,
		AuthClient:   client.NewAuthClient(apiClient),
		UserClient:   client.NewUserClient(apiClient),
		CreatedUsers: make([]UserCleanupInfo, 0),
	}
}

func (tc *TestContext) TrackUserForCleanup(userID int64, username, accessToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	userInfo := UserCleanupInfo{
		ID:          userID,
		Username:    username,
		AccessToken: accessToken,
	}
	tc.CreatedUsers = append(tc.CreatedUsers, userInfo)
	log.Debug("Added user to cleanup list", "user_id", userID, "username", username)
}

func (tc *TestContext) Cleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if len(tc.CreatedUsers) == 0 {
		return
	}

	log.Info("Starting cleanup process", "users_to_delete", len(tc.CreatedUsers))

	var successfulUserDeletions int

	for _, userInfo := range tc.CreatedUsers {
		log.Debug("Attempting to delete user", "user_id", userInfo.ID, "username", userInfo.Username)

		tc.APIClient.SetToken(userInfo.AccessToken)

		err := tc.UserClient.DeleteUser(userInfo.ID)
		if err != nil {
			log.Warn("Failed to delete user during cleanup",
				"user_id", userInfo.ID,
				"username", userInfo.Username,
				"error", err.Error())
		} else {
			log.Debug("Successfully deleted user", "user_id", userInfo.ID, "username", userInfo.Username)
			successfulUserDeletions++
		}
	}

	tc.APIClient.SetToken("")

	log.Info("Cleanup process completed",
		"successful_user_deletions", successfulUserDeletions,
		"total_users", len(tc.CreatedUsers))

	tc.CreatedUsers = []UserCleanupInfo{}
}

func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")
	log = logger.New(cfg.Env)
	log.Info("Starting user gateway tests", "env", cfg.Env)

	code := m.Run()
	os.Exit(code)
}
