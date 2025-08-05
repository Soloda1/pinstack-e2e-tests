package gateway_notification

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetUserNotificationFeedTest(t *testing.T, tc *TestContext) (senderToken string, recipientID int64, recipientToken string, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get user notification feed test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get user notification feed test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

	recipientTokens, err := tc.AuthClient.Register(*recipientRegisterReq)
	require.NoError(t, err, "Failed to register recipient user")

	recipientUser, err := tc.UserClient.GetUserByUsername(recipientRegisterReq.Username)
	require.NoError(t, err, "Failed to get recipient user info")

	tc.TrackUserForCleanup(recipientUser.ID, recipientUser.Username, recipientTokens.AccessToken)

	return senderTokens.AccessToken, recipientUser.ID, recipientTokens.AccessToken, func() {
		log.Info("Get user notification feed test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetUserNotificationFeedEmptyFeed(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, recipientToken, teardown := setupGetUserNotificationFeedTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)

	require.NoError(t, err, "Failed to get empty notification feed")
	require.NotNil(t, feedResp, "Feed response should not be nil")

	assert.Equal(t, 0, len(feedResp.Notifications), "Notifications list should be empty")
	assert.Equal(t, 0, feedResp.Total, "Total count should be 0")
	assert.Equal(t, 1, feedResp.Page, "Page should be 1")
	assert.Equal(t, 10, feedResp.Limit, "Limit should be 10")
	assert.Equal(t, 0, feedResp.TotalPages, "Total pages should be 0")

	log.Info("Successfully retrieved empty notification feed", "recipient_id", recipientID)
}

func TestGetUserNotificationFeedWithNotifications(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupGetUserNotificationFeedTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)
	notificationsToSend := 5
	var sentNotificationIDs []int64

	for i := 0; i < notificationsToSend; i++ {
		sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
		sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
		require.NoError(t, err, "Failed to send notification %d", i+1)

		sentNotificationIDs = append(sentNotificationIDs, sendResp.NotificationID)
		tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)
	}

	tc.APIClient.SetToken(recipientToken)

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)

	require.NoError(t, err, "Failed to get notification feed")
	require.NotNil(t, feedResp, "Feed response should not be nil")

	assert.Equal(t, notificationsToSend, len(feedResp.Notifications), "Should return all sent notifications")
	assert.Equal(t, notificationsToSend, feedResp.Total, "Total count should match sent notifications")
	assert.Equal(t, 1, feedResp.Page, "Page should be 1")
	assert.Equal(t, 10, feedResp.Limit, "Limit should be 10")
	assert.Equal(t, 1, feedResp.TotalPages, "Total pages should be 1")

	for _, notification := range feedResp.Notifications {
		assert.Equal(t, recipientID, notification.UserID, "User ID should match recipient")
		assert.Contains(t, fixtures.NotificationTypes, notification.Type, "Type should be valid notification type")
		assert.NotNil(t, notification.Payload, "Payload should not be nil")
		assert.False(t, notification.IsRead, "Notifications should be unread by default")
		assert.NotZero(t, notification.CreatedAt, "Created at should not be zero")
		assert.Contains(t, sentNotificationIDs, notification.ID, "Notification ID should be in sent list")
	}

	log.Info("Successfully retrieved notification feed with notifications",
		"recipient_id", recipientID,
		"notifications_count", len(feedResp.Notifications))
}

func TestGetUserNotificationFeedPagination(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupGetUserNotificationFeedTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)
	notificationsToSend := 15

	for i := 0; i < notificationsToSend; i++ {
		sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
		sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
		require.NoError(t, err, "Failed to send notification %d", i+1)

		tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)
	}

	tc.APIClient.SetToken(recipientToken)

	feedResp1, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)
	require.NoError(t, err, "Failed to get first page of notification feed")
	require.NotNil(t, feedResp1, "First page response should not be nil")

	assert.Equal(t, 10, len(feedResp1.Notifications), "First page should return 10 notifications")
	assert.Equal(t, notificationsToSend, feedResp1.Total, "Total count should match sent notifications")
	assert.Equal(t, 1, feedResp1.Page, "Page should be 1")
	assert.Equal(t, 10, feedResp1.Limit, "Limit should be 10")
	assert.Equal(t, 2, feedResp1.TotalPages, "Total pages should be 2")

	feedResp2, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 2, 10)
	require.NoError(t, err, "Failed to get second page of notification feed")
	require.NotNil(t, feedResp2, "Second page response should not be nil")

	assert.Equal(t, 5, len(feedResp2.Notifications), "Second page should return 5 notifications")
	assert.Equal(t, notificationsToSend, feedResp2.Total, "Total count should match sent notifications")
	assert.Equal(t, 2, feedResp2.Page, "Page should be 2")
	assert.Equal(t, 10, feedResp2.Limit, "Limit should be 10")
	assert.Equal(t, 2, feedResp2.TotalPages, "Total pages should be 2")

	page1IDs := make(map[int64]bool)
	for _, notification := range feedResp1.Notifications {
		page1IDs[notification.ID] = true
	}

	for _, notification := range feedResp2.Notifications {
		assert.False(t, page1IDs[notification.ID], "Notification should not appear on both pages")
	}

	log.Info("Successfully tested pagination",
		"total_notifications", notificationsToSend,
		"page1_count", len(feedResp1.Notifications),
		"page2_count", len(feedResp2.Notifications))
}

func TestGetUserNotificationFeedUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, _, teardown := setupGetUserNotificationFeedTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, feedResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized request for notification feed", "recipient_id", recipientID)
}

func TestGetUserNotificationFeedInvalidPagination(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	registerReq := fixtures.GenerateRegisterRequest()
	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register user")

	user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info")

	tc.TrackUserForCleanup(user.ID, user.Username, tokens.AccessToken)

	tc.APIClient.SetToken(tokens.AccessToken)

	testCases := []struct {
		name        string
		page        int
		limit       int
		description string
	}{
		{
			name:        "zero_page",
			page:        0,
			limit:       10,
			description: "zero page number",
		},
		{
			name:        "negative_page",
			page:        -1,
			limit:       10,
			description: "negative page number",
		},
		{
			name:        "zero_limit",
			page:        1,
			limit:       0,
			description: "zero limit",
		},
		{
			name:        "negative_limit",
			page:        1,
			limit:       -5,
			description: "negative limit",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			feedResp, err := tc.NotificationClient.GetUserNotificationFeed(user.ID, testCase.page, testCase.limit)

			require.Error(t, err, "Should fail for %s", testCase.description)
			assert.Contains(t, err.Error(), custom_errors.ErrInvalidInput.Error(), "Error should be validation failed for %s", testCase.description)
			assert.Nil(t, feedResp, "Response should be nil on error")

			log.Info("Correctly rejected request with invalid pagination",
				"page", testCase.page,
				"limit", testCase.limit,
				"description", testCase.description)
		})
	}
}
