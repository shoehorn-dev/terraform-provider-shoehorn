package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// FeatureFlag represents a Shoehorn feature flag.
type FeatureFlag struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	DefaultEnabled bool   `json:"default_enabled"`
	OverrideCount  int    `json:"override_count,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// CreateFeatureFlagRequest is the request body for creating a feature flag.
type CreateFeatureFlagRequest struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	DefaultEnabled bool   `json:"default_enabled"`
}

// UpdateFeatureFlagRequest is the request body for updating a feature flag.
type UpdateFeatureFlagRequest struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	DefaultEnabled *bool  `json:"default_enabled,omitempty"`
}

// featureFlagListResponse wraps the list response.
type featureFlagListResponse struct {
	Flags []FeatureFlag `json:"flags"`
}

// GetFeatureFlag retrieves a feature flag by key.
// Uses the list endpoint and filters by key since there's no single-get endpoint.
func (c *Client) GetFeatureFlag(ctx context.Context, key string) (*FeatureFlag, error) {
	body, err := c.Get(ctx, "/api/v1/admin/features")
	if err != nil {
		return nil, fmt.Errorf("get feature flags: %w", err)
	}

	var resp featureFlagListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal feature flags response: %w", err)
	}

	for _, flag := range resp.Flags {
		if flag.Key == key {
			return &flag, nil
		}
	}

	return nil, fmt.Errorf("feature flag %q not found", key)
}

// ListFeatureFlags retrieves all feature flags.
func (c *Client) ListFeatureFlags(ctx context.Context) ([]FeatureFlag, error) {
	body, err := c.Get(ctx, "/api/v1/admin/features")
	if err != nil {
		return nil, fmt.Errorf("list feature flags: %w", err)
	}

	var resp featureFlagListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal feature flags response: %w", err)
	}

	return resp.Flags, nil
}

// CreateFeatureFlag creates a new feature flag.
func (c *Client) CreateFeatureFlag(ctx context.Context, req CreateFeatureFlagRequest) (*FeatureFlag, error) {
	body, err := c.Post(ctx, "/api/v1/admin/features", req)
	if err != nil {
		return nil, fmt.Errorf("create feature flag: %w", err)
	}

	var flag FeatureFlag
	if err := json.Unmarshal(body, &flag); err != nil {
		return nil, fmt.Errorf("unmarshal create feature flag response: %w", err)
	}

	return &flag, nil
}

// UpdateFeatureFlag updates a feature flag by key.
func (c *Client) UpdateFeatureFlag(ctx context.Context, key string, req UpdateFeatureFlagRequest) (*FeatureFlag, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/admin/features/%s", key), req)
	if err != nil {
		return nil, fmt.Errorf("update feature flag %s: %w", key, err)
	}

	var flag FeatureFlag
	if err := json.Unmarshal(body, &flag); err != nil {
		return nil, fmt.Errorf("unmarshal update feature flag response: %w", err)
	}

	return &flag, nil
}

// DeleteFeatureFlag deletes a feature flag by key.
func (c *Client) DeleteFeatureFlag(ctx context.Context, key string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/admin/features/%s", key)); err != nil {
		return fmt.Errorf("delete feature flag %s: %w", key, err)
	}
	return nil
}
