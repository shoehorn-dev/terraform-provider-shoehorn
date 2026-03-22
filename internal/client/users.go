package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// DirectoryUser represents a user in the IdP directory.
type DirectoryUser struct {
	ID          string          `json:"id"`
	Username    string          `json:"username"`
	FirstName   string          `json:"firstName"`
	LastName    string          `json:"lastName"`
	Email       string          `json:"email"`
	Enabled     bool            `json:"enabled"`
	GitProvider string          `json:"git_provider,omitempty"`
	Provider    string          `json:"provider,omitempty"`
	Bundles     []BundleSummary `json:"bundles,omitempty"`
}

// BundleSummary is a slim bundle representation returned in directory user lists.
type BundleSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// directoryUserListResponse wraps the list users response.
type directoryUserListResponse struct {
	Items    []DirectoryUser `json:"items"`
	Provider string          `json:"provider,omitempty"`
}

// ListDirectoryUsers retrieves all IdP users.
func (c *Client) ListDirectoryUsers(ctx context.Context) ([]DirectoryUser, error) {
	body, err := c.Get(ctx, "/api/v1/users")
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	var resp directoryUserListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal users response: %w", err)
	}

	return resp.Items, nil
}

// GetDirectoryUser retrieves a single IdP user by ID.
func (c *Client) GetDirectoryUser(ctx context.Context, userID string) (*DirectoryUser, error) {
	path := fmt.Sprintf("/api/v1/users/%s", url.PathEscape(userID))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("get user %q: %w", userID, err)
	}

	var user DirectoryUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user response: %w", err)
	}

	return &user, nil
}
