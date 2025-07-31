package gateway_posts

import (
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCreatePostTest(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	// Register a new user for post creation tests
	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up create post test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for post creation test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Create post test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestCreatePostSuccess(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, userID, teardown := setupCreatePostTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	// Create a post request with title, content, tags, and media
	postReq := fixtures.GenerateCreatePostRequest()

	// Send the request
	createdPost, err := tc.PostClient.CreatePost(*postReq)
	require.NoError(t, err)

	// Track post for cleanup
	tc.TrackPostForCleanup(createdPost.ID, userID, accessToken)

	// Validate the response
	assert.NotEqual(t, 0, createdPost.ID, "Post ID should not be zero")
	assert.Equal(t, postReq.Title, createdPost.Title, "Post title should match the request")
	assert.Equal(t, postReq.Content, createdPost.Content, "Post content should match the request")
	assert.Equal(t, userID, createdPost.AuthorID, "Author ID should match the authenticated user")

	// Validate media items
	assert.Equal(t, len(postReq.MediaItems), len(createdPost.Media), "Number of media items should match")
	if len(postReq.MediaItems) > 0 {
		mediaMap := make(map[string]bool)
		for _, item := range createdPost.Media {
			mediaMap[item.URL] = true
		}
		for _, requestItem := range postReq.MediaItems {
			assert.True(t, mediaMap[requestItem.URL], "Media URL should be present in response")
		}
	}

	// Validate tags
	assert.Equal(t, len(postReq.Tags), len(createdPost.Tags), "Number of tags should match")
	if len(postReq.Tags) > 0 {
		tagMap := make(map[string]bool)
		for _, tag := range createdPost.Tags {
			tagMap[tag.Name] = true
		}
		for _, requestTag := range postReq.Tags {
			assert.True(t, tagMap[requestTag], "Tag name should be present in response")
		}
	}

	// Validate that the post can be retrieved via GetPostByID
	retrievedPost, err := tc.PostClient.GetPostByID(createdPost.ID)
	require.NoError(t, err)

	assert.Equal(t, createdPost.ID, retrievedPost.ID, "Retrieved post ID should match created post ID")
	assert.Equal(t, createdPost.Title, retrievedPost.Title, "Retrieved post title should match")
	assert.Equal(t, createdPost.Content, retrievedPost.Content, "Retrieved post content should match")
}

func TestCreatePostUnauthorized(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, teardown := setupCreatePostTest(t, tc)
	defer teardown()

	// Clear the token
	tc.APIClient.SetToken("")

	postReq := fixtures.GenerateCreatePostRequest()

	// Attempt to create a post without authentication
	_, err := tc.PostClient.CreatePost(*postReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthenticated")

	// Try with invalid token
	tc.APIClient.SetToken("invalid_token")
	_, err = tc.PostClient.CreatePost(*postReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestCreatePostValidationErrors(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, teardown := setupCreatePostTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	testCases := []struct {
		name        string
		postReq     fixtures.CreatePostRequest
		expectedErr string
	}{
		{
			name: "EmptyTitle",
			postReq: fixtures.CreatePostRequest{
				Title:   "",
				Content: "Test content",
			},
			expectedErr: "validation failed",
		},
		{
			name: "TitleTooLong",
			postReq: fixtures.CreatePostRequest{
				Title:   string(make([]byte, 256)), // Create a very long title
				Content: "Test content",
			},
			expectedErr: "validation failed",
		},
		{
			name: "InvalidMediaType",
			postReq: fixtures.CreatePostRequest{
				Title:   "Test Title",
				Content: "Test content",
				MediaItems: []fixtures.MediaItemInput{
					{
						Type: "invalid_type",
						URL:  "https://example.com/image.jpg",
					},
				},
			},
			expectedErr: "validation failed",
		},
		{
			name: "InvalidMediaURL",
			postReq: fixtures.CreatePostRequest{
				Title:   "Test Title",
				Content: "Test content",
				MediaItems: []fixtures.MediaItemInput{
					{
						Type: "image",
						URL:  "not-a-valid-url",
					},
				},
			},
			expectedErr: "validation failed",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewTestContext()
			defer ctx.Cleanup()

			accessToken, _, teardown := setupCreatePostTest(t, ctx)
			defer teardown()

			ctx.APIClient.SetToken(accessToken)

			_, err := ctx.PostClient.CreatePost(tc.postReq)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}
