package gateway_notification

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSendNotificationTest(t *testing.T, tc *TestContext) (senderToken string, senderId int64, recipientID int64, recipientToken string, teardown func()) {
	t.Helper()

	// Register sender user
	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up send notification test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	// Register recipient user
	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up send notification test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

	recipientTokens, err := tc.AuthClient.Register(*recipientRegisterReq)
	require.NoError(t, err, "Failed to register recipient user")

	recipientUser, err := tc.UserClient.GetUserByUsername(recipientRegisterReq.Username)
	require.NoError(t, err, "Failed to get recipient user info")

	tc.TrackUserForCleanup(recipientUser.ID, recipientUser.Username, recipientTokens.AccessToken)

	return senderTokens.AccessToken, senderUser.ID, recipientUser.ID, recipientTokens.AccessToken, func() {
		log.Info("Send notification test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestSendNotificationSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup test users (sender and recipient)
	senderAccessToken, _, recipientID, recipientAccessToken, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	// Set sender's token for authorization
	tc.APIClient.SetToken(senderAccessToken)

	// Create notification request with different types
	notificationTypes := fixtures.NotificationTypes

	for _, notificationType := range notificationTypes {
		t.Run(notificationType, func(t *testing.T) {
			// Generate notification request using fixtures generator
			notificationReqPtr := fixtures.GenerateSendNotificationRequest(recipientID) // Отправляем уведомление получателю
			notificationReq := *notificationReqPtr                                      // Разыменование указателя для получения значения
			notificationReq.Type = notificationType                                     // Override with specific type for this test

			// Send notification
			response, err := tc.NotificationClient.SendNotification(notificationReq)
			require.NoError(t, err, "Failed to send notification")
			assert.NotNil(t, response, "Response should not be nil")
			assert.NotEmpty(t, response.Message, "Message should not be empty")
			assert.Greater(t, response.NotificationID, int64(0), "Notification ID should be positive")

			// Track notification for cleanup
			tc.TrackNotificationForCleanup(response.NotificationID, recipientID, senderAccessToken, recipientAccessToken)

			// Verify notification was created by getting it
			// Важно: уведомления может читать только получатель, поэтому нужно установить токен получателя
			// Но для упрощения тестирования мы просто проверим успешность создания уведомления
			log.Info("Successfully sent notification", "notification_id", response.NotificationID, "type", notificationType)
		})
	}
}

func TestSendNotificationUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup test users
	_, _, recipientID, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	// Clear token to simulate unauthorized request
	tc.APIClient.SetToken("")

	// Generate and customize notification request
	notificationReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeSystem
	notificationReq.Payload = map[string]string{
		"message": "Unauthorized notification",
	}

	// Attempt should fail with authentication error
	_, err := tc.NotificationClient.SendNotification(notificationReq)
	require.Error(t, err, "Unauthorized request should fail")
	assert.Contains(t, err.Error(), "unauth", "Error should mention authorization issue")
}

func TestSendNotificationInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup test users
	_, _, recipientID, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	// Set invalid token
	tc.APIClient.SetToken("invalid_token_12345")

	// Generate notification request
	notificationReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeSystem

	// Attempt should fail with invalid token error
	_, err := tc.NotificationClient.SendNotification(notificationReq)
	require.Error(t, err, "Invalid token request should fail")
	assert.Contains(t, err.Error(), "invalid token", "Error should mention authorization issue")
}

func TestSendNotificationToNonExistentUser(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup test user (sender)
	senderAccessToken, _, _, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderAccessToken)

	// Generate and customize notification request with non-existent user
	notificationReqPtr := fixtures.GenerateSendNotificationRequest(999999) // Non-existent user ID
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeSystem

	// Attempt should fail with user not found error
	_, err := tc.NotificationClient.SendNotification(notificationReq)
	require.Error(t, err, "Sending to non-existent user should fail")
	assert.Contains(t, err.Error(), "not found", "Error should mention user not found")
}

func TestSendNotificationValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup test users
	senderAccessToken, _, recipientID, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderAccessToken)

	// Prepare base request first using generator
	baseNotifReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
	baseNotifReq := *baseNotifReqPtr

	// Define test cases with specific modifications
	testCases := []struct {
		name          string
		setupNotifReq func() fixtures.SendNotificationRequest
		expectedErr   string
	}{
		{
			name: "EmptyType",
			setupNotifReq: func() fixtures.SendNotificationRequest {
				req := baseNotifReq
				req.Type = ""
				return req
			},
			expectedErr: "validation failed",
		},
		{
			name: "ZeroUserID",
			setupNotifReq: func() fixtures.SendNotificationRequest {
				req := baseNotifReq
				req.UserID = 0
				return req
			},
			expectedErr: "validation failed",
		},
		{
			name: "NegativeUserID",
			setupNotifReq: func() fixtures.SendNotificationRequest {
				req := baseNotifReq
				req.UserID = -1
				return req
			},
			expectedErr: "validation failed",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := NewTestContext()
			defer ctx.Cleanup()

			// Get a new token for each subtest to avoid conflicts
			accessToken, _, _, _, teardown := setupSendNotificationTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			// Получаем настроенный запрос из setupNotifReq функции
			notifReq := tc.setupNotifReq()
			_, err := ctx.NotificationClient.SendNotification(notifReq)
			require.Error(t, err, "Invalid request should fail")
			assert.Contains(t, err.Error(), tc.expectedErr, "Error should mention validation failure")
		})
	}
}

func TestSendNotificationSelfNotification(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	// Setup single user
	accessToken, userID, teardown := setupTestUser(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	// Try to send notification to self
	notificationReqPtr := fixtures.GenerateSendNotificationRequest(userID)
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeFollowCreated

	// This might be allowed or not depending on business logic
	// Adjust the test based on your application's behavior
	response, err := tc.NotificationClient.SendNotification(notificationReq)
	if err != nil {
		// If self-notifications are not allowed
		assert.Contains(t, err.Error(), "forbidden", "Self-notification should be forbidden")
	} else {
		// If self-notifications are allowed, track for cleanup
		tc.TrackNotificationForCleanup(response.NotificationID, userID, accessToken, accessToken)
		assert.NotNil(t, response, "Self-notification should be allowed")
	}
}
