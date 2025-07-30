package gateway_auth

import (
	"flag"
	"os"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
)

var (
	cfg        *config.Config
	log        *logger.Logger
	apiClient  *client.Client
	authClient *client.AuthClient

	createdUsers []UserCleanupInfo
)

type UserCleanupInfo struct {
	ID          int64
	Username    string
	AccessToken string
}

func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")

	log = logger.New(cfg.Env)
	log.Info("Starting auth gateway tests", "env", cfg.Env)

	apiClient = client.NewClient(cfg, log)
	authClient = client.NewAuthClient(apiClient)

	log.Info("Setup completed, starting tests")
	code := m.Run()

	if cfg.Test.Cleanup {
		log.Info("Tests finished, cleaning up test data")
		cleanup()
	}

	os.Exit(code)
}

func cleanup() {
	log.Info("Starting cleanup process", "users_to_delete", len(createdUsers))

	var successfulUserDeletions int

	userClient := client.NewUserClient(apiClient)

	for _, userInfo := range createdUsers {
		log.Debug("Attempting to delete user", "user_id", userInfo.ID, "username", userInfo.Username)

		apiClient.SetToken(userInfo.AccessToken)

		err := userClient.DeleteUser(userInfo.ID)
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

	apiClient.SetToken("")

	log.Info("Cleanup process completed",
		"successful_user_deletions", successfulUserDeletions,
		"total_users", len(createdUsers))

	createdUsers = []UserCleanupInfo{}
}

func trackUserForCleanup(userID int64, username, accessToken string) {
	userInfo := UserCleanupInfo{
		ID:          userID,
		Username:    username,
		AccessToken: accessToken,
	}
	createdUsers = append(createdUsers, userInfo)
	log.Debug("Added user to cleanup list", "user_id", userID, "username", username)
}
