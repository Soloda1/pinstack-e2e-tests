package gateway_notification

import (
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
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

type NotificationCleanupInfo struct {
	ID                   int64
	UserID               int64
	SenderAccessToken    string
	RecipientAccessToken string
}

type TestContext struct {
	APIClient            *client.Client
	AuthClient           *client.AuthClient
	UserClient           *client.UserClient
	NotificationClient   *client.NotificationClient
	CreatedUsers         []UserCleanupInfo
	CreatedNotifications []NotificationCleanupInfo
	mu                   sync.Mutex
}

func NewTestContext() *TestContext {
	apiClient := client.NewClient(cfg, log)
	return &TestContext{
		APIClient:            apiClient,
		AuthClient:           client.NewAuthClient(apiClient),
		UserClient:           client.NewUserClient(apiClient),
		NotificationClient:   client.NewNotificationClient(apiClient),
		CreatedUsers:         make([]UserCleanupInfo, 0),
		CreatedNotifications: make([]NotificationCleanupInfo, 0),
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

func (tc *TestContext) TrackNotificationForCleanup(notificationID, userID int64, senderAccessToken, recipientAccessToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	notificationInfo := NotificationCleanupInfo{
		ID:                   notificationID,
		UserID:               userID,
		SenderAccessToken:    senderAccessToken,
		RecipientAccessToken: recipientAccessToken,
	}
	tc.CreatedNotifications = append(tc.CreatedNotifications, notificationInfo)
	log.Debug("Added notification to cleanup list", "notification_id", notificationID, "user_id", userID)
}

func (tc *TestContext) Cleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Clean up notifications first
	if len(tc.CreatedNotifications) > 0 {
		log.Info("Starting notification cleanup process", "notifications_to_delete", len(tc.CreatedNotifications))

		var successfulNotificationDeletions int

		for _, notificationInfo := range tc.CreatedNotifications {
			log.Debug("Attempting to delete notification", "notification_id", notificationInfo.ID, "user_id", notificationInfo.UserID)

			tc.APIClient.SetToken(notificationInfo.RecipientAccessToken)

			resp, err := tc.NotificationClient.RemoveNotification(notificationInfo.ID)
			if err != nil {
				log.Warn("Failed to delete notification during cleanup",
					"notification_id", notificationInfo.ID,
					"user_id", notificationInfo.UserID,
					"error", err.Error())
			} else {
				log.Debug("Successfully deleted notification", "notification_id", notificationInfo.ID, "success", resp.Success)
				successfulNotificationDeletions++
			}
		}

		log.Info("Notification cleanup process completed",
			"successful_deletions", successfulNotificationDeletions,
			"total_notifications", len(tc.CreatedNotifications))
	}

	// Then clean up users
	if len(tc.CreatedUsers) > 0 {
		log.Info("Starting user cleanup process", "users_to_delete", len(tc.CreatedUsers))

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

		log.Info("User cleanup process completed",
			"successful_deletions", successfulUserDeletions,
			"total_users", len(tc.CreatedUsers))
	}

	tc.APIClient.SetToken("")
	tc.CreatedNotifications = []NotificationCleanupInfo{}
	tc.CreatedUsers = []UserCleanupInfo{}
}

// Helper function for setting up test users
func setupTestUser(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up test user", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		t.Fatalf("Failed to get user info for test: %v", err)
	}

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")
	log = logger.New(cfg.Env)
	log.Info("Starting notification gateway tests", "env", cfg.Env)

	code := m.Run()

	os.Exit(code)
}
