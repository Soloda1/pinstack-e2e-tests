package gateway_notification

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRemoveNotificationTest(t *testing.T, tc *TestContext) (senderToken string, recipientID int64, recipientToken string, notificationID int64, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up remove notification test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up remove notification test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

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
		log.Info("Remove notification test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestRemoveNotificationSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, recipientToken, notificationID, teardown := setupRemoveNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)

	notification, err := tc.NotificationClient.GetNotificationByID(notificationID)
	require.NoError(t, err, "Failed to get notification before removal")
	require.NotNil(t, notification, "Notification should exist before removal")

	removeResp, err := tc.NotificationClient.RemoveNotification(notificationID)

	require.NoError(t, err, "Failed to remove notification")
	require.NotNil(t, removeResp, "Response should not be nil")

	assert.True(t, removeResp.Success, "Operation should be successful")
	assert.NotEmpty(t, removeResp.Message, "Response should have a message")

	_, err = tc.NotificationClient.GetNotificationByID(notificationID)
	require.Error(t, err, "Should fail to get removed notification")
	assert.Contains(t, err.Error(), custom_errors.ErrNotificationNotFound.Error(), "Error should be notification not found")

	log.Info("Successfully removed notification", "notification_id", notificationID)
}

func TestRemoveNotificationUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, notificationID, teardown := setupRemoveNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	removeResp, err := tc.NotificationClient.RemoveNotification(notificationID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, removeResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized request for remove notification", "notification_id", notificationID)
}

func TestRemoveNotificationForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, _, _, notificationID, teardown := setupRemoveNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)

	removeResp, err := tc.NotificationClient.RemoveNotification(notificationID)

	require.Error(t, err, "Should fail when accessing other user's notification")
	assert.Contains(t, err.Error(), custom_errors.ErrNotificationAccessDenied.Error(), "Error should be access denied")
	assert.Nil(t, removeResp, "Response should be nil on error")

	log.Info("Correctly rejected forbidden request for remove notification", "notification_id", notificationID)
}

func TestRemoveNotificationAlreadyRemoved(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, recipientToken, notificationID, teardown := setupRemoveNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientToken)

	removeResp1, err := tc.NotificationClient.RemoveNotification(notificationID)
	require.NoError(t, err, "Failed to remove notification first time")
	assert.True(t, removeResp1.Success, "First operation should be successful")

	removeResp2, err := tc.NotificationClient.RemoveNotification(notificationID)

	require.Error(t, err, "Should fail when removing already removed notification")
	assert.Contains(t, err.Error(), custom_errors.ErrNotificationNotFound.Error(), "Error should be notification not found")
	assert.Nil(t, removeResp2, "Response should be nil on error")

	log.Info("Correctly handled attempt to remove already removed notification", "notification_id", notificationID)
}

func TestRemoveNotificationInvalidID(t *testing.T) {
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
			// Try to remove notification with invalid ID
			removeResp, err := tc.NotificationClient.RemoveNotification(testCase.notificationID)

			// Verify error
			require.Error(t, err, "Should fail with %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedError.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, removeResp, "Response should be nil on error for %s", testCase.description)

			log.Info("Correctly rejected invalid ID", "test_case", testCase.name, "notification_id", testCase.notificationID)
		})
	}
}

func TestRemoveNotificationInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, notificationID, teardown := setupRemoveNotificationTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("invalid-token-12345")

	removeResp, err := tc.NotificationClient.RemoveNotification(notificationID)

	require.Error(t, err, "Should fail with invalid token")
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error(), "Error should be unauthenticated")
	assert.Nil(t, removeResp, "Response should be nil on error")

	log.Info("Correctly rejected invalid token for remove notification", "notification_id", notificationID)
}
