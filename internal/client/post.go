package client

import (
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/Soloda1/pinstack-e2e-tests/internal/fixtures"
)

type PostClient struct {
	client *Client
}

func NewPostClient(client *Client) *PostClient {
	return &PostClient{
		client: client,
	}
}

func (pc *PostClient) CreatePost(req fixtures.CreatePostRequest) (*fixtures.CreatePostResponse, error) {
	pc.client.log.Info("Creating new post",
		slog.String("title", req.Title),
		slog.Int("media_count", len(req.MediaItems)),
		slog.Int("tags_count", len(req.Tags)),
	)

	var response fixtures.CreatePostResponse
	err := pc.client.Post("/v1/posts", req, &response)
	if err != nil {
		pc.client.log.Error("Failed to create post",
			slog.String("title", req.Title),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	pc.client.log.Info("Post created successfully",
		slog.Int64("post_id", response.ID),
		slog.String("title", response.Title),
	)
	return &response, nil
}

func (pc *PostClient) GetPostByID(postID int64) (*fixtures.Post, error) {
	pc.client.log.Debug("Getting post by ID", slog.Int64("post_id", postID))

	var response fixtures.Post
	err := pc.client.Get("/v1/posts/"+strconv.FormatInt(postID, 10), nil, &response)
	if err != nil {
		pc.client.log.Error("Failed to get post by ID",
			slog.Int64("post_id", postID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	pc.client.log.Debug("Got post by ID successfully",
		slog.Int64("post_id", postID),
		slog.String("title", response.Title),
	)
	return &response, nil
}

func (pc *PostClient) UpdatePost(postID int64, req fixtures.UpdatePostRequest) (*fixtures.UpdatePostResponse, error) {
	pc.client.log.Info("Updating post",
		slog.Int64("post_id", postID),
		slog.String("title", req.Title),
	)

	var response fixtures.UpdatePostResponse
	err := pc.client.Put("/v1/posts/"+strconv.FormatInt(postID, 10), req, &response)
	if err != nil {
		pc.client.log.Error("Failed to update post",
			slog.Int64("post_id", postID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	pc.client.log.Info("Post updated successfully",
		slog.Int64("post_id", postID),
		slog.String("title", response.Title),
	)
	return &response, nil
}

func (pc *PostClient) DeletePost(postID int64) error {
	pc.client.log.Info("Deleting post", slog.Int64("post_id", postID))

	err := pc.client.Delete("/v1/posts/"+strconv.FormatInt(postID, 10), nil)
	if err != nil {
		pc.client.log.Error("Failed to delete post",
			slog.Int64("post_id", postID),
			slog.String("error", err.Error()),
		)
		return err
	}

	pc.client.log.Info("Post deleted successfully", slog.Int64("post_id", postID))
	return nil
}

func (pc *PostClient) ListPosts(authorID int64, createdAfter, createdBefore time.Time, offset, limit int) (*fixtures.ListPostsResponse, error) {
	pc.client.log.Debug("Listing posts",
		slog.Int64("author_id", authorID),
		slog.Int("offset", offset),
		slog.Int("limit", limit),
	)

	queryParams := url.Values{}

	if authorID > 0 {
		queryParams.Set("author_id", strconv.FormatInt(authorID, 10))
	}

	if !createdAfter.IsZero() {
		queryParams.Set("created_after", createdAfter.Format(time.RFC3339))
		pc.client.log.Debug("Filter: created after", slog.String("date", createdAfter.Format(time.RFC3339)))
	}

	if !createdBefore.IsZero() {
		queryParams.Set("created_before", createdBefore.Format(time.RFC3339))
		pc.client.log.Debug("Filter: created before", slog.String("date", createdBefore.Format(time.RFC3339)))
	}

	if offset > 0 {
		queryParams.Set("offset", strconv.Itoa(offset))
	}

	if limit > 0 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}

	var response fixtures.ListPostsResponse
	err := pc.client.Get("/v1/posts/list", queryParams, &response)
	if err != nil {
		pc.client.log.Error("Failed to list posts",
			slog.Int64("author_id", authorID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	pc.client.log.Debug("Posts listed successfully",
		slog.Int("total_count", response.Total),
		slog.Int("returned_count", len(response.Posts)),
	)
	return &response, nil
}
