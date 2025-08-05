package gateway_relation

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetFollowersTest(t *testing.T, tc *TestContext) (targetUserToken string, targetUserID int64, followerTokens []string, followerIDs []int64, teardown func()) {
	t.Helper()

	targetRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get followers test - registering target user", "test", t.Name(), "username", targetRegisterReq.Username)

	targetTokens, err := tc.AuthClient.Register(*targetRegisterReq)
	require.NoError(t, err, "Failed to register target user")

	targetUser, err := tc.UserClient.GetUserByUsername(targetRegisterReq.Username)
	require.NoError(t, err, "Failed to get target user info")

	tc.TrackUserForCleanup(targetUser.ID, targetUser.Username, targetTokens.AccessToken)

	numFollowers := 3
	var followerTokensList []string
	var followerIDsList []int64

	for i := 0; i < numFollowers; i++ {
		followerRegisterReq := fixtures.GenerateRegisterRequest()
		log.Info("Setting up get followers test - registering follower", "test", t.Name(), "follower_index", i, "username", followerRegisterReq.Username)

		followerTokens, err := tc.AuthClient.Register(*followerRegisterReq)
		require.NoError(t, err, "Failed to register follower user %d", i)

		followerUser, err := tc.UserClient.GetUserByUsername(followerRegisterReq.Username)
		require.NoError(t, err, "Failed to get follower user info %d", i)

		tc.TrackUserForCleanup(followerUser.ID, followerUser.Username, followerTokens.AccessToken)

		followerTokensList = append(followerTokensList, followerTokens.AccessToken)
		followerIDsList = append(followerIDsList, followerUser.ID)
	}

	return targetTokens.AccessToken, targetUser.ID, followerTokensList, followerIDsList, func() {
		log.Info("Get followers test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetFollowersSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	targetUserToken, targetUserID, followerTokens, followerIDs, teardown := setupGetFollowersTest(t, tc)
	defer teardown()

	for i, followerToken := range followerTokens {
		tc.APIClient.SetToken(followerToken)
		followResp, err := tc.RelationClient.Follow(targetUserID)
		require.NoError(t, err, "Failed to create follow relation for follower %d", i)
		require.NotNil(t, followResp, "Follow response should not be nil")

		tc.TrackRelationForCleanup(followerIDs[i], targetUserID, followerToken)
	}

	tc.DiscoverAndTrackAllNotifications(targetUserID, targetUserToken)

	tc.APIClient.SetToken(targetUserToken)

	followersResp, err := tc.RelationClient.GetFollowers(targetUserID, 1, 10)

	require.NoError(t, err, "Failed to get followers")
	require.NotNil(t, followersResp, "Response should not be nil")

	assert.Equal(t, int32(1), followersResp.Page, "Page should be 1")
	assert.Equal(t, int32(10), followersResp.Limit, "Limit should be 10")
	assert.Equal(t, int64(len(followerIDs)), followersResp.Total, "Total should match number of followers")
	assert.Len(t, followersResp.Followers, len(followerIDs), "Should return all followers")

	// Verify all expected followers are in the response
	returnedFollowerIDs := make(map[int64]bool)
	for _, follower := range followersResp.Followers {
		returnedFollowerIDs[follower.ID] = true
	}

	for _, expectedFollowerID := range followerIDs {
		assert.True(t, returnedFollowerIDs[expectedFollowerID], "Expected follower %d should be in response", expectedFollowerID)
	}

	log.Info("Successfully retrieved followers", "target_user_id", targetUserID, "followers_count", len(followersResp.Followers))
}

func TestGetFollowersEmptyList(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	targetUserToken, targetUserID, _, _, teardown := setupGetFollowersTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(targetUserToken)

	followersResp, err := tc.RelationClient.GetFollowers(targetUserID, 1, 10)

	require.NoError(t, err, "Should succeed even with no followers")
	require.NotNil(t, followersResp, "Response should not be nil")

	assert.Equal(t, int32(1), followersResp.Page, "Page should be 1")
	assert.Equal(t, int32(10), followersResp.Limit, "Limit should be 10")
	assert.Equal(t, int64(0), followersResp.Total, "Total should be 0")
	assert.Empty(t, followersResp.Followers, "Followers list should be empty")

	log.Info("Successfully handled empty followers list", "target_user_id", targetUserID)
}

func TestGetFollowersUserNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	targetUserToken, _, _, _, teardown := setupGetFollowersTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(targetUserToken)

	nonExistentUserID := int64(999999)
	followersResp, err := tc.RelationClient.GetFollowers(nonExistentUserID, 1, 10)

	require.Error(t, err, "Should fail for non-existent user")
	assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error(), "Error should be user not found")
	assert.Nil(t, followersResp, "Response should be nil on error")

	log.Info("Correctly rejected get followers request for non-existent user", "user_id", nonExistentUserID)
}

func TestGetFollowersPagination(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	targetUserToken, targetUserID, followerTokens, followerIDs, teardown := setupGetFollowersTest(t, tc)
	defer teardown()

	for i, followerToken := range followerTokens {
		tc.APIClient.SetToken(followerToken)
		_, err := tc.RelationClient.Follow(targetUserID)
		require.NoError(t, err, "Failed to create follow relation for follower %d", i)

		tc.TrackRelationForCleanup(followerIDs[i], targetUserID, followerToken)
	}

	tc.DiscoverAndTrackAllNotifications(targetUserID, targetUserToken)

	tc.APIClient.SetToken(targetUserToken)

	followersResp, err := tc.RelationClient.GetFollowers(targetUserID, 1, 2)

	require.NoError(t, err, "Failed to get followers with pagination")
	require.NotNil(t, followersResp, "Response should not be nil")

	assert.Equal(t, int32(1), followersResp.Page, "Page should be 1")
	assert.Equal(t, int32(2), followersResp.Limit, "Limit should be 2")
	assert.LessOrEqual(t, len(followersResp.Followers), 2, "Should return at most 2 followers")

	log.Info("Successfully tested followers pagination",
		"target_user_id", targetUserID,
		"page", followersResp.Page,
		"limit", followersResp.Limit,
		"returned_count", len(followersResp.Followers),
		"total", followersResp.Total)
}

func TestGetFollowersValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	targetUserToken, _, _, _, teardown := setupGetFollowersTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(targetUserToken)

	testCases := []struct {
		name        string
		userID      int64
		page        int
		limit       int
		description string
		expectedErr error
	}{
		{
			name:        "zero_user_id",
			userID:      0,
			page:        1,
			limit:       10,
			description: "zero user ID",
			expectedErr: custom_errors.ErrValidationFailed,
		},
		{
			name:        "negative_user_id",
			userID:      -1,
			page:        1,
			limit:       10,
			description: "negative user ID",
			expectedErr: custom_errors.ErrValidationFailed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			followersResp, err := tc.RelationClient.GetFollowers(testCase.userID, testCase.page, testCase.limit)

			require.Error(t, err, "Should fail with %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedErr.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, followersResp, "Response should be nil on error for %s", testCase.description)

			log.Info("Correctly rejected invalid parameters", "test_case", testCase.name)
		})
	}
}
