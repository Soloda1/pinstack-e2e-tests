package client

import (
	"log/slog"
	"net/url"
	"strconv"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
)

type UserClient struct {
	client *Client
}

func NewUserClient(client *Client) *UserClient {
	return &UserClient{
		client: client,
	}
}

func (uc *UserClient) CreateUser(req fixtures.CreateUserRequest) (*fixtures.CreateUserResponse, error) {
	uc.client.log.Info("Creating new user", slog.String("username", req.Username), slog.String("email", req.Email))

	var response fixtures.CreateUserResponse
	err := uc.client.Post("/v1/users", req, &response)
	if err != nil {
		uc.client.log.Error("Failed to create user", slog.String("username", req.Username), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Info("User created successfully", slog.Int64("user_id", response.ID), slog.String("username", response.Username))
	return &response, nil
}

func (uc *UserClient) GetUserByID(userID int64) (*fixtures.User, error) {
	uc.client.log.Debug("Getting user by ID", slog.Int64("user_id", userID))

	var response fixtures.User
	err := uc.client.Get("/v1/users/"+strconv.FormatInt(userID, 10), nil, &response)
	if err != nil {
		uc.client.log.Error("Failed to get user by ID", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Debug("Got user by ID successfully", slog.Int64("user_id", userID), slog.String("username", response.Username))
	return &response, nil
}

func (uc *UserClient) GetUserByUsername(username string) (*fixtures.User, error) {
	uc.client.log.Debug("Getting user by username", slog.String("username", username))

	var response fixtures.User
	err := uc.client.Get("/v1/users/username/"+url.PathEscape(username), nil, &response)
	if err != nil {
		uc.client.log.Error("Failed to get user by username", slog.String("username", username), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Debug("Got user by username successfully", slog.String("username", username), slog.Int64("user_id", response.ID))
	return &response, nil
}

func (uc *UserClient) GetUserByEmail(email string) (*fixtures.User, error) {
	uc.client.log.Debug("Getting user by email", slog.String("email", email))

	var response fixtures.User
	err := uc.client.Get("/v1/users/email/"+url.PathEscape(email), nil, &response)
	if err != nil {
		uc.client.log.Error("Failed to get user by email", slog.String("email", email), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Debug("Got user by email successfully", slog.String("email", email), slog.Int64("user_id", response.ID))
	return &response, nil
}

func (uc *UserClient) UpdateUser(req fixtures.UpdateUserRequest) (*fixtures.UpdateUserResponse, error) {
	uc.client.log.Info("Updating user", slog.Int64("user_id", req.ID))

	var response fixtures.UpdateUserResponse
	err := uc.client.Put("/v1/users/"+strconv.FormatInt(req.ID, 10), req, &response)
	if err != nil {
		uc.client.log.Error("Failed to update user", slog.Int64("user_id", req.ID), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Info("User updated successfully", slog.Int64("user_id", req.ID))
	return &response, nil
}

func (uc *UserClient) UpdateAvatar(userID int64, req fixtures.UpdateAvatarRequest) error {
	uc.client.log.Info("Updating user avatar", slog.Int64("user_id", userID))

	err := uc.client.Put("/v1/users/"+strconv.FormatInt(userID, 10)+"/avatar", req, nil)
	if err != nil {
		uc.client.log.Error("Failed to update avatar", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return err
	}

	uc.client.log.Info("User avatar updated successfully", slog.Int64("user_id", userID))
	return nil
}

func (uc *UserClient) DeleteUser(userID int64) error {
	uc.client.log.Info("Deleting user", slog.Int64("user_id", userID))

	err := uc.client.Delete("/v1/users/"+strconv.FormatInt(userID, 10), nil)
	if err != nil {
		uc.client.log.Error("Failed to delete user", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return err
	}

	uc.client.log.Info("User deleted successfully", slog.Int64("user_id", userID))
	return nil
}

func (uc *UserClient) SearchUsers(query string, page, limit int) (*fixtures.SearchUsersResponse, error) {
	uc.client.log.Debug("Searching users",
		slog.String("query", query),
		slog.Int("page", page),
		slog.Int("limit", limit),
	)

	queryParams := url.Values{}
	queryParams.Set("query", query)

	if page > 0 {
		queryParams.Set("page", strconv.Itoa(page))
	}

	if limit > 0 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}

	var response fixtures.SearchUsersResponse
	err := uc.client.Get("/v1/users/search", queryParams, &response)
	if err != nil {
		uc.client.log.Error("Failed to search users",
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	uc.client.log.Debug("User search completed",
		slog.String("query", query),
		slog.Int("results_count", response.Total),
	)
	return &response, nil
}
