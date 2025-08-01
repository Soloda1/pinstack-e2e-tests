package gateway_notification

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReadNotificationTest(t *testing.T, tc *TestContext) (senderToken string, recipientID int64, recipientToken string, notificationID int64, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up read notification test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up read notification test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

	recipientTokens, err := tc.AuthClient.Register(*recipientRegisterReq)
	require.NoError(t, err, "Failed to register recipient user")

	recipientUser, err := tc.UserClient.GetUserByUsername(recipientRegisterReq.Username)
	require.NoError(t, err, "Failed to get recipient user info")

	tc.TrackUserForCleanup(recipientUser.ID, recipientUser.Username, recipientTokens.AccessToken)

	tc.APIClient.SetToken(senderTokens.AccessToken)
	sendReq := fixtures.GenerateSendNotificationRequest(recipientUser.ID)

	sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
	require.NoError(t, err, "Failed to send notification for test setup")
	require.NotEmpty(t, sendResp.NotificationID, "Notification ID should not be empty")

	tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientUser.ID, senderTokens.AccessToken, recipientTokens.AccessToken)

	return senderTokens.AccessToken, recipientUser.ID, recipientTokens.AccessToken, sendResp.NotificationID, func() {
		log.Info("Read notification test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestReadNotificationSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, recipientToken, notificationID, teardown := setupReadNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)

	notification, err := tc.NotificationClient.GetNotificationByID(notificationID)
	require.NoError(t, err, "Failed to get notification before marking as read")
	assert.False(t, notification.IsRead, "Notification should be unread initially")

	readResp, err := tc.NotificationClient.ReadNotification(notificationID)

	require.NoError(t, err, "Failed to mark notification as read")
	require.NotNil(t, readResp, "Response should not be nil")

	assert.True(t, readResp.Success, "Operation should be successful")
	assert.NotEmpty(t, readResp.Message, "Response should have a message")

	updatedNotification, err := tc.NotificationClient.GetNotificationByID(notificationID)
	require.NoError(t, err, "Failed to get notification after marking as read")
	assert.True(t, updatedNotification.IsRead, "Notification should be marked as read")

	log.Info("Successfully marked notification as read", "notification_id", notificationID)
}

func TestReadNotificationAlreadyRead(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, recipientToken, notificationID, teardown := setupReadNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)

	readResp1, err := tc.NotificationClient.ReadNotification(notificationID)
	require.NoError(t, err, "Failed to mark notification as read first time")
	assert.True(t, readResp1.Success, "First operation should be successful")

	readResp2, err := tc.NotificationClient.ReadNotification(notificationID)
	require.NoError(t, err, "Should succeed even when notification is already read")
	assert.True(t, readResp2.Success, "Second operation should be successful")

	notification, err := tc.NotificationClient.GetNotificationByID(notificationID)
	require.NoError(t, err, "Failed to get notification after double read")
	assert.True(t, notification.IsRead, "Notification should remain marked as read")

	log.Info("Successfully handled double read of notification", "notification_id", notificationID)
}

func TestReadNotificationUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, notificationID, teardown := setupReadNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	readResp, err := tc.NotificationClient.ReadNotification(notificationID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, readResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized request for read notification", "notification_id", notificationID)
}

func TestReadNotificationForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, _, _, notificationID, teardown := setupReadNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)

	readResp, err := tc.NotificationClient.ReadNotification(notificationID)

	require.Error(t, err, "Should fail when accessing other user's notification")
	assert.Contains(t, err.Error(), custom_errors.ErrNotificationAccessDenied.Error(), "Error should be access denied")
	assert.Nil(t, readResp, "Response should be nil on error")

	log.Info("Correctly rejected forbidden request for read notification", "notification_id", notificationID)
}

func TestReadNotificationInvalidID(t *testing.T) {
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
		name           string
		notificationID int64
		description    string
		expectedError  error
	}{
		{
			name:           "zero_id",
			notificationID: 0,
			description:    "zero notification ID",
			expectedError:  custom_errors.ErrValidationFailed,
		},
		{
			name:           "negative_id",
			notificationID: -1,
			description:    "negative notification ID",
			expectedError:  custom_errors.ErrValidationFailed,
		},
		{
			name:           "non_existent_id",
			notificationID: 999999,
			description:    "non-existent notification ID",
			expectedError:  custom_errors.ErrNotificationNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			readResp, err := tc.NotificationClient.ReadNotification(testCase.notificationID)

			require.Error(t, err, "Should fail for %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedError.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, readResp, "Response should be nil on error")

			log.Info("Correctly rejected request for invalid notification ID",
				"notification_id", testCase.notificationID,
				"description", testCase.description,
				"expected_error", testCase.expectedError.Error())
		})
	}
}

func TestReadNotificationInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, notificationID, teardown := setupReadNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("invalid_token_12345")

	readResp, err := tc.NotificationClient.ReadNotification(notificationID)

	require.Error(t, err, "Should fail with invalid token")
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error(), "Error should be invalid token")
	assert.Nil(t, readResp, "Response should be nil on error")

	log.Info("Correctly rejected invalid token request for read notification", "notification_id", notificationID)
}

func TestReadNotificationMultipleNotifications(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, _, teardown := setupReadNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)
	notificationsToSend := 3
	var notificationIDs []int64

	for i := 0; i < notificationsToSend; i++ {
		sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
		sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
		require.NoError(t, err, "Failed to send notification %d", i+1)

		notificationIDs = append(notificationIDs, sendResp.NotificationID)
		tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)
	}

	tc.APIClient.SetToken(recipientToken)

	for i, notificationID := range notificationIDs {
		readResp, err := tc.NotificationClient.ReadNotification(notificationID)
		require.NoError(t, err, "Failed to mark notification %d as read", i+1)
		assert.True(t, readResp.Success, "Operation should be successful for notification %d", i+1)

		notification, err := tc.NotificationClient.GetNotificationByID(notificationID)
		require.NoError(t, err, "Failed to get notification %d after marking as read", i+1)
		assert.True(t, notification.IsRead, "Notification %d should be marked as read", i+1)
	}

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)
	require.NoError(t, err, "Failed to get notification feed")

	readCount := 0
	for _, notification := range feedResp.Notifications {
		if notification.IsRead {
			readCount++
		}
	}

	assert.GreaterOrEqual(t, readCount, notificationsToSend, "At least sent notifications should be marked as read")

	log.Info("Successfully marked multiple notifications as read individually",
		"recipient_id", recipientID,
		"notifications_count", notificationsToSend,
		"read_count", readCount)
}
