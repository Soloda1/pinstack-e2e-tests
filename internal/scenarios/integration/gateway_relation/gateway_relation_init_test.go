package gateway_relation

import (
	"flag"
	"os"
	"sync"
	"testing"
	"time"

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

type RelationCleanupInfo struct {
	FollowerID    int64
	FolloweeID    int64
	FollowerToken string
}

type NotificationCleanupInfo struct {
	ID                   int64
	UserID               int64
	RecipientAccessToken string
}

type TestContext struct {
	APIClient            *client.Client
	AuthClient           *client.AuthClient
	UserClient           *client.UserClient
	RelationClient       *client.RelationClient
	NotificationClient   *client.NotificationClient
	CreatedUsers         []UserCleanupInfo
	CreatedRelations     []RelationCleanupInfo
	CreatedNotifications []NotificationCleanupInfo
	mu                   sync.Mutex
}

func NewTestContext() *TestContext {
	apiClient := client.NewClient(cfg, log)
	return &TestContext{
		APIClient:            apiClient,
		AuthClient:           client.NewAuthClient(apiClient),
		UserClient:           client.NewUserClient(apiClient),
		RelationClient:       client.NewRelationClient(apiClient),
		NotificationClient:   client.NewNotificationClient(apiClient),
		CreatedUsers:         make([]UserCleanupInfo, 0),
		CreatedRelations:     make([]RelationCleanupInfo, 0),
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

func (tc *TestContext) TrackRelationForCleanup(followerID, followeeID int64, followerToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	relationInfo := RelationCleanupInfo{
		FollowerID:    followerID,
		FolloweeID:    followeeID,
		FollowerToken: followerToken,
	}
	tc.CreatedRelations = append(tc.CreatedRelations, relationInfo)
	log.Debug("Added relation to cleanup list", "follower_id", followerID, "followee_id", followeeID)
}

func (tc *TestContext) TrackNotificationForCleanup(notificationID, userID int64, recipientAccessToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	notificationInfo := NotificationCleanupInfo{
		ID:                   notificationID,
		UserID:               userID,
		RecipientAccessToken: recipientAccessToken,
	}
	tc.CreatedNotifications = append(tc.CreatedNotifications, notificationInfo)
	log.Debug("Added notification to cleanup list", "notification_id", notificationID, "user_id", userID)
}

func (tc *TestContext) DiscoverAndTrackAllNotifications(userID int64, userToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.APIClient.SetToken(userToken)

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(userID, 1, 100)
	if err != nil {
		log.Warn("Failed to get notification feed for notification discovery",
			"user_id", userID,
			"error", err.Error())
		return
	}

	for _, notification := range feedResp.Notifications {
		notificationInfo := NotificationCleanupInfo{
			ID:                   notification.ID,
			UserID:               notification.UserID,
			RecipientAccessToken: userToken,
		}
		tc.CreatedNotifications = append(tc.CreatedNotifications, notificationInfo)
		log.Debug("Discovered and tracked notification",
			"notification_id", notification.ID,
			"user_id", userID,
			"type", notification.Type)
	}
}

func (tc *TestContext) Cleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if len(tc.CreatedRelations) > 0 {
		log.Info("Starting relation cleanup process", "relations_to_delete", len(tc.CreatedRelations))

		var successfulRelationDeletions int

		for _, relationInfo := range tc.CreatedRelations {
			log.Debug("Attempting to unfollow relation",
				"follower_id", relationInfo.FollowerID,
				"followee_id", relationInfo.FolloweeID)

			tc.APIClient.SetToken(relationInfo.FollowerToken)

			resp, err := tc.RelationClient.Unfollow(relationInfo.FolloweeID)
			if err != nil {
				log.Warn("Failed to unfollow during cleanup",
					"follower_id", relationInfo.FollowerID,
					"followee_id", relationInfo.FolloweeID,
					"error", err.Error())
			} else {
				log.Debug("Successfully unfollowed relation",
					"follower_id", relationInfo.FollowerID,
					"followee_id", relationInfo.FolloweeID,
					"message", resp.Message)
				successfulRelationDeletions++
			}
		}

		log.Info("Relation cleanup process completed",
			"successful_deletions", successfulRelationDeletions,
			"total_relations", len(tc.CreatedRelations))
	}

	if len(tc.CreatedNotifications) > 0 {
		log.Info("Starting notification cleanup process", "notifications_to_delete", len(tc.CreatedNotifications))

		var successfulNotificationDeletions int

		for _, notificationInfo := range tc.CreatedNotifications {
			log.Debug("Attempting to delete notification",
				"notification_id", notificationInfo.ID,
				"user_id", notificationInfo.UserID)

			tc.APIClient.SetToken(notificationInfo.RecipientAccessToken)

			resp, err := tc.NotificationClient.RemoveNotification(notificationInfo.ID)
			if err != nil {
				log.Warn("Failed to delete notification during cleanup",
					"notification_id", notificationInfo.ID,
					"user_id", notificationInfo.UserID,
					"error", err.Error())
			} else {
				log.Debug("Successfully deleted notification",
					"notification_id", notificationInfo.ID,
					"success", resp.Success)
				successfulNotificationDeletions++
			}
		}

		log.Info("Notification cleanup process completed",
			"successful_deletions", successfulNotificationDeletions,
			"total_notifications", len(tc.CreatedNotifications))
	}

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
	tc.CreatedRelations = []RelationCleanupInfo{}
	tc.CreatedNotifications = []NotificationCleanupInfo{}
	tc.CreatedUsers = []UserCleanupInfo{}
}

var outboxTickInterval time.Duration

func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")
	log = logger.New(cfg.Env)
	outboxTickInterval = cfg.Outbox.TickInterval()
	log.Info("Starting relation gateway tests", "env", cfg.Env)

	code := m.Run()

	os.Exit(code)
}
