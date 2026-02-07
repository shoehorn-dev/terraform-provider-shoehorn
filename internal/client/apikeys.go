package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// APIKey represents a Shoehorn API key.
type APIKey struct {
	ID         string   `json:"id"`
	TenantID   string   `json:"tenant_id,omitempty"`
	Name       string   `json:"name"`
	Description string  `json:"description,omitempty"`
	KeyPrefix  string   `json:"key_prefix"`
	Scopes     []string `json:"scopes"`
	LastUsedAt string   `json:"last_used_at,omitempty"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
	RevokedAt  string   `json:"revoked_at,omitempty"`
	CreatedBy  string   `json:"created_by,omitempty"`
	CreatedAt  string   `json:"created_at,omitempty"`
	UpdatedAt  string   `json:"updated_at,omitempty"`
}

// CreateAPIKeyRequest is the request body for creating an API key.
type CreateAPIKeyRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	Scopes        []string `json:"scopes"`
	ExpiresInDays *int     `json:"expires_in_days,omitempty"`
}

// CreateAPIKeyResponse is the response from creating an API key.
// The raw key is only returned once on creation.
type CreateAPIKeyResponse struct {
	Key    APIKey `json:"key"`
	RawKey string `json:"raw_key"`
}

// apiKeyListResponse wraps the list response.
type apiKeyListResponse struct {
	Keys  []APIKey `json:"keys"`
	Total int      `json:"total"`
}

// ListAPIKeys retrieves all API keys.
func (c *Client) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	body, err := c.Get(ctx, "/api/v1/admin/api-keys")
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	var resp apiKeyListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal api keys response: %w", err)
	}

	return resp.Keys, nil
}

// GetAPIKey retrieves an API key by ID from the list.
func (c *Client) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	keys, err := c.ListAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.ID == id {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("api key %q not found", id)
}

// CreateAPIKey generates a new API key.
func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	body, err := c.Post(ctx, "/api/v1/admin/api-keys", req)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	var resp CreateAPIKeyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create api key response: %w", err)
	}

	return &resp, nil
}

// RevokeAPIKey revokes an API key by ID.
func (c *Client) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/v1/admin/api-keys/%s/revoke", id), nil)
	if err != nil {
		return fmt.Errorf("revoke api key %s: %w", id, err)
	}
	return nil
}
