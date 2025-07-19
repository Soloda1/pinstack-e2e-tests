package client

import (
	"github.com/Soloda1/pinstack-e2e-tests/internal/custom_errors"
	"log/slog"
)

type RelationClient struct {
	client *Client
}

func NewRelationClient(client *Client) *RelationClient {
	return &RelationClient{
		client: client,
	}
}

func (rc *RelationClient) Follow(followeeID int) (*FollowResponse, error) {
	rc.client.log.Info("Following user", slog.Int("followee_id", followeeID))

	req := FollowRequest{
		FolloweeID: followeeID,
	}

	var response FollowResponse
	err := rc.client.Post("/v1/relation/follow", req, &response)
	if err != nil {
		rc.client.log.Error("Failed to follow user",
			slog.Int("followee_id", followeeID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrFollowRelationCreateFail
	}

	rc.client.log.Info("User followed successfully",
		slog.Int("followee_id", followeeID),
		slog.String("message", response.Message),
	)
	return &response, nil
}

func (rc *RelationClient) Unfollow(followeeID int) (*UnfollowResponse, error) {
	rc.client.log.Info("Unfollowing user", slog.Int("followee_id", followeeID))

	req := UnfollowRequest{
		FolloweeID: followeeID,
	}

	var response UnfollowResponse
	err := rc.client.Post("/v1/relation/unfollow", req, &response)
	if err != nil {
		rc.client.log.Error("Failed to unfollow user",
			slog.Int("followee_id", followeeID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrFollowRelationDeleteFail
	}

	rc.client.log.Info("User unfollowed successfully",
		slog.Int("followee_id", followeeID),
		slog.String("message", response.Message),
	)
	return &response, nil
}
