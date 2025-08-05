package gateway_notification

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReadAllNotificationsTest(t *testing.T, tc *TestContext) (senderToken string, recipientID int64, recipientToken string, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up read all notifications test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up read all notifications test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

	recipientTokens, err := tc.AuthClient.Register(*recipientRegisterReq)
	require.NoError(t, err, "Failed to register recipient user")

	recipientUser, err := tc.UserClient.GetUserByUsername(recipientRegisterReq.Username)
	require.NoError(t, err, "Failed to get recipient user info")

	tc.TrackUserForCleanup(recipientUser.ID, recipientUser.Username, recipientTokens.AccessToken)

	return senderTokens.AccessToken, recipientUser.ID, recipientTokens.AccessToken, func() {
		log.Info("Read all notifications test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestReadAllNotificationsSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupReadAllNotificationsTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)
	notificationsToSend := 3

	for i := 0; i < notificationsToSend; i++ {
		sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
		sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
		require.NoError(t, err, "Failed to send notification %d", i+1)

		tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)
	}

	tc.APIClient.SetToken(recipientToken)

	readAllResp, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)

	require.NoError(t, err, "Failed to mark all notifications as read")
	require.NotNil(t, readAllResp, "Response should not be nil")

	assert.True(t, readAllResp.Success, "Operation should be successful")
	assert.NotEmpty(t, readAllResp.Message, "Response should have a message")

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)
	require.NoError(t, err, "Failed to get notification feed")

	for _, notification := range feedResp.Notifications {
		assert.True(t, notification.IsRead, "All notifications should be marked as read")
	}

	log.Info("Successfully marked all notifications as read",
		"recipient_id", recipientID,
		"notifications_count", notificationsToSend)
}

func TestReadAllNotificationsNoNotifications(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, recipientToken, teardown := setupReadAllNotificationsTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)

	readAllResp, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)

	require.NoError(t, err, "Should succeed even with no notifications")
	require.NotNil(t, readAllResp, "Response should not be nil")

	assert.True(t, readAllResp.Success, "Operation should be successful")
	assert.NotEmpty(t, readAllResp.Message, "Response should have a message")

	log.Info("Successfully handled read all notifications with no notifications", "recipient_id", recipientID)
}

func TestReadAllNotificationsUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, _, teardown := setupReadAllNotificationsTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	readAllResp, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, readAllResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized request for read all notifications", "recipient_id", recipientID)
}

func TestReadAllNotificationsInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, _, teardown := setupReadAllNotificationsTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("invalid_token_12345")

	readAllResp, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)

	require.Error(t, err, "Should fail with invalid token")
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error(), "Error should be invalid token")
	assert.Nil(t, readAllResp, "Response should be nil on error")

	log.Info("Correctly rejected invalid token request for read all notifications", "recipient_id", recipientID)
}

func TestReadAllNotificationsAlreadyRead(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupReadAllNotificationsTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)
	sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
	sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
	require.NoError(t, err, "Failed to send notification")

	tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)

	tc.APIClient.SetToken(recipientToken)

	readAllResp1, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)
	require.NoError(t, err, "Failed to mark all notifications as read first time")
	assert.True(t, readAllResp1.Success, "First operation should be successful")

	readAllResp2, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)
	require.NoError(t, err, "Should succeed even when notifications are already read")
	assert.True(t, readAllResp2.Success, "Second operation should be successful")

	log.Info("Successfully handled read all notifications when already read", "recipient_id", recipientID)
}

func TestReadAllNotificationsWithMixedReadStatus(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupReadAllNotificationsTest(t, tc)
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

	for i := 0; i < 2; i++ {
		readResp, err := tc.NotificationClient.ReadNotification(sentNotificationIDs[i])
		require.NoError(t, err, "Failed to mark notification %d as read", i+1)
		assert.True(t, readResp.Success, "Individual read operation should be successful")
	}

	readAllResp, err := tc.NotificationClient.ReadAllUserNotifications(recipientID)
	require.NoError(t, err, "Failed to mark all notifications as read")
	assert.True(t, readAllResp.Success, "Read all operation should be successful")

	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(recipientID, 1, 10)
	require.NoError(t, err, "Failed to get notification feed")

	for _, notification := range feedResp.Notifications {
		assert.True(t, notification.IsRead, "All notifications should be marked as read")
	}

	log.Info("Successfully marked all notifications as read with mixed initial status",
		"recipient_id", recipientID,
		"notifications_count", notificationsToSend)
}
