package gateway_posts

import (
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUpdatePostTest(t *testing.T, tc *TestContext) (string, int64, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up update post test", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err, "Failed to register test user")

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err, "Failed to get user info for update post test")

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)
	tc.APIClient.SetToken(tokens.AccessToken)

	postReq := fixtures.GenerateCreatePostRequest()
	createdPost, err := tc.PostClient.CreatePost(*postReq)
	require.NoError(t, err, "Failed to create test post")

	tc.TrackPostForCleanup(createdPost.ID, userByUsername.ID, tokens.AccessToken)

	log.Info("Created test post for update test",
		"post_id", createdPost.ID,
		"title", createdPost.Title,
		"author_id", userByUsername.ID)

	return tokens.AccessToken, userByUsername.ID, createdPost.ID, func() {
		log.Info("Update post test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestUpdatePost(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, authorID, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	updateReq := fixtures.GenerateUpdatePostRequest()

	updatedPost, err := tc.PostClient.UpdatePost(postID, *updateReq)
	require.NoError(t, err, "Failed to update post")

	assert.Equal(t, postID, updatedPost.ID, "Post ID should not change")
	assert.Equal(t, authorID, updatedPost.Author.ID, "Post author should not change")
	assert.Equal(t, updateReq.Title, updatedPost.Title, "Title should be updated")
	assert.Equal(t, updateReq.Content, updatedPost.Content, "Content should be updated")

	assert.Equal(t, len(updateReq.Tags), len(updatedPost.Tags), "Should have the same number of tags")
	tagMap := make(map[string]bool)
	for _, tag := range updatedPost.Tags {
		tagMap[tag.Name] = true
	}
	for _, tag := range updateReq.Tags {
		assert.True(t, tagMap[tag], "Updated post should contain tag %s", tag)
	}

	assert.Equal(t, updateReq.Title, updatedPost.Title)
	assert.Equal(t, updateReq.Content, updatedPost.Content)
	assert.Equal(t, authorID, updatedPost.Author.ID)

	tagNameMap := make(map[string]bool)
	for _, tag := range updatedPost.Tags {
		tagNameMap[tag.Name] = true
	}

	for _, tagName := range updateReq.Tags {
		assert.True(t, tagNameMap[tagName], "Tag should be in the updated post")
	}

	fetchedPost, err := tc.PostClient.GetPostByID(postID)
	require.NoError(t, err)
	assert.Equal(t, updateReq.Title, fetchedPost.Title)
	assert.Equal(t, updateReq.Content, fetchedPost.Content)
}

func TestUpdatePostNotFound(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	accessToken, _, _, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	tc.APIClient.SetToken(accessToken)

	nonExistentPostID := int64(999999) // Use a very large ID that likely doesn't exist
	updateReq := fixtures.GenerateUpdatePostRequest()

	_, err := tc.PostClient.UpdatePost(nonExistentPostID, *updateReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrPostNotFound.Error())
}

func TestUpdatePostForbidden(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Registering second user for forbidden test", "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	require.NoError(t, err)

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	require.NoError(t, err)

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	tc.APIClient.SetToken(tokens.AccessToken)

	updateReq := fixtures.GenerateUpdatePostRequest()

	_, err = tc.PostClient.UpdatePost(postID, *updateReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), custom_errors.ErrForbidden.Error())
}

func TestUpdatePostWithInvalidData(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	updateReq := fixtures.UpdatePostRequest{
		Title:   "", // Empty title should be rejected
		Content: "Valid content",
	}

	_, err := tc.PostClient.UpdatePost(postID, updateReq)
	if err != nil {
		assert.Contains(t, err.Error(), custom_errors.ErrValidationFailed.Error())
	}

	extremelyLongContent := ""
	for i := 0; i < 10000; i++ {
		extremelyLongContent += "too long content "
	}

	updateReq = fixtures.UpdatePostRequest{
		Title:   "Valid Title",
		Content: extremelyLongContent,
	}

	_, err = tc.PostClient.UpdatePost(postID, updateReq)
	if err != nil {
		assert.Contains(t, err.Error(), custom_errors.ErrValidationFailed.Error())
	}
}

func TestPartialUpdatePost(t *testing.T) {
	t.Parallel()
	tc := NewTestContext()
	defer tc.Cleanup()

	_, _, postID, teardown := setupUpdatePostTest(t, tc)
	defer teardown()

	originalPost, err := tc.PostClient.GetPostByID(postID)
	require.NoError(t, err)

	titleOnlyUpdate := fixtures.UpdatePostRequest{
		Title: "Updated Title Only",
	}

	updatedPost, err := tc.PostClient.UpdatePost(postID, titleOnlyUpdate)
	require.NoError(t, err)

	assert.Equal(t, titleOnlyUpdate.Title, updatedPost.Title)
	assert.Equal(t, originalPost.Content, updatedPost.Content)

	contentOnlyUpdate := fixtures.UpdatePostRequest{
		Content: "This is updated content only",
	}

	updatedPost, err = tc.PostClient.UpdatePost(postID, contentOnlyUpdate)
	require.NoError(t, err)

	assert.Equal(t, titleOnlyUpdate.Title, updatedPost.Title) // Title should remain from previous update
	assert.Equal(t, contentOnlyUpdate.Content, updatedPost.Content)
}
