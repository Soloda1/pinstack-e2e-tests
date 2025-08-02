package gateway_relation

import (
	"log/slog"
	"testing"
	"time"

	"github.com/Soloda1/pinstack-system-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFollowUserTest(t *testing.T, tc *TestContext) (followerToken string, followerID int64, followeeToken string, followeeID int64, teardown func()) {
	t.Helper()

	followerRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up follow user test - registering follower", "test", t.Name(), "username", followerRegisterReq.Username)

	followerTokens, err := tc.AuthClient.Register(*followerRegisterReq)
	require.NoError(t, err, "Failed to register follower user")

	followerUser, err := tc.UserClient.GetUserByUsername(followerRegisterReq.Username)
	require.NoError(t, err, "Failed to get follower user info")

	tc.TrackUserForCleanup(followerUser.ID, followerUser.Username, followerTokens.AccessToken)

	followeeRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up follow user test - registering followee", "test", t.Name(), "username", followeeRegisterReq.Username)

	followeeTokens, err := tc.AuthClient.Register(*followeeRegisterReq)
	require.NoError(t, err, "Failed to register followee user")

	followeeUser, err := tc.UserClient.GetUserByUsername(followeeRegisterReq.Username)
	require.NoError(t, err, "Failed to get followee user info")

	tc.TrackUserForCleanup(followeeUser.ID, followeeUser.Username, followeeTokens.AccessToken)

	return followerTokens.AccessToken, followerUser.ID, followeeTokens.AccessToken, followeeUser.ID, func() {
		log.Info("Follow user test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestFollowUserSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, followeeToken, followeeID, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	followResp, err := tc.RelationClient.Follow(followeeID)

	require.NoError(t, err, "Failed to follow user")
	require.NotNil(t, followResp, "Response should not be nil")
	assert.NotEmpty(t, followResp.Message, "Response should have a message")

	tc.TrackRelationForCleanup(followerID, followeeID, followerToken)
	tc.DiscoverAndTrackAllNotifications(followeeID, followeeToken)

	log.Info("Successfully followed user", "follower_id", followerID, "followee_id", followeeID)
}

func TestFollowUserUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, followeeID, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	followResp, err := tc.RelationClient.Follow(followeeID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, followResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized follow request", "followee_id", followeeID)
}

func TestFollowUserInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, followeeID, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("invalid_token_12345")

	followResp, err := tc.RelationClient.Follow(followeeID)

	require.Error(t, err, "Should fail with invalid token")
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error(), "Error should be unauthenticated")
	assert.Nil(t, followResp, "Response should be nil on error")

	log.Info("Correctly rejected invalid token follow request", "followee_id", followeeID)
}

func TestFollowUserNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, _, _, _, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	nonExistentUserID := int64(999999)
	followResp, err := tc.RelationClient.Follow(nonExistentUserID)

	require.Error(t, err, "Should fail when following non-existent user")
	assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error(), "Error should be user not found")
	assert.Nil(t, followResp, "Response should be nil on error")

	log.Info("Correctly rejected follow request for non-existent user", "followee_id", nonExistentUserID)
}

func TestFollowUserSelf(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, _, _, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	followResp, err := tc.RelationClient.Follow(followerID)

	require.Error(t, err, "Should fail when trying to follow self")
	assert.Contains(t, err.Error(), custom_errors.ErrSelfFollow.Error(), "Error should be self follow")
	assert.Nil(t, followResp, "Response should be nil on error")

	log.Info("Correctly rejected self-follow request", "user_id", followerID)
}

func TestFollowUserAlreadyFollowing(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, followeeToken, followeeID, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	followResp1, err := tc.RelationClient.Follow(followeeID)
	require.NoError(t, err, "First follow should succeed")
	assert.NotNil(t, followResp1, "First response should not be nil")

	tc.TrackRelationForCleanup(followerID, followeeID, followerToken)
	tc.DiscoverAndTrackAllNotifications(followeeID, followeeToken)

	tc.APIClient.SetToken(followerToken)
	followResp2, err := tc.RelationClient.Follow(followeeID)

	require.Error(t, err, "Should fail when already following user")
	assert.Contains(t, err.Error(), custom_errors.ErrAlreadyFollowing.Error(), "Error should be already following")
	assert.Nil(t, followResp2, "Second response should be nil on error")

	log.Info("Correctly rejected duplicate follow request", "follower_id", followerID, "followee_id", followeeID)
}

func TestFollowUserValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, _, _, _, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	testCases := []struct {
		name        string
		followeeID  int64
		description string
		expectedErr error
	}{
		{
			name:        "zero_followee_id",
			followeeID:  0,
			description: "zero followee ID",
			expectedErr: custom_errors.ErrValidationFailed,
		},
		{
			name:        "negative_followee_id",
			followeeID:  -1,
			description: "negative followee ID",
			expectedErr: custom_errors.ErrValidationFailed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			followResp, err := tc.RelationClient.Follow(testCase.followeeID)

			require.Error(t, err, "Should fail with %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedErr.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, followResp, "Response should be nil on error for %s", testCase.description)

			log.Info("Correctly rejected invalid followee ID", "test_case", testCase.name, "followee_id", testCase.followeeID)
		})
	}
}

func TestFollowUserWithNotificationGeneration(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, followeeToken, followeeID, teardown := setupFollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	followResp, err := tc.RelationClient.Follow(followeeID)
	require.NoError(t, err, "Failed to follow user")
	assert.NotNil(t, followResp, "Response should not be nil")

	tc.TrackRelationForCleanup(followerID, followeeID, followerToken)

	tc.DiscoverAndTrackAllNotifications(followeeID, followeeToken)

	// Wait kafka outbox worker
	time.Sleep(1100 * time.Millisecond)
	// Verify that followee received notification
	tc.APIClient.SetToken(followeeToken)
	feedResp, err := tc.NotificationClient.GetUserNotificationFeed(followeeID, 1, 10)
	require.NoError(t, err, "Failed to get notification feed")

	var foundFollowNotification bool
	for _, notification := range feedResp.Notifications {
		log.Debug("notifications", slog.Any("notification", notification))
		if notification.Type == "follow_created" {
			foundFollowNotification = true
			log.Info("Found follow_created notification", "notification_id", notification.ID, "followee_id", followeeID)
			break
		}
	}

	assert.True(t, foundFollowNotification, "Should have created follow_created notification")

	log.Info("Successfully verified follow user with notification generation",
		"follower_id", followerID,
		"followee_id", followeeID,
		"notification_found", foundFollowNotification)
}
