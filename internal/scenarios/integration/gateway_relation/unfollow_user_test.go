package gateway_relation

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUnfollowUserTest(t *testing.T, tc *TestContext) (followerToken string, followerID int64, followeeToken string, followeeID int64, teardown func()) {
	t.Helper()

	followerRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up unfollow user test - registering follower", "test", t.Name(), "username", followerRegisterReq.Username)

	followerTokens, err := tc.AuthClient.Register(*followerRegisterReq)
	require.NoError(t, err, "Failed to register follower user")

	followerUser, err := tc.UserClient.GetUserByUsername(followerRegisterReq.Username)
	require.NoError(t, err, "Failed to get follower user info")

	tc.TrackUserForCleanup(followerUser.ID, followerUser.Username, followerTokens.AccessToken)

	followeeRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up unfollow user test - registering followee", "test", t.Name(), "username", followeeRegisterReq.Username)

	followeeTokens, err := tc.AuthClient.Register(*followeeRegisterReq)
	require.NoError(t, err, "Failed to register followee user")

	followeeUser, err := tc.UserClient.GetUserByUsername(followeeRegisterReq.Username)
	require.NoError(t, err, "Failed to get followee user info")

	tc.TrackUserForCleanup(followeeUser.ID, followeeUser.Username, followeeTokens.AccessToken)

	return followerTokens.AccessToken, followerUser.ID, followeeTokens.AccessToken, followeeUser.ID, func() {
		log.Info("Unfollow user test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func setupUnfollowUserTestWithExistingRelation(t *testing.T, tc *TestContext) (followerToken string, followerID int64, followeeToken string, followeeID int64, teardown func()) {
	t.Helper()

	followerToken, followerID, followeeToken, followeeID, baseTeardown := setupUnfollowUserTest(t, tc)

	tc.APIClient.SetToken(followerToken)
	followResp, err := tc.RelationClient.Follow(followeeID)
	require.NoError(t, err, "Failed to create follow relation for unfollow test")
	require.NotNil(t, followResp, "Follow response should not be nil")

	tc.TrackRelationForCleanup(followerID, followeeID, followerToken)

	tc.DiscoverAndTrackAllNotifications(followeeID, followeeToken)

	return followerToken, followerID, followeeToken, followeeID, func() {
		baseTeardown()
	}
}

func TestUnfollowUserSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, _, followeeID, teardown := setupUnfollowUserTestWithExistingRelation(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	unfollowResp, err := tc.RelationClient.Unfollow(followeeID)

	require.NoError(t, err, "Failed to unfollow user")
	require.NotNil(t, unfollowResp, "Response should not be nil")
	assert.NotEmpty(t, unfollowResp.Message, "Response should have a message")

	log.Info("Successfully unfollowed user", "follower_id", followerID, "followee_id", followeeID)
}

func TestUnfollowUserUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, followeeID, teardown := setupUnfollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("")

	unfollowResp, err := tc.RelationClient.Unfollow(followeeID)

	require.Error(t, err, "Should fail without authentication")
	assert.Contains(t, err.Error(), custom_errors.ErrUnauthenticated.Error(), "Error should be unauthenticated")
	assert.Nil(t, unfollowResp, "Response should be nil on error")

	log.Info("Correctly rejected unauthorized unfollow request", "followee_id", followeeID)
}

func TestUnfollowUserInvalidToken(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, _, followeeID, teardown := setupUnfollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken("invalid_token_12345")

	unfollowResp, err := tc.RelationClient.Unfollow(followeeID)

	require.Error(t, err, "Should fail with invalid token")
	assert.Contains(t, err.Error(), custom_errors.ErrInvalidToken.Error(), "Error should be unauthenticated")
	assert.Nil(t, unfollowResp, "Response should be nil on error")

	log.Info("Correctly rejected invalid token unfollow request", "followee_id", followeeID)
}

func TestUnfollowUserSelf(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, _, _, teardown := setupUnfollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	unfollowResp, err := tc.RelationClient.Unfollow(followerID)

	require.Error(t, err, "Should fail when trying to unfollow self")
	assert.Contains(t, err.Error(), custom_errors.ErrSelfUnfollow.Error(), "Error should be self unfollow")
	assert.Nil(t, unfollowResp, "Response should be nil on error")

	log.Info("Correctly rejected self-unfollow request", "user_id", followerID)
}

func TestUnfollowUserRelationNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, _, followeeID, teardown := setupUnfollowUserTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	unfollowResp, err := tc.RelationClient.Unfollow(followeeID)

	require.Error(t, err, "Should fail when follow relation doesn't exist")
	assert.Contains(t, err.Error(), custom_errors.ErrFollowRelationNotFound.Error(), "Error should be follow relation not found")
	assert.Nil(t, unfollowResp, "Response should be nil on error")

	log.Info("Correctly rejected unfollow request for non-existent relation", "follower_id", followerID, "followee_id", followeeID)
}

func TestUnfollowUserValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, _, _, _, teardown := setupUnfollowUserTest(t, tc)
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
			unfollowResp, err := tc.RelationClient.Unfollow(testCase.followeeID)

			require.Error(t, err, "Should fail with %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedErr.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, unfollowResp, "Response should be nil on error for %s", testCase.description)

			log.Info("Correctly rejected invalid followee ID", "test_case", testCase.name, "followee_id", testCase.followeeID)
		})
	}
}

func TestUnfollowUserDoubleUnfollow(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, _, followeeID, teardown := setupUnfollowUserTestWithExistingRelation(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerToken)

	unfollowResp1, err := tc.RelationClient.Unfollow(followeeID)
	require.NoError(t, err, "First unfollow should succeed")
	assert.NotNil(t, unfollowResp1, "First response should not be nil")

	unfollowResp2, err := tc.RelationClient.Unfollow(followeeID)

	require.Error(t, err, "Should fail when unfollowing already unfollowed user")
	assert.Contains(t, err.Error(), custom_errors.ErrFollowRelationNotFound.Error(), "Error should be follow relation not found")
	assert.Nil(t, unfollowResp2, "Second response should be nil on error")

	log.Info("Correctly rejected duplicate unfollow request", "follower_id", followerID, "followee_id", followeeID)
}

func TestUnfollowUserWithoutNotificationGeneration(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerToken, followerID, followeeToken, followeeID, teardown := setupUnfollowUserTestWithExistingRelation(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followeeToken)
	initialFeedResp, err := tc.NotificationClient.GetUserNotificationFeed(followeeID, 1, 10)
	require.NoError(t, err, "Failed to get initial notification feed")
	initialNotificationCount := len(initialFeedResp.Notifications)

	tc.APIClient.SetToken(followerToken)
	unfollowResp, err := tc.RelationClient.Unfollow(followeeID)
	require.NoError(t, err, "Failed to unfollow user")
	assert.NotNil(t, unfollowResp, "Response should not be nil")

	tc.APIClient.SetToken(followeeToken)
	finalFeedResp, err := tc.NotificationClient.GetUserNotificationFeed(followeeID, 1, 10)
	require.NoError(t, err, "Failed to get final notification feed")
	finalNotificationCount := len(finalFeedResp.Notifications)

	assert.Equal(t, initialNotificationCount, finalNotificationCount, "Unfollow should not create new notifications")

	log.Info("Successfully verified unfollow user without notification generation",
		"follower_id", followerID,
		"followee_id", followeeID,
		"initial_notifications", initialNotificationCount,
		"final_notifications", finalNotificationCount)
}
