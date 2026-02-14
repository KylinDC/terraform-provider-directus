package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Directus API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// Config holds the configuration for creating a new client
type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

// ErrorResponse represents an error response from the Directus v11 API.
// Directus v11 returns errors in the format: {"errors": [{"message": "...", "extensions": {"code": "..."}}]}
type ErrorResponse struct {
	Errors []DirectusError `json:"errors"`
}

// DirectusError represents a single error entry in the Directus error response.
type DirectusError struct {
	Message    string `json:"message"`
	Extensions struct {
		Code string `json:"code"`
	} `json:"extensions"`
}

// NewClient creates a new Directus API client
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(config.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	if config.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	client := &Client{
		BaseURL: config.BaseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Token: config.Token,
	}

	return client, nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	fullURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, c.handleErrorResponse(resp)
	}

	return resp, nil
}

// handleErrorResponse parses and returns an error from the Directus v11 API response.
// Directus v11 format: {"errors": [{"message": "...", "extensions": {"code": "..."}}]}
func (c *Client) handleErrorResponse(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response: %w", resp.StatusCode, err)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if len(errorResp.Errors) > 0 && errorResp.Errors[0].Message != "" {
		firstErr := errorResp.Errors[0]
		if firstErr.Extensions.Code != "" {
			return fmt.Errorf("HTTP %d [%s]: %s", resp.StatusCode, firstErr.Extensions.Code, firstErr.Message)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, firstErr.Message)
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
}

// Get retrieves a single item from a collection
func (c *Client) Get(ctx context.Context, collection, id string, result interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}
	if id == "" {
		return fmt.Errorf("id is required")
	}

	path := c.buildCollectionPath(collection, id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// GetWithParams retrieves a single item from a collection with query parameters.
// This is useful for requesting expanded relational fields (e.g., ?fields=*,policies.id,policies.policy).
func (c *Client) GetWithParams(ctx context.Context, collection, id string, params map[string]string, result interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}
	if id == "" {
		return fmt.Errorf("id is required")
	}

	path := c.buildCollectionPath(collection, id)
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		path += "?" + q.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// buildCollectionPath builds the correct API path for a collection
// System collections (roles, policies, users, etc.) use /{collection} format
// Custom collections use /items/{collection} format
func (c *Client) buildCollectionPath(collection string, id string) string {
	systemCollections := map[string]bool{
		"collections": true,
		"roles":       true,
		"policies":    true,
		"users":       true,
		"folders":     true,
		"files":       true,
		"activity":    true,
		"revisions":   true,
		"webhooks":    true,
		"flows":       true,
		"operations":  true,
		"dashboards":  true,
		"panels":      true,
		"shares":      true,
		"settings":    true,
	}

	if systemCollections[collection] {
		if id != "" {
			return fmt.Sprintf("/%s/%s", collection, id)
		}
		return fmt.Sprintf("/%s", collection)
	}

	if id != "" {
		return fmt.Sprintf("/items/%s/%s", collection, id)
	}
	return fmt.Sprintf("/items/%s", collection)
}

// List retrieves multiple items from a collection
func (c *Client) List(ctx context.Context, collection string, result interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}

	path := c.buildCollectionPath(collection, "")
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Create creates a new item in a collection
func (c *Client) Create(ctx context.Context, collection string, data interface{}, result interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}
	if data == nil {
		return fmt.Errorf("data is required")
	}

	path := c.buildCollectionPath(collection, "")
	resp, err := c.doRequest(ctx, http.MethodPost, path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Update updates an existing item in a collection
func (c *Client) Update(ctx context.Context, collection, id string, data interface{}, result interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if data == nil {
		return fmt.Errorf("data is required")
	}

	path := c.buildCollectionPath(collection, id)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Delete deletes an item from a collection
func (c *Client) Delete(ctx context.Context, collection, id string) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}
	if id == "" {
		return fmt.Errorf("id is required")
	}

	path := c.buildCollectionPath(collection, id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Ping checks if the Directus server is reachable
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/server/ping", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read ping response: %w", err)
	}

	if string(body) != "pong" {
		return fmt.Errorf("unexpected ping response: %s", string(body))
	}

	return nil
}
