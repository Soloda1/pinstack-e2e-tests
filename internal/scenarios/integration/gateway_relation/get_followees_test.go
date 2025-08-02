package gateway_relation

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGetFolloweesTest(t *testing.T, tc *TestContext) (followerUserToken string, followerUserID int64, followeeTokens []string, followeeIDs []int64, teardown func()) {
	t.Helper()

	followerRegisterReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up get followees test - registering follower user", "test", t.Name(), "username", followerRegisterReq.Username)

	followerTokens, err := tc.AuthClient.Register(*followerRegisterReq)
	require.NoError(t, err, "Failed to register follower user")

	followerUser, err := tc.UserClient.GetUserByUsername(followerRegisterReq.Username)
	require.NoError(t, err, "Failed to get follower user info")

	tc.TrackUserForCleanup(followerUser.ID, followerUser.Username, followerTokens.AccessToken)

	numFollowees := 3
	var followeeTokensList []string
	var followeeIDsList []int64

	for i := 0; i < numFollowees; i++ {
		followeeRegisterReq := fixtures.GenerateRegisterRequest()
		log.Info("Setting up get followees test - registering followee", "test", t.Name(), "followee_index", i, "username", followeeRegisterReq.Username)

		followeeTokens, err := tc.AuthClient.Register(*followeeRegisterReq)
		require.NoError(t, err, "Failed to register followee user %d", i)

		followeeUser, err := tc.UserClient.GetUserByUsername(followeeRegisterReq.Username)
		require.NoError(t, err, "Failed to get followee user info %d", i)

		tc.TrackUserForCleanup(followeeUser.ID, followeeUser.Username, followeeTokens.AccessToken)

		followeeTokensList = append(followeeTokensList, followeeTokens.AccessToken)
		followeeIDsList = append(followeeIDsList, followeeUser.ID)
	}

	return followerTokens.AccessToken, followerUser.ID, followeeTokensList, followeeIDsList, func() {
		log.Info("Get followees test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestGetFolloweesSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerUserToken, followerUserID, followeeTokens, followeeIDs, teardown := setupGetFolloweesTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerUserToken)
	for i, followeeID := range followeeIDs {
		_, err := tc.RelationClient.Follow(followeeID)
		require.NoError(t, err, "Failed to create follow relation for followee %d", i)

		tc.TrackRelationForCleanup(followerUserID, followeeID, followerUserToken)
	}

	for i, followeeToken := range followeeTokens {
		tc.DiscoverAndTrackAllNotifications(followeeIDs[i], followeeToken)
	}

	tc.APIClient.SetToken(followerUserToken)

	followeesResp, err := tc.RelationClient.GetFollowees(followerUserID, 1, 10)

	require.NoError(t, err, "Failed to get followees")
	require.NotNil(t, followeesResp, "Response should not be nil")

	assert.Equal(t, 1, followeesResp.Page, "Page should be 1")
	assert.Equal(t, 10, followeesResp.Limit, "Limit should be 10")
	assert.Equal(t, len(followeeIDs), followeesResp.Total, "Total should match number of followees")
	assert.Len(t, followeesResp.Followees, len(followeeIDs), "Should return all followees")

	returnedFolloweeIDs := make(map[int64]bool)
	for _, followeeID := range followeesResp.Followees {
		returnedFolloweeIDs[followeeID] = true
	}

	for _, expectedFolloweeID := range followeeIDs {
		assert.True(t, returnedFolloweeIDs[expectedFolloweeID], "Expected followee %d should be in response", expectedFolloweeID)
	}

	log.Info("Successfully retrieved followees", "follower_user_id", followerUserID, "followees_count", len(followeesResp.Followees))
}

func TestGetFolloweesEmptyList(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerUserToken, followerUserID, _, _, teardown := setupGetFolloweesTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerUserToken)

	followeesResp, err := tc.RelationClient.GetFollowees(followerUserID, 1, 10)

	require.NoError(t, err, "Should succeed even with no followees")
	require.NotNil(t, followeesResp, "Response should not be nil")

	assert.Equal(t, 1, followeesResp.Page, "Page should be 1")
	assert.Equal(t, 10, followeesResp.Limit, "Limit should be 10")
	assert.Equal(t, 0, followeesResp.Total, "Total should be 0")
	assert.Empty(t, followeesResp.Followees, "Followees list should be empty")

	log.Info("Successfully handled empty followees list", "follower_user_id", followerUserID)
}

func TestGetFolloweesUserNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerUserToken, _, _, _, teardown := setupGetFolloweesTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerUserToken)

	nonExistentUserID := int64(999999)
	followeesResp, err := tc.RelationClient.GetFollowees(nonExistentUserID, 1, 10)

	require.Error(t, err, "Should fail for non-existent user")
	assert.Contains(t, err.Error(), custom_errors.ErrUserNotFound.Error(), "Error should be user not found")
	assert.Nil(t, followeesResp, "Response should be nil on error")

	log.Info("Correctly rejected get followees request for non-existent user", "user_id", nonExistentUserID)
}

func TestGetFolloweesPagination(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerUserToken, followerUserID, followeeTokens, followeeIDs, teardown := setupGetFolloweesTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(followerUserToken)
	for i, followeeID := range followeeIDs {
		_, err := tc.RelationClient.Follow(followeeID)
		require.NoError(t, err, "Failed to create follow relation for followee %d", i)

		tc.TrackRelationForCleanup(followerUserID, followeeID, followerUserToken)
	}

	for i, followeeToken := range followeeTokens {
		tc.DiscoverAndTrackAllNotifications(followeeIDs[i], followeeToken)
	}

	tc.APIClient.SetToken(followerUserToken)

	followeesResp, err := tc.RelationClient.GetFollowees(followerUserID, 1, 2)

	require.NoError(t, err, "Failed to get followees with pagination")
	require.NotNil(t, followeesResp, "Response should not be nil")

	assert.Equal(t, 1, followeesResp.Page, "Page should be 1")
	assert.Equal(t, 2, followeesResp.Limit, "Limit should be 2")
	assert.LessOrEqual(t, len(followeesResp.Followees), 2, "Should return at most 2 followees")

	log.Info("Successfully tested followees pagination",
		"follower_user_id", followerUserID,
		"page", followeesResp.Page,
		"limit", followeesResp.Limit,
		"returned_count", len(followeesResp.Followees),
		"total", followeesResp.Total)
}

func TestGetFolloweesValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	followerUserToken, _, _, _, teardown := setupGetFolloweesTest(t, tc)
	defer teardown()

	// Set valid token
	tc.APIClient.SetToken(followerUserToken)

	// Test cases for validation errors
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
			followeesResp, err := tc.RelationClient.GetFollowees(testCase.userID, testCase.page, testCase.limit)

			require.Error(t, err, "Should fail with %s", testCase.description)
			assert.Contains(t, err.Error(), testCase.expectedErr.Error(), "Error should match expected type for %s", testCase.description)
			assert.Nil(t, followeesResp, "Response should be nil on error for %s", testCase.description)

			log.Info("Correctly rejected invalid parameters", "test_case", testCase.name)
		})
	}
}
