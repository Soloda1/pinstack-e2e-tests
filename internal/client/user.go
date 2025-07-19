package client

import (
	"log/slog"
	"net/url"
	"strconv"
)

type UserClient struct {
	client *Client
}

func NewUserClient(client *Client) *UserClient {
	return &UserClient{
		client: client,
	}
}

func (uc *UserClient) CreateUser(req CreateUserRequest) (*CreateUserResponse, error) {
	uc.client.log.Info("Creating new user", slog.String("username", req.Username), slog.String("email", req.Email))

	var response CreateUserResponse
	err := uc.client.Post("/v1/users", req, &response)
	if err != nil {
		uc.client.log.Error("Failed to create user", slog.String("username", req.Username), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Info("User created successfully", slog.Int("user_id", response.ID), slog.String("username", response.Username))
	return &response, nil
}

func (uc *UserClient) GetUserByID(userID int) (*User, error) {
	uc.client.log.Debug("Getting user by ID", slog.Int("user_id", userID))

	var response User
	err := uc.client.Get("/v1/users/"+strconv.Itoa(userID), nil, &response)
	if err != nil {
		uc.client.log.Error("Failed to get user by ID", slog.Int("user_id", userID), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Debug("Got user by ID successfully", slog.Int("user_id", userID), slog.String("username", response.Username))
	return &response, nil
}

func (uc *UserClient) GetUserByUsername(username string) (*User, error) {
	uc.client.log.Debug("Getting user by username", slog.String("username", username))

	var response User
	err := uc.client.Get("/v1/users/username/"+url.PathEscape(username), nil, &response)
	if err != nil {
		uc.client.log.Error("Failed to get user by username", slog.String("username", username), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Debug("Got user by username successfully", slog.String("username", username), slog.Int("user_id", response.ID))
	return &response, nil
}

func (uc *UserClient) GetUserByEmail(email string) (*User, error) {
	uc.client.log.Debug("Getting user by email", slog.String("email", email))

	var response User
	err := uc.client.Get("/v1/users/email/"+url.PathEscape(email), nil, &response)
	if err != nil {
		uc.client.log.Error("Failed to get user by email", slog.String("email", email), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Debug("Got user by email successfully", slog.String("email", email), slog.Int("user_id", response.ID))
	return &response, nil
}

func (uc *UserClient) UpdateUser(req UpdateUserRequest) (*UpdateUserResponse, error) {
	uc.client.log.Info("Updating user", slog.Int("user_id", req.ID))

	var response UpdateUserResponse
	err := uc.client.Put("/v1/users/"+strconv.Itoa(req.ID), req, &response)
	if err != nil {
		uc.client.log.Error("Failed to update user", slog.Int("user_id", req.ID), slog.String("error", err.Error()))
		return nil, err
	}

	uc.client.log.Info("User updated successfully", slog.Int("user_id", req.ID))
	return &response, nil
}

func (uc *UserClient) UpdateAvatar(userID int, req UpdateAvatarRequest) error {
	uc.client.log.Info("Updating user avatar", slog.Int("user_id", userID))

	err := uc.client.Put("/v1/users/"+strconv.Itoa(userID)+"/avatar", req, nil)
	if err != nil {
		uc.client.log.Error("Failed to update avatar", slog.Int("user_id", userID), slog.String("error", err.Error()))
		return err
	}

	uc.client.log.Info("User avatar updated successfully", slog.Int("user_id", userID))
	return nil
}

func (uc *UserClient) DeleteUser(userID int) error {
	uc.client.log.Info("Deleting user", slog.Int("user_id", userID))

	err := uc.client.Delete("/v1/users/"+strconv.Itoa(userID), nil)
	if err != nil {
		uc.client.log.Error("Failed to delete user", slog.Int("user_id", userID), slog.String("error", err.Error()))
		return err
	}

	uc.client.log.Info("User deleted successfully", slog.Int("user_id", userID))
	return nil
}

func (uc *UserClient) SearchUsers(query string, page, limit int) (*SearchUsersResponse, error) {
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

	var response SearchUsersResponse
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
