package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Soloda1/pinstack-e2e-tests/config"
	"github.com/Soloda1/pinstack-e2e-tests/internal/custom_errors"
	"github.com/Soloda1/pinstack-e2e-tests/internal/fixtures"
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
			c.log.Error("Failed to marshal request body", slog.String("path", path), slog.String("error", err.Error()))
			return custom_errors.ErrJSONMarshalFailed
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, reqURL.String(), reqBody)
	if err != nil {
		c.log.Error("Failed to create request", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrRequestCreationFailed
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.log.Error("Failed to execute request", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrRequestFailed
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			c.log.Error("Failed to close response Body", slog.String("path", path), slog.String("error", err.Error()))
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("Failed to read response", slog.String("path", path), slog.String("error", err.Error()))
		return custom_errors.ErrResponseReadFailed
	}

	c.log.Debug("response body", slog.Any("body", string(respBody)))

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		var errorResp fixtures.ErrorBody
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			c.log.Debug("API error failed to unmarshal", slog.String("status code", resp.Status), slog.String("body", string(respBody)))
			return custom_errors.ErrJSONUnmarshalFailed
		}
		c.log.Error("API error", slog.String("status code", resp.Status), slog.Any("errors", errorResp))
		return fmt.Errorf("%s", errorResp.Message)
	}

	if result != nil {
		var baseResp fixtures.BaseResponse
		if err := json.Unmarshal(respBody, &baseResp); err != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return custom_errors.ErrJSONUnmarshalFailed
			}
		} else {
			if baseResp.Data != nil {
				dataJSON, err := json.Marshal(baseResp.Data)
				if err != nil {
					return custom_errors.ErrJSONUnmarshalFailed
				}
				if err := json.Unmarshal(dataJSON, result); err != nil {
					return custom_errors.ErrJSONUnmarshalFailed
				}
			}
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
