package gateway_notification

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSendNotificationTest(t *testing.T, tc *TestContext) (senderToken string, senderId int64, recipientID int64, recipientToken string, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up send notification test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

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

	senderAccessToken, _, recipientID, recipientAccessToken, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderAccessToken)

	notificationTypes := fixtures.NotificationTypes

	for _, notificationType := range notificationTypes {
		t.Run(notificationType, func(t *testing.T) {
			notificationReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
			notificationReq := *notificationReqPtr
			notificationReq.Type = notificationType // Override with specific type for this test

			response, err := tc.NotificationClient.SendNotification(notificationReq)
			require.NoError(t, err, "Failed to send notification")
			assert.NotNil(t, response, "Response should not be nil")
			assert.NotEmpty(t, response.Message, "Message should not be empty")
			assert.Greater(t, response.NotificationID, int64(0), "Notification ID should be positive")

			tc.TrackNotificationForCleanup(response.NotificationID, recipientID, senderAccessToken, recipientAccessToken)

			log.Info("Successfully sent notification", "notification_id", response.NotificationID, "type", notificationType)
		})
	}
}

func TestSendNotificationUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, recipientID, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	notificationReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeSystem
	notificationReq.Payload = map[string]string{
		"message": "Unauthorized notification",
	}

	_, err := tc.NotificationClient.SendNotification(notificationReq)
	require.Error(t, err, "Unauthorized request should fail")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
}

func TestSendNotificationInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, recipientID, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("invalid_token_12345")

	notificationReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeSystem

	_, err := tc.NotificationClient.SendNotification(notificationReq)
	require.Error(t, err, "Invalid token request should fail")
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error(), "Error should be invalid token")
}

func TestSendNotificationToNonExistentUser(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderAccessToken, _, _, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderAccessToken)

	notificationReqPtr := fixtures.GenerateSendNotificationRequest(999999) // Non-existent user ID
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeSystem

	_, err := tc.NotificationClient.SendNotification(notificationReq)
	require.Error(t, err, "Sending to non-existent user should fail")
	assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error(), "Error should be user not found")
}

func TestSendNotificationValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderAccessToken, _, recipientID, _, teardown := setupSendNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderAccessToken)

	baseNotifReqPtr := fixtures.GenerateSendNotificationRequest(recipientID)
	baseNotifReq := *baseNotifReqPtr

	testCases := []struct {
		name          string
		setupNotifReq func() fixtures.SendNotificationRequest
		expectedErr   error
	}{
		{
			name: "EmptyType",
			setupNotifReq: func() fixtures.SendNotificationRequest {
				req := baseNotifReq
				req.Type = ""
				return req
			},
			expectedErr: custom_errors.ErrValidationFailed,
		},
		{
			name: "ZeroUserID",
			setupNotifReq: func() fixtures.SendNotificationRequest {
				req := baseNotifReq
				req.UserID = 0
				return req
			},
			expectedErr: custom_errors.ErrValidationFailed,
		},
		{
			name: "NegativeUserID",
			setupNotifReq: func() fixtures.SendNotificationRequest {
				req := baseNotifReq
				req.UserID = -1
				return req
			},
			expectedErr: custom_errors.ErrValidationFailed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, _, _, teardown := setupSendNotificationTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			notifReq := tc.setupNotifReq()
			_, err := ctx.NotificationClient.SendNotification(notifReq)
			require.Error(t, err, "Invalid request should fail")
			assert.Contains(t, err.Error(), tc.expectedErr.Error(), "Error should match expected validation error")
		})
	}
}

func TestSendNotificationSelfNotification(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupTestUser(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	notificationReqPtr := fixtures.GenerateSendNotificationRequest(userID)
	notificationReq := *notificationReqPtr
	notificationReq.Type = fixtures.NotificationTypeFollowCreated

	response, err := tc.NotificationClient.SendNotification(notificationReq)
	if err != nil {
		assert.Contains(t, err.Error(), custom_errors.ErrForbidden.Error(), "Self-notification should be forbidden")
	} else {
		tc.TrackNotificationForCleanup(response.NotificationID, userID, accessToken, accessToken)
		assert.NotNil(t, response, "Self-notification should be allowed")
	}
}
