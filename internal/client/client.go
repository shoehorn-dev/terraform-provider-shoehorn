package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const maxRetries = 3

// Client is the HTTP client for the Shoehorn API. It handles authentication,
// JSON serialization, and automatic retries on transient errors.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

// APIError represents an error response from the Shoehorn API. It captures the
// HTTP status code and any code/message fields from the JSON error body.
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// Error returns a human-readable description of the API error including the HTTP status code.
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("shoehorn API error (HTTP %d): %s - %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("shoehorn API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// NewClient creates a new Shoehorn API client with the given base URL, API key,
// and HTTP timeout. The base URL is normalized by stripping any trailing slash.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		UserAgent: "terraform-provider-shoehorn",
	}
}

// isRetryable returns true if the error is a transient connection error worth retrying.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "Client.Timeout")
}

// doRequest executes an HTTP request with authentication and returns the response body,
// status code, and any error. It retries up to maxRetries times on transient connection
// errors and 5xx server errors, using linear backoff between attempts.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	url := c.BaseURL + path

	var jsonData []byte
	if body != nil {
		var err error
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
	}

	tflog.Trace(ctx, "API request", map[string]any{"method": method, "path": path})

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			tflog.Warn(ctx, "retrying API request", map[string]any{
				"method":  method,
				"path":    path,
				"attempt": attempt + 1,
				"error":   lastErr.Error(),
			})
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}

		var reqBody io.Reader
		if jsonData != nil {
			reqBody = bytes.NewBuffer(jsonData)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, 0, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			if isRetryable(err) {
				continue
			}
			return nil, 0, lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			if isRetryable(err) {
				continue
			}
			return nil, resp.StatusCode, lastErr
		}

		tflog.Trace(ctx, "API response", map[string]any{"method": method, "path": path, "status": resp.StatusCode})

		if resp.StatusCode >= 500 {
			lastErr = &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
			continue
		}

		if resp.StatusCode >= 400 {
			apiErr := &APIError{StatusCode: resp.StatusCode}
			if err := json.Unmarshal(respBody, apiErr); err != nil {
				apiErr.Message = string(respBody)
			}
			// If standard code/message fields are empty, use the raw body
			// (catches validation error responses with "errors" array format)
			if apiErr.Message == "" && apiErr.Code == "" {
				body := string(respBody)
				if body != "" {
					apiErr.Message = body
				} else {
					apiErr.Message = http.StatusText(resp.StatusCode)
				}
			}
			return nil, resp.StatusCode, apiErr
		}

		return respBody, resp.StatusCode, nil
	}

	tflog.Error(ctx, "API request failed after all retries", map[string]any{
		"method":      method,
		"path":        path,
		"max_retries": maxRetries,
		"error":       lastErr.Error(),
	})
	return nil, 0, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// Get performs an authenticated GET request to the given API path and returns the response body.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, path, nil)
	return body, err
}

// Post performs an authenticated POST request to the given API path with a JSON-encoded body.
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	respBody, _, err := c.doRequest(ctx, http.MethodPost, path, body)
	return respBody, err
}

// Put performs an authenticated PUT request to the given API path with a JSON-encoded body.
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	respBody, _, err := c.doRequest(ctx, http.MethodPut, path, body)
	return respBody, err
}

// Delete performs an authenticated DELETE request to the given API path.
func (c *Client) Delete(ctx context.Context, path string) error {
	_, _, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}
