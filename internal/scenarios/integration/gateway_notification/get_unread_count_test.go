package gateway_notification

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetUnreadCountTest(t *testing.T, tc *TestContext) (senderToken string, recipientID int64, recipientToken string, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get unread count test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get unread count test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

	recipientTokens, err := tc.AuthClient.Register(*recipientRegisterReq)
	require.NoError(t, err, "Failed to register recipient user")

	recipientUser, err := tc.UserClient.GetUserByUsername(recipientRegisterReq.Username)
	require.NoError(t, err, "Failed to get recipient user info")

	tc.TrackUserForCleanup(recipientUser.ID, recipientUser.Username, recipientTokens.AccessToken)

	return senderTokens.AccessToken, recipientUser.ID, recipientTokens.AccessToken, func() {
		log.Info("Get unread count test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetUnreadCountSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupGetUnreadCountTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)
	countResp, err := tc.NotificationClient.GetUnreadCount(recipientID)
	require.NoError(t, err, "Failed to get initial unread count")
	assert.Equal(t, 0, countResp.Count, "Initial unread count should be 0")

	tc.APIClient.SetToken(senderToken)
	notificationsToSend := 3

	for i := 0; i < notificationsToSend; i++ {
		sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
		sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
		require.NoError(t, err, "Failed to send notification %d", i+1)

		tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)
	}

	tc.APIClient.SetToken(recipientToken)
	countResp, err = tc.NotificationClient.GetUnreadCount(recipientID)
	require.NoError(t, err, "Failed to get unread count after sending notifications")
	assert.Equal(t, notificationsToSend, countResp.Count, "Unread count should match sent notifications")

	log.Info("Successfully verified unread count", "expected", notificationsToSend, "actual", countResp.Count)
}

func TestGetUnreadCountAfterReadingNotifications(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, recipientID, recipientToken, teardown := setupGetUnreadCountTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)
	notificationsToSend := 3
	var sentNotificationIDs []int64

	for i := 0; i < notificationsToSend; i++ {
		sendReq := fixtures.GenerateSendNotificationRequest(recipientID)
		sendResp, err := tc.NotificationClient.SendNotification(*sendReq)
		require.NoError(t, err, "Failed to send notification %d", i+1)

		sentNotificationIDs = append(sentNotificationIDs, sendResp.NotificationID)
		tc.TrackNotificationForCleanup(sendResp.NotificationID, recipientID, senderToken, recipientToken)
	}

	tc.APIClient.SetToken(recipientToken)
	countResp, err := tc.NotificationClient.GetUnreadCount(recipientID)
	require.NoError(t, err, "Failed to get initial unread count")
	assert.Equal(t, notificationsToSend, countResp.Count, "Initial unread count should match sent notifications")

	_, err = tc.NotificationClient.ReadNotification(sentNotificationIDs[0])
	require.NoError(t, err, "Failed to mark notification as read")

	countResp, err = tc.NotificationClient.GetUnreadCount(recipientID)
	require.NoError(t, err, "Failed to get unread count after reading one notification")
	assert.Equal(t, notificationsToSend-1, countResp.Count, "Unread count should decrease by 1")

	_, err = tc.NotificationClient.ReadAllUserNotifications(recipientID)
	require.NoError(t, err, "Failed to mark all notifications as read")

	countResp, err = tc.NotificationClient.GetUnreadCount(recipientID)
	require.NoError(t, err, "Failed to get unread count after reading all notifications")
	assert.Equal(t, 0, countResp.Count, "Unread count should be 0 after reading all")

	log.Info("Successfully verified unread count changes after reading notifications")
}

func TestGetUnreadCountUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, _, teardown := setupGetUnreadCountTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	countResp, err := tc.NotificationClient.GetUnreadCount(recipientID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, countResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized request for unread count", "user_id", recipientID)
}

func TestGetUnreadCountOwnUserID(t *testing.T) {
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

	countResp, err := tc.NotificationClient.GetUnreadCount(user.ID)

	require.NoError(t, err, "User should be able to get their own unread count")
	require.NotNil(t, countResp, "Response should not be nil")
	assert.GreaterOrEqual(t, countResp.Count, 0, "Unread count should be non-negative")

	log.Info("Successfully retrieved own unread count", "user_id", user.ID, "count", countResp.Count)
}
