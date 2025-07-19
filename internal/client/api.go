// filepath: /home/solo/GolandProjects/pinstack-e2e-tests/internal/client/api.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Soloda1/pinstack-e2e-tests/config"
	"github.com/Soloda1/pinstack-e2e-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-e2e-tests/internal/logger"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
	log        *logger.Logger
}

func NewClient(cfg *config.Config, log *logger.Logger) *Client {
	return &Client{
		BaseURL: cfg.API.BaseURL,
		HTTPClient: &http.Client{
			Timeout: cfg.API.Timeout,
		},
		log: log,
	}
}

func (c *Client) SetToken(token string) {
	c.Token = token
}

func (c *Client) makeRequest(method, path string, queryParams url.Values, body interface{}, result interface{}) error {
	reqURL, err := url.Parse(c.BaseURL + path)
	if err != nil {
		c.log.Error("Invalid URL", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrInvalidURL
	}

	if queryParams != nil {
		reqURL.RawQuery = queryParams.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			c.log.Error("failed to marshal request body", slog.String("path", path), slog.String("error", err.Error()))
			return custom_errors.ErrJSONMarshalFailed
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, reqURL.String(), reqBody)
	if err != nil {
		c.log.Error("failed to create request", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrRequestCreationFailed
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.log.Error("failed to execute request", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrRequestFailed
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("failed to read response", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrResponseReadFailed
	}

	if resp.StatusCode >= 400 {
		var errorResp ErrorBody
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			c.log.Debug("api error failed to unmarshall", slog.String("status code", resp.Status), slog.String("body", string(respBody)))
			return custom_errors.ErrJSONUnmarshalFailed
		}
		c.log.Error("api error", slog.String("status code", resp.Status), slog.Any("errors", errorResp))
		return fmt.Errorf("%s", errorResp.Message)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			c.log.Error("failed to unmarshal response", slog.String("path", path), slog.String("error", err.Error()))
			return custom_errors.ErrJSONUnmarshalFailed
		}
	}

	return nil
}

func (c *Client) Get(path string, queryParams url.Values, result interface{}) error {
	return c.makeRequest(http.MethodGet, path, queryParams, nil, result)
}

func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.makeRequest(http.MethodPost, path, nil, body, result)
}

func (c *Client) Put(path string, body interface{}, result interface{}) error {
	return c.makeRequest(http.MethodPut, path, nil, body, result)
}

func (c *Client) Delete(path string, result interface{}) error {
	return c.makeRequest(http.MethodDelete, path, nil, nil, result)
}
