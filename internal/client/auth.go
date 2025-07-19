package client

import (
	"github.com/Soloda1/pinstack-e2e-tests/internal/custom_errors"
	"log/slog"
)

type AuthClient struct {
	client *Client
}

func NewAuthClient(client *Client) *AuthClient {
	return &AuthClient{
		client: client,
	}
}

func (ac *AuthClient) Register(req RegisterRequest) (*RegisterResponse, error) {
	ac.client.log.Info("Registering new user", slog.String("username", req.Username), slog.String("email", req.Email))

	var response RegisterResponse
	err := ac.client.Post("/v1/auth/register", req, &response)
	if err != nil {
		ac.client.log.Error("Failed to register user", slog.String("username", req.Username), slog.String("error", err.Error()))
		return nil, custom_errors.ErrRegistrationFailed
	}

	ac.client.SetToken(response.AccessToken)
	ac.client.log.Info("User registered successfully", slog.String("username", req.Username))

	return &response, nil
}

func (ac *AuthClient) Login(req LoginRequest) (*LoginResponse, error) {
	ac.client.log.Info("User login attempt", slog.String("login", req.Login))

	var response LoginResponse
	err := ac.client.Post("/v1/auth/login", req, &response)
	if err != nil {
		ac.client.log.Error("Login failed", slog.String("login", req.Login), slog.String("error", err.Error()))
		return nil, custom_errors.ErrLoginFailed
	}

	ac.client.SetToken(response.AccessToken)
	ac.client.log.Info("User logged in successfully", slog.String("login", req.Login))

	return &response, nil
}

func (ac *AuthClient) RefreshToken(req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	ac.client.log.Debug("Refreshing access token")

	var response RefreshTokenResponse
	err := ac.client.Post("/v1/auth/refresh", req, &response)
	if err != nil {
		ac.client.log.Error("Failed to refresh token", slog.String("error", err.Error()))
		return nil, custom_errors.ErrRefreshTokenFailed
	}

	ac.client.SetToken(response.AccessToken)
	ac.client.log.Debug("Token refreshed successfully")

	return &response, nil
}

func (ac *AuthClient) Logout(req LogoutRequest) error {
	ac.client.log.Info("User logout attempt")

	err := ac.client.Post("/v1/auth/logout", req, nil)
	if err != nil {
		ac.client.log.Error("Logout failed", slog.String("error", err.Error()))
		return custom_errors.ErrLogoutFailed
	}

	ac.client.SetToken("")
	ac.client.log.Info("User logged out successfully")

	return nil
}

func (ac *AuthClient) UpdatePassword(req UpdatePasswordRequest) (*UpdatePasswordResponse, error) {
	ac.client.log.Info("Updating user password")

	var response UpdatePasswordResponse
	err := ac.client.Post("/v1/auth/update-password", req, &response)
	if err != nil {
		ac.client.log.Error("Failed to update password", slog.String("error", err.Error()))
		return nil, custom_errors.ErrPasswordUpdateFailed
	}

	ac.client.log.Info("Password updated successfully")
	return &response, nil
}
