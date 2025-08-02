package client

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
)

type RelationClient struct {
	client *Client
}

func NewRelationClient(client *Client) *RelationClient {
	return &RelationClient{
		client: client,
	}
}

func (rc *RelationClient) Follow(followeeID int64) (*fixtures.FollowResponse, error) {
	rc.client.log.Info("Following user", slog.Int64("followee_id", followeeID))

	req := fixtures.FollowRequest{
		FolloweeID: followeeID,
	}

	var response fixtures.FollowResponse
	err := rc.client.Post("/v1/relation/follow", req, &response)
	if err != nil {
		rc.client.log.Error("Failed to follow user",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	rc.client.log.Info("User followed successfully",
		slog.Int64("followee_id", followeeID),
		slog.String("message", response.Message),
	)
	return &response, nil
}

func (rc *RelationClient) Unfollow(followeeID int64) (*fixtures.UnfollowResponse, error) {
	rc.client.log.Info("Unfollowing user", slog.Int64("followee_id", followeeID))

	req := fixtures.UnfollowRequest{
		FolloweeID: followeeID,
	}

	var response fixtures.UnfollowResponse
	err := rc.client.Post("/v1/relation/unfollow", req, &response)
	if err != nil {
		rc.client.log.Error("Failed to unfollow user",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	rc.client.log.Info("User unfollowed successfully",
		slog.Int64("followee_id", followeeID),
		slog.String("message", response.Message),
	)
	return &response, nil
}

func (rc *RelationClient) GetFollowers(userID int64, page, limit int) (*fixtures.GetFollowersResponse, error) {
	rc.client.log.Info("Getting user followers",
		slog.Int64("user_id", userID),
		slog.Int("page", page),
		slog.Int("limit", limit))

	path := fmt.Sprintf("/v1/relation/%d/followers", userID)
	queryParams := url.Values{}
	queryParams.Add("page", strconv.Itoa(page))
	queryParams.Add("limit", strconv.Itoa(limit))

	var response fixtures.GetFollowersResponse
	err := rc.client.Get(path, queryParams, &response)
	if err != nil {
		rc.client.log.Error("Failed to get user followers",
			slog.Int64("user_id", userID),
			slog.Int("page", page),
			slog.Int("limit", limit),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	rc.client.log.Info("User followers retrieved successfully",
		slog.Int64("user_id", userID),
		slog.Int("followers_count", len(response.Followers)),
		slog.Int("total", response.Total),
		slog.Int("page", response.Page),
	)
	return &response, nil
}

func (rc *RelationClient) GetFollowees(userID int64, page, limit int) (*fixtures.GetFolloweesResponse, error) {
	rc.client.log.Info("Getting user followees",
		slog.Int64("user_id", userID),
		slog.Int("page", page),
		slog.Int("limit", limit))

	path := fmt.Sprintf("/v1/relation/%d/followees", userID)
	queryParams := url.Values{}
	queryParams.Add("page", strconv.Itoa(page))
	queryParams.Add("limit", strconv.Itoa(limit))

	var response fixtures.GetFolloweesResponse
	err := rc.client.Get(path, queryParams, &response)
	if err != nil {
		rc.client.log.Error("Failed to get user followees",
			slog.Int64("user_id", userID),
			slog.Int("page", page),
			slog.Int("limit", limit),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	rc.client.log.Info("User followees retrieved successfully",
		slog.Int64("user_id", userID),
		slog.Int("followees_count", len(response.Followees)),
		slog.Int("total", response.Total),
		slog.Int("page", response.Page),
	)
	return &response, nil
}
