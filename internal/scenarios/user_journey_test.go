package scenarios

import (
	"flag"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
)

var (
	cfg            *config.Config
	log            *logger.Logger
	apiClient      *client.Client
	authClient     *client.AuthClient
	userClient     *client.UserClient
	postClient     *client.PostClient
	relationClient *client.RelationClient
	notifClient    *client.NotificationClient
)

// setup prepares the test environment for a single test by registering a new user
func setup(t *testing.T) (*fixtures.UserJourney, string, string, func()) {
	t.Helper()

	// Create a new UserJourney for each test to isolate test data
	journey := fixtures.CreateUserJourney()

	log.Info("Setting up test", "test", t.Name(), "username", journey.RegisterRequest.Username)

	resp, err := authClient.Register(*journey.RegisterRequest)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	apiClient.SetToken(resp.AccessToken)

	// Return the journey, tokens, and a cleanup function
	return journey, resp.AccessToken, resp.RefreshToken, func() {
		log.Info("Test complete, local cleanup", "test", t.Name())
		// Add test-specific cleanup code here if needed
		apiClient.SetToken("")
	}
}

// TestMain runs before any test in the package and sets up the testing environment
func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad()

	log = logger.New(cfg.Env)
	log.Info("Starting user journey e2e tests", "env", cfg.Env)

	apiClient = client.NewClient(cfg, log)
	authClient = client.NewAuthClient(apiClient)
	userClient = client.NewUserClient(apiClient)
	postClient = client.NewPostClient(apiClient)
	relationClient = client.NewRelationClient(apiClient)
	notifClient = client.NewNotificationClient(apiClient)

	// Run the tests
	log.Info("Setup completed, starting tests")
	code := m.Run()

	// Clean up
	if cfg.Test.Cleanup {
		log.Info("Tests finished, cleaning up test data")
		cleanup()
	}

	os.Exit(code)
}

// cleanup handles test data cleanup after all tests
func cleanup() {
	log.Info("Cleanup not implemented")
}

// TestUserJourney tests the complete user journey from registration to usage
func TestUserJourney(t *testing.T) {
	t.Run("1. Registration and Login", testUserRegistrationAndLogin)
	t.Run("2. Profile Management", testUserProfileManagement)
	t.Run("3. Post Creation", testPostCreation)
	t.Run("4. Following Users", testFollowingUsers)
	t.Run("5. Notifications", testNotifications)
}

// TestUserRegistrationAndLogin tests the user registration and login process
func testUserRegistrationAndLogin(t *testing.T) {
	journey, accessToken, refreshToken, teardown := setup(t)
	defer teardown()

	if accessToken == "" {
		t.Fatal("Expected a valid access token after registration")
	}
	if refreshToken == "" {
		t.Fatal("Expected a valid refresh token after registration")
	}

	// Clear the token to simulate a new session
	apiClient.SetToken("")

	// Test login with the registered credentials
	loginReq := fixtures.GenerateLoginRequest(journey.RegisterRequest.Username, journey.RegisterRequest.Password)
	loginResp, err := authClient.Login(*loginReq)
	if err != nil {
		t.Fatalf("Failed to login with registered user: %v", err)
	}

	// Verify login was successful
	if loginResp.AccessToken == "" {
		t.Fatal("Expected a valid access token after login")
	}
	if loginResp.RefreshToken == "" {
		t.Fatal("Expected a valid refresh token after login")
	}

	// Set the token for subsequent API calls
	apiClient.SetToken(loginResp.AccessToken)

	log.Info("Successfully completed registration and login test",
		"username", journey.RegisterRequest.Username)
}

// testUserProfileManagement tests retrieving, updating, and managing user profile
func testUserProfileManagement(t *testing.T) {
	journey, _, _, teardown := setup(t)
	defer teardown()

	user, err := userClient.GetUserByUsername(journey.RegisterRequest.Username)
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}

	// Verify profile data matches what we registered
	if user.Username != journey.RegisterRequest.Username {
		t.Errorf("Expected username %s, got %s", journey.RegisterRequest.Username, user.Username)
	}
	if user.Email != journey.RegisterRequest.Email {
		t.Errorf("Expected email %s, got %s", journey.RegisterRequest.Email, user.Email)
	}
	if user.FullName == "" {
		t.Errorf("Expected full name to be set, got empty string")
	} else if user.FullName != journey.RegisterRequest.FullName {
		// Log a warning instead of failing the test, as the API may transform the full name
		t.Logf("Note: Full name doesn't match exactly. Expected %s, got %s", journey.RegisterRequest.FullName, user.FullName)
	}

	// Update user profile
	updatedBio := "Updated bio for e2e testing"
	updateReq := fixtures.UpdateUserRequest{
		ID:  user.ID,
		Bio: updatedBio,
	}

	updatedUser, err := userClient.UpdateUser(updateReq)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify the update was successful
	if updatedUser.Bio != updatedBio {
		t.Errorf("Expected updated bio %s, got %s", updatedBio, updatedUser.Bio)
	}

	// Update avatar
	newAvatarURL := "https://example.com/new-avatar.jpg"
	avatarReq := fixtures.UpdateAvatarRequest{
		AvatarURL: newAvatarURL,
	}

	err = userClient.UpdateAvatar(user.ID, avatarReq)
	if err != nil {
		t.Fatalf("Failed to update avatar: %v", err)
	}

	// Get user again to verify avatar update
	updatedUserWithAvatar, err := userClient.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUserWithAvatar.AvatarURL != newAvatarURL {
		t.Errorf("Expected avatar URL %s, got %s", newAvatarURL, updatedUserWithAvatar.AvatarURL)
	}

	log.Info("Successfully completed profile management test",
		"username", journey.RegisterRequest.Username,
		"user_id", user.ID)
}

// testPostCreation tests creating and retrieving posts
func testPostCreation(t *testing.T) {
	_, _, _, teardown := setup(t)
	defer teardown()

	postReq := fixtures.GenerateCreatePostRequest()

	createdPost, err := postClient.CreatePost(*postReq)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Verify the post was created with correct data
	if createdPost.Title != postReq.Title {
		t.Errorf("Expected post title %s, got %s", postReq.Title, createdPost.Title)
	}

	if createdPost.Content != postReq.Content {
		t.Errorf("Expected post content %s, got %s", postReq.Content, createdPost.Content)
	}

	// Verify media items
	if len(createdPost.Media) != len(postReq.MediaItems) {
		t.Errorf("Expected %d media items, got %d", len(postReq.MediaItems), len(createdPost.Media))
	}

	// Verify tags
	if len(createdPost.Tags) != len(postReq.Tags) {
		t.Errorf("Expected %d tags, got %d", len(postReq.Tags), len(createdPost.Tags))
	}

	// Get the post by ID
	retrievedPost, err := postClient.GetPostByID(createdPost.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve post: %v", err)
	}

	// Verify retrieved post data
	if retrievedPost.ID != createdPost.ID {
		t.Errorf("Expected post ID %d, got %d", createdPost.ID, retrievedPost.ID)
	}

	if retrievedPost.Title != createdPost.Title {
		t.Errorf("Expected post title %s, got %s", createdPost.Title, retrievedPost.Title)
	}

	// List posts by author
	listResp, err := postClient.ListPosts(createdPost.AuthorID, time.Time{}, time.Time{}, 0, 10)
	if err != nil {
		t.Fatalf("Failed to list posts: %v", err)
	}

	// Verify the post is in the list
	found := false
	for _, post := range listResp.Posts {
		if post.ID == createdPost.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created post not found in the list of posts")
	}

	log.Info("Successfully completed post creation test",
		"post_id", createdPost.ID,
		"title", createdPost.Title)
}

// testFollowingUsers tests following and unfollowing users
func testFollowingUsers(t *testing.T) {
	journey, accessToken, _, teardown := setup(t)
	defer teardown()

	firstUserToken := accessToken

	// Register another test user to follow/unfollow
	otherJourney, _, _, otherTeardown := setup(t)
	defer otherTeardown()

	// Restore first user's token before calling Follow
	apiClient.SetToken(firstUserToken)

	// Get the ID of the user to follow
	otherUser, err := userClient.GetUserByUsername(otherJourney.RegisterRequest.Username)
	if err != nil {
		t.Fatalf("Failed to get other user: %v", err)
	}

	// Follow the other user
	followResp, err := relationClient.Follow(otherUser.ID)
	if err != nil {
		t.Fatalf("Failed to follow user: %v", err)
	}

	if followResp.Message == "" {
		t.Error("Expected a success message in follow response")
	}

	log.Info("Successfully followed user",
		"follower", journey.RegisterRequest.Username,
		"followee", otherUser.Username)

	// Ensure first user's token is still active before calling Unfollow
	apiClient.SetToken(firstUserToken)

	// Unfollow the user
	unfollowResp, err := relationClient.Unfollow(otherUser.ID)
	if err != nil {
		t.Fatalf("Failed to unfollow user: %v", err)
	}

	if unfollowResp.Message == "" {
		t.Error("Expected a success message in unfollow response")
	}

	log.Info("Successfully unfollowed user",
		"follower", journey.RegisterRequest.Username,
		"followee", otherUser.Username)
}

// testNotifications tests sending and receiving notifications
func testNotifications(t *testing.T) {
	journey, _, _, teardown := setup(t)
	defer teardown()

	user, err := userClient.GetUserByUsername(journey.RegisterRequest.Username)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	notifPayload := map[string]interface{}{
		"message": "Test notification for e2e testing",
	}

	notifReq := fixtures.SendNotificationRequest{
		UserID:  user.ID,
		Type:    fixtures.NotificationTypeSystem,
		Payload: notifPayload,
	}

	// Send the notification
	sendResp, err := notifClient.SendNotification(notifReq)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	var notificationID int64

	// Log notification ID and API response
	log.Info("Notification sent",
		slog.String("message", sendResp.Message),
		slog.Any("response", sendResp))

	// Get notification feed
	feedResp, err := notifClient.GetUserNotificationFeed(user.ID, 1, 10)
	if err != nil {
		t.Fatalf("Failed to get notification feed: %v", err)
	}

	// Verify that the feed is not empty
	if len(feedResp.Notifications) == 0 {
		t.Error("Expected at least one notification in feed after sending")
	} else {
		// Get the latest notification for further checks
		latestNotif := feedResp.Notifications[0]

		// Set the ID for subsequent operations
		notificationID = latestNotif.ID

		log.Info("Found notification in feed",
			slog.Int64("notification_id", latestNotif.ID),
			slog.String("type", latestNotif.Type))

		// Verify notification data
		if latestNotif.Type != fixtures.NotificationTypeSystem {
			t.Errorf("Expected notification type %s, got %s", fixtures.NotificationTypeSystem, latestNotif.Type)
		}

		if latestNotif.IsRead {
			t.Error("Expected notification to be unread")
		}
	}

	// Get unread count
	unreadResp, err := notifClient.GetUnreadCount(user.ID)
	if err != nil {
		t.Fatalf("Failed to get unread count: %v", err)
	}

	if unreadResp.Count < 1 {
		t.Error("Expected at least 1 unread notification")
	}

	readResp, err := notifClient.ReadNotification(notificationID)
	if err != nil {
		t.Errorf("Failed to mark notification as read: %v. Make sure token is valid and user owns the notification.", err)
	} else {
		if !readResp.Success {
			t.Errorf("Expected successful read notification response but got unsuccessful.")
		} else {
			// Verify notification is marked as read
			notif, err := notifClient.GetNotificationByID(notificationID)
			if err != nil {
				t.Errorf("Failed to get notification by ID: %v", err)
			} else if !notif.IsRead {
				t.Errorf("Notification was not marked as read. API returned IsRead=false.")
			} else {
				log.Info("Successfully marked notification as read", "notification_id", notificationID)
			}
		}
	}

	// Get the final unread notification count
	finalUnreadResp, err := notifClient.GetUnreadCount(user.ID)
	if err != nil {
		t.Errorf("Failed to get final unread count: %v", err)
	} else if finalUnreadResp.Count >= unreadResp.Count {
		// We expect the unread count to decrease after marking notifications as read
		t.Errorf("Unexpected unread count after marking notifications as read. Expected < %d, got %d",
			unreadResp.Count, finalUnreadResp.Count)
	}

	log.Info("Completed notifications test",
		"user_id", user.ID,
		"notification_id", notificationID,
		"initial_unread_count", unreadResp.Count,
		"final_unread_count", finalUnreadResp.Count)
}
