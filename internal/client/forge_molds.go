package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ForgeMold represents a Shoehorn Forge mold template.
type ForgeMold struct {
	ID          string                 `json:"id"`
	Slug        string                 `json:"slug"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Visibility  string                 `json:"visibility"`
	Tags        []string               `json:"tags,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Category    string                 `json:"category"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
	Actions     []ForgeMoldAction      `json:"actions"`
	Published   bool                   `json:"published,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAt   string                 `json:"updated_at,omitempty"`
}

// ForgeMoldAction represents an action that can be performed on a forge mold.
type ForgeMoldAction struct {
	Action      string `json:"action"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
}

// CreateForgeMoldRequest is the request body for creating a forge mold.
type CreateForgeMoldRequest struct {
	Slug        string                 `json:"slug"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Visibility  string                 `json:"visibility"`
	Tags        []string               `json:"tags,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Category    string                 `json:"category"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
	Actions     []ForgeMoldAction      `json:"actions"`
}

// UpdateForgeMoldRequest is the request body for updating a forge mold.
type UpdateForgeMoldRequest struct {
	Version     string                 `json:"version"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
}

// forgeMoldResponse wraps a single mold response.
type forgeMoldResponse struct {
	Mold ForgeMold `json:"mold"`
}

// forgeMoldListResponse wraps the list response.
type forgeMoldListResponse struct {
	Molds []ForgeMold `json:"molds"`
}

// publishForgeMoldRequest is the request body for publishing a forge mold.
type publishForgeMoldRequest struct {
	Version string `json:"version"`
}

// ListForgeMolds retrieves all forge molds.
func (c *Client) ListForgeMolds(ctx context.Context) ([]ForgeMold, error) {
	body, err := c.Get(ctx, "/api/v1/forge/molds")
	if err != nil {
		return nil, fmt.Errorf("list forge molds: %w", err)
	}

	var resp forgeMoldListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal forge molds response: %w", err)
	}

	return resp.Molds, nil
}

// GetForgeMold retrieves a forge mold by slug.
func (c *Client) GetForgeMold(ctx context.Context, slug string) (*ForgeMold, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/forge/molds/%s", url.PathEscape(slug)))
	if err != nil {
		return nil, fmt.Errorf("get forge mold %s: %w", slug, err)
	}

	var resp forgeMoldResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal forge mold response: %w", err)
	}

	return &resp.Mold, nil
}

// CreateForgeMold creates a new forge mold.
func (c *Client) CreateForgeMold(ctx context.Context, req CreateForgeMoldRequest) (*ForgeMold, error) {
	body, err := c.Post(ctx, "/api/v1/forge/molds", req)
	if err != nil {
		return nil, fmt.Errorf("create forge mold: %w", err)
	}

	var resp forgeMoldResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create forge mold response: %w", err)
	}

	return &resp.Mold, nil
}

// UpdateForgeMold updates an existing forge mold by slug.
func (c *Client) UpdateForgeMold(ctx context.Context, slug string, req UpdateForgeMoldRequest) (*ForgeMold, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/forge/molds/%s", url.PathEscape(slug)), req)
	if err != nil {
		return nil, fmt.Errorf("update forge mold %s: %w", slug, err)
	}

	var resp forgeMoldResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal update forge mold response: %w", err)
	}

	return &resp.Mold, nil
}

// DeleteForgeMold deletes a forge mold by slug. The version parameter is required
// by the API and is sent as a query parameter.
func (c *Client) DeleteForgeMold(ctx context.Context, slug, version string) error {
	params := url.Values{}
	params.Set("version", version)
	path := fmt.Sprintf("/api/v1/forge/molds/%s", url.PathEscape(slug)) + "?" + params.Encode()
	if err := c.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete forge mold %s: %w", slug, err)
	}
	return nil
}

// PublishForgeMold publishes a forge mold by slug with the given version.
func (c *Client) PublishForgeMold(ctx context.Context, slug, version string) (*ForgeMold, error) {
	body, err := c.Post(ctx, fmt.Sprintf("/api/v1/forge/molds/%s/publish", url.PathEscape(slug)), publishForgeMoldRequest{
		Version: version,
	})
	if err != nil {
		return nil, fmt.Errorf("publish forge mold %s: %w", slug, err)
	}

	var resp forgeMoldResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal publish forge mold response: %w", err)
	}

	return &resp.Mold, nil
}
