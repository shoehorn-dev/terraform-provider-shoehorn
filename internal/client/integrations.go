package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Integration represents a Shoehorn external integration.
type Integration struct {
	ID         int                    `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	TeamID     string                 `json:"team_id,omitempty"`
	CreatedBy  string                 `json:"created_by,omitempty"`
	CreatedAt  string                 `json:"created_at,omitempty"`
	UpdatedAt  string                 `json:"updated_at,omitempty"`
	LastSyncAt string                 `json:"last_sync_at,omitempty"`
	LastError  string                 `json:"last_error,omitempty"`
}

// CreateIntegrationRequest is the request body for creating an integration.
type CreateIntegrationRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
	TeamID string                 `json:"team_id,omitempty"`
}

// UpdateIntegrationRequest is the request body for updating an integration.
type UpdateIntegrationRequest struct {
	Name   string                 `json:"name,omitempty"`
	Status string                 `json:"status,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// integrationResponse wraps a single integration response.
type integrationResponse struct {
	Integration Integration `json:"integration"`
}

// integrationListResponse wraps the list response.
type integrationListResponse struct {
	Integrations []Integration `json:"integrations"`
	Count        int           `json:"count"`
}

// IntegrationStatus represents a system integration status from /api/v1/integrations.
type IntegrationStatus struct {
	Type     string                 `json:"type"`
	Provider string                 `json:"provider"`
	Status   string                 `json:"status"`
	Config   map[string]interface{} `json:"config,omitempty"`
	LastSync string                 `json:"last_sync,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// integrationStatusResponse wraps the system integrations status response.
type integrationStatusResponse struct {
	Integrations []IntegrationStatus `json:"integrations"`
	Total        int                 `json:"total"`
	Healthy      int                 `json:"healthy"`
	LastUpdated  string              `json:"last_updated"`
}

// GetIntegrationsStatus retrieves system integration statuses from /api/v1/integrations.
func (c *Client) GetIntegrationsStatus(ctx context.Context) ([]IntegrationStatus, int, int, error) {
	body, err := c.Get(ctx, "/api/v1/integrations")
	if err != nil {
		return nil, 0, 0, fmt.Errorf("get integrations status: %w", err)
	}

	var resp integrationStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, 0, fmt.Errorf("unmarshal integrations status response: %w", err)
	}

	return resp.Integrations, resp.Total, resp.Healthy, nil
}

// ListIntegrations retrieves all user-created integration configs.
func (c *Client) ListIntegrations(ctx context.Context) ([]Integration, error) {
	body, err := c.Get(ctx, "/api/v1/integrations/configs")
	if err != nil {
		return nil, fmt.Errorf("list integrations: %w", err)
	}

	var resp integrationListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal integrations response: %w", err)
	}

	return resp.Integrations, nil
}

// GetIntegration retrieves an integration by ID.
func (c *Client) GetIntegration(ctx context.Context, id int) (*Integration, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/integrations/%d", id))
	if err != nil {
		return nil, fmt.Errorf("get integration %d: %w", id, err)
	}

	var resp integrationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal integration response: %w", err)
	}

	return &resp.Integration, nil
}

// CreateIntegration creates a new integration.
func (c *Client) CreateIntegration(ctx context.Context, req CreateIntegrationRequest) (*Integration, error) {
	body, err := c.Post(ctx, "/api/v1/integrations", req)
	if err != nil {
		return nil, fmt.Errorf("create integration: %w", err)
	}

	var resp integrationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create integration response: %w", err)
	}

	return &resp.Integration, nil
}

// UpdateIntegration updates an existing integration.
func (c *Client) UpdateIntegration(ctx context.Context, id int, req UpdateIntegrationRequest) (*Integration, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/integrations/%d", id), req)
	if err != nil {
		return nil, fmt.Errorf("update integration %d: %w", id, err)
	}

	var resp integrationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal update integration response: %w", err)
	}

	return &resp.Integration, nil
}

// DeleteIntegration deletes an integration by ID.
func (c *Client) DeleteIntegration(ctx context.Context, id int) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/integrations/%d", id)); err != nil {
		return fmt.Errorf("delete integration %d: %w", id, err)
	}
	return nil
}
