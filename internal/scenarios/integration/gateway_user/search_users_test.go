package gateway_user

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSearchUsersTest(t *testing.T, tc *TestContext) (string, []fixtures.User, func()) {
	t.Helper()

	// Create multiple test users with known usernames for search testing
	var users []fixtures.User
	userCount := 3

	log.Info("Setting up search users test", "test", t.Name(), "user_count", userCount)

	// Generate a common prefix for test users to make them easier to search
	searchPrefix := "searchtest"

	for i := 0; i < userCount; i++ {
		registerReq := fixtures.GenerateRegisterRequest()
		// Add a prefix to make sure we can search for these specific users
		registerReq.Username = searchPrefix + registerReq.Username

		tokens, err := tc.AuthClient.Register(*registerReq)
		require.NoError(t, err, "Failed to register test user")

		user, err := tc.UserClient.GetUserByUsername(registerReq.Username)
		require.NoError(t, err, "Failed to get user info")

		tc.TrackUserForCleanup(user.ID, user.Username, tokens.AccessToken)
		users = append(users, *user)
	}

	// Register one more user for authentication
	registerReq := fixtures.GenerateRegisterRequest()
	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register auth user")

	authUser, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get auth user info")

	tc.TrackUserForCleanup(authUser.ID, authUser.Username, tokens.AccessToken)

	return tokens.AccessToken, users, func() {
		log.Info("Search users test complete, local cleanup", "test", t.Name())
	}
}

func TestSearchUsersSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, testUsers, teardown := setupSearchUsersTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	t.Run("SearchByUsername", func(t *testing.T) {
		// Search by the common prefix we added to test users
		searchQuery := "searchtest"
		response, err := tc.UserClient.SearchUsers(searchQuery, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, response)

		// We should find at least the number of users we created
		assert.GreaterOrEqual(t, len(response.Users), len(testUsers))

		// Verify that all our test users are included in the results
		var foundCount int
		for _, testUser := range testUsers {
			for _, resultUser := range response.Users {
				if testUser.ID == resultUser.ID {
					foundCount++
					break
				}
			}
		}

		assert.Equal(t, len(testUsers), foundCount, "Not all test users were found in search results")
	})

	t.Run("SearchWithPagination", func(t *testing.T) {
		searchQuery := "searchtest"
		// Get first page with only 1 result per page
		page1, err := tc.UserClient.SearchUsers(searchQuery, 1, 1)
		require.NoError(t, err)
		require.NotNil(t, page1)
		assert.Len(t, page1.Users, 1, "Should return exactly 1 result on page 1")

		// Get second page
		page2, err := tc.UserClient.SearchUsers(searchQuery, 2, 1)
		require.NoError(t, err)
		require.NotNil(t, page2)
		assert.Len(t, page2.Users, 1, "Should return exactly 1 result on page 2")

		// Make sure pages contain different users
		assert.NotEqual(t, page1.Users[0].ID, page2.Users[0].ID, "Pages should contain different users")
	})
}

func TestSearchUsersNoResults(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupSearchUsersTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	// Search for a very specific username that doesn't exist
	searchQuery := "thisusershoulddefinitelynotexist12345"
	response, err := tc.UserClient.SearchUsers(searchQuery, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return an empty list, not an error
	assert.Equal(t, 0, len(response.Users))
	assert.Equal(t, 0, response.Total)
}

func TestSearchUsersValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupSearchUsersTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		query       string
		page        int
		limit       int
		expectedErr string
	}{
		{
			name:        "EmptyQuery",
			query:       "",
			page:        1,
			limit:       10,
			expectedErr: "invalid search query",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, teardown := setupSearchUsersTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			_, err := ctx.UserClient.SearchUsers(tc.query, tc.page, tc.limit)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}
