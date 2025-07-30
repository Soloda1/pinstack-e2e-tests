package client

import (
	"log/slog"

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
