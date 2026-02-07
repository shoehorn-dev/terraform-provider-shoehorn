package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// PlatformPolicy represents a Shoehorn platform policy.
type PlatformPolicy struct {
	ID            string `json:"id"`
	Key           string `json:"key"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Category      string `json:"category,omitempty"`
	Enabled       bool   `json:"enabled"`
	Enforcement   string `json:"enforcement,omitempty"`
	AffectedUsers int    `json:"affected_users,omitempty"`
	System        bool   `json:"system,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// UpdatePolicyRequest is the request body for updating a platform policy.
type UpdatePolicyRequest struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Enforcement string `json:"enforcement,omitempty"`
}

// policyListResponse wraps the list response.
type policyListResponse struct {
	Policies []PlatformPolicy `json:"policies"`
}

// ListPolicies retrieves all platform policies.
func (c *Client) ListPolicies(ctx context.Context) ([]PlatformPolicy, error) {
	body, err := c.Get(ctx, "/api/v1/admin/policies")
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	var resp policyListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal policies response: %w", err)
	}

	return resp.Policies, nil
}

// GetPolicy retrieves a policy by key from the list.
func (c *Client) GetPolicy(ctx context.Context, key string) (*PlatformPolicy, error) {
	policies, err := c.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range policies {
		if p.Key == key {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("policy %q not found", key)
}

// UpdatePolicy updates a platform policy by ID.
func (c *Client) UpdatePolicy(ctx context.Context, id string, req UpdatePolicyRequest) (*PlatformPolicy, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/admin/policies/%s", id), req)
	if err != nil {
		return nil, fmt.Errorf("update policy %s: %w", id, err)
	}

	var policy PlatformPolicy
	if err := json.Unmarshal(body, &policy); err != nil {
		return nil, fmt.Errorf("unmarshal update policy response: %w", err)
	}

	return &policy, nil
}
