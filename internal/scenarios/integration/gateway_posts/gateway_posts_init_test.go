package gateway_posts

import (
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
)

var (
	cfg *config.Config
	log *logger.Logger
)

type UserCleanupInfo struct {
	ID          int64
	Username    string
	AccessToken string
}

type PostCleanupInfo struct {
	ID          int64
	AuthorID    int64
	AccessToken string
}

type TestContext struct {
	APIClient    *client.Client
	AuthClient   *client.AuthClient
	UserClient   *client.UserClient
	PostClient   *client.PostClient
	CreatedUsers []UserCleanupInfo
	CreatedPosts []PostCleanupInfo
	mu           sync.Mutex
}

func NewTestContext() *TestContext {
	apiClient := client.NewClient(cfg, log)
	return &TestContext{
		APIClient:    apiClient,
		AuthClient:   client.NewAuthClient(apiClient),
		UserClient:   client.NewUserClient(apiClient),
		PostClient:   client.NewPostClient(apiClient),
		CreatedUsers: make([]UserCleanupInfo, 0),
		CreatedPosts: make([]PostCleanupInfo, 0),
	}
}

func (tc *TestContext) TrackUserForCleanup(userID int64, username, accessToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	userInfo := UserCleanupInfo{
		ID:          userID,
		Username:    username,
		AccessToken: accessToken,
	}
	tc.CreatedUsers = append(tc.CreatedUsers, userInfo)
	log.Debug("Added user to cleanup list", "user_id", userID, "username", username)
}

func (tc *TestContext) TrackPostForCleanup(postID, authorID int64, accessToken string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	postInfo := PostCleanupInfo{
		ID:          postID,
		AuthorID:    authorID,
		AccessToken: accessToken,
	}
	tc.CreatedPosts = append(tc.CreatedPosts, postInfo)
	log.Debug("Added post to cleanup list", "post_id", postID, "author_id", authorID)
}

func (tc *TestContext) Cleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Clean up posts first
	if len(tc.CreatedPosts) > 0 {
		log.Info("Starting post cleanup process", "posts_to_delete", len(tc.CreatedPosts))

		var successfulPostDeletions int

		for _, postInfo := range tc.CreatedPosts {
			log.Debug("Attempting to delete post", "post_id", postInfo.ID, "author_id", postInfo.AuthorID)

			// Use the access token of the user who created the post
			tc.APIClient.SetToken(postInfo.AccessToken)

			err := tc.PostClient.DeletePost(postInfo.ID)
			if err != nil {
				log.Warn("Failed to delete post during cleanup",
					"post_id", postInfo.ID,
					"author_id", postInfo.AuthorID,
					"error", err.Error())
			} else {
				log.Debug("Successfully deleted post", "post_id", postInfo.ID)
				successfulPostDeletions++
			}
		}

		log.Info("Post cleanup process completed",
			"successful_deletions", successfulPostDeletions,
			"total_posts", len(tc.CreatedPosts))
	}

	// Then clean up users
	if len(tc.CreatedUsers) > 0 {
		log.Info("Starting user cleanup process", "users_to_delete", len(tc.CreatedUsers))

		var successfulUserDeletions int

		for _, userInfo := range tc.CreatedUsers {
			log.Debug("Attempting to delete user", "user_id", userInfo.ID, "username", userInfo.Username)

			tc.APIClient.SetToken(userInfo.AccessToken)

			err := tc.UserClient.DeleteUser(userInfo.ID)
			if err != nil {
				log.Warn("Failed to delete user during cleanup",
					"user_id", userInfo.ID,
					"username", userInfo.Username,
					"error", err.Error())
			} else {
				log.Debug("Successfully deleted user", "user_id", userInfo.ID, "username", userInfo.Username)
				successfulUserDeletions++
			}
		}

		log.Info("User cleanup process completed",
			"successful_deletions", successfulUserDeletions,
			"total_users", len(tc.CreatedUsers))
	}

	tc.APIClient.SetToken("")
	tc.CreatedPosts = []PostCleanupInfo{}
	tc.CreatedUsers = []UserCleanupInfo{}
}

// Helper function for setting up test users
func setupTestUser(t *testing.T, tc *TestContext) (string, int64, func()) {
	t.Helper()

	registerReq := fixtures.GenerateRegisterRequest()
	log.Info("Setting up test user", "test", t.Name(), "username", registerReq.Username)

	tokens, err := tc.AuthClient.Register(*registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	userByUsername, err := tc.UserClient.GetUserByUsername(registerReq.Username)
	if err != nil {
		t.Fatalf("Failed to get user info for test: %v", err)
	}

	tc.TrackUserForCleanup(userByUsername.ID, userByUsername.Username, tokens.AccessToken)

	return tokens.AccessToken, userByUsername.ID, func() {
		log.Info("Test complete, local cleanup", "test", t.Name())
		tc.APIClient.SetToken("")
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")
	log = logger.New(cfg.Env)
	log.Info("Starting posts gateway tests", "env", cfg.Env)

	code := m.Run()

	os.Exit(code)
}
