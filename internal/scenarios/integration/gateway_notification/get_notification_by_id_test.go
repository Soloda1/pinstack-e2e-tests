package gateway_notification

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetNotificationByIDTest(t *testing.T, tc *TestContext) (senderToken string, recipientID int64, recipientToken string, notificationID int64, teardown func()) {
	t.Helper()

	senderRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get notification by ID test - registering sender", "test", t.Name(), "username", senderRegisterReq.Username)

	senderTokens, err := tc.AuthClient.Register(*senderRegisterReq)
	require.NoError(t, err, "Failed to register sender user")

	senderUser, err := tc.UserClient.GetUserByUsername(senderRegisterReq.Username)
	require.NoError(t, err, "Failed to get sender user info")

	tc.TrackUserForCleanup(senderUser.ID, senderUser.Username, senderTokens.AccessToken)

	recipientRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get notification by ID test - registering recipient", "test", t.Name(), "username", recipientRegisterReq.Username)

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
		log.Info("Get notification by ID test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetNotificationByIDSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, recipientID, recipientAccessToken, notificationID, teardown := setupGetNotificationByIDTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(recipientAccessToken)

	notification, err := tc.NotificationClient.GetNotificationByID(notificationID)

	require.NoError(t, err, "Failed to get notification by ID")
	require.NotNil(t, notification, "Notification should not be nil")

	assert.Equal(t, notificationID, notification.ID, "Notification ID should match")
	assert.Equal(t, recipientID, notification.UserID, "User ID should match recipient")
	assert.Contains(t, fixtures.NotificationTypes, notification.Type, "Type should be valid notification type")
	assert.NotNil(t, notification.Payload, "Payload should not be nil")
	assert.False(t, notification.IsRead, "Notification should be unread by default")
	assert.NotZero(t, notification.CreatedAt, "Created at should not be zero")

	log.Info("Successfully retrieved notification by ID", "notification_id", notificationID, "type", notification.Type)
}

func TestGetNotificationByIDUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, notificationID, teardown := setupGetNotificationByIDTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	notification, err := tc.NotificationClient.GetNotificationByID(notificationID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, notification, "Notification should be nil on error")

	log.Info("Correctly rejected unauthorized request for notification", "notification_id", notificationID)
}

func TestGetNotificationByIDForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	senderToken, _, _, notificationID, teardown := setupGetNotificationByIDTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(senderToken)

	notification, err := tc.NotificationClient.GetNotificationByID(notificationID)

	require.Error(t, err, "Should fail when accessing other user's notification")
	assert.Contains(t, err.Error(), custom_errors.ErrNotificationAccessDenied.Error(), "Error should be access denied")
	assert.Nil(t, notification, "Notification should be nil on error")

	log.Info("Correctly rejected forbidden request for notification", "notification_id", notificationID)
}

func TestGetNotificationByIDInvalidID(t *testing.T) {
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
			notification, err := tc.NotificationClient.GetNotificationByID(testCase.notificationID)

			require.Error(t, err, "Should fail for %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedError.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, notification, "Notification should be nil on error")

			log.Info("Correctly rejected request for invalid notification ID",
				"notification_id", testCase.notificationID,
				"description", testCase.description,
				"expected_error", testCase.expectedError.Error())
		})
	}
}
