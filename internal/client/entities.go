package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Entity represents a Shoehorn catalog entity from the GET response.
type Entity struct {
	Service        EntityService  `json:"service"`
	Description    string         `json:"description,omitempty"`
	Owner          []OwnerInfo    `json:"owner,omitempty"`
	Lifecycle      string         `json:"lifecycle,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	Links          []LinkInfo     `json:"links,omitempty"`
	Relations      []RelationInfo `json:"relations,omitempty"`
	Integrations   *Integrations           `json:"integrations,omitempty"`
	Interfaces     map[string]interface{} `json:"interfaces,omitempty"`
	RepositoryPath string                 `json:"repositoryPath,omitempty"`
	CreatedAt      string         `json:"createdAt,omitempty"`
	UpdatedAt      string         `json:"updatedAt,omitempty"`
}

// Integrations represents the integrations block in an entity response.
type Integrations struct {
	Changelog *ChangelogIntegration `json:"changelog,omitempty"`
	Licenses  []LicenseInfo         `json:"licenses,omitempty"`
}

// ChangelogIntegration represents the changelog integration.
type ChangelogIntegration struct {
	Path string `json:"path,omitempty"`
}

// LicenseInfo represents a license entry.
type LicenseInfo struct {
	Title     string `json:"title"`
	Vendor    string `json:"vendor,omitempty"`
	Purchased string `json:"purchased,omitempty"`
	Expires   string `json:"expires,omitempty"`
	Seats     int    `json:"seats,omitempty"`
	Cost      string `json:"cost,omitempty"`
	Contract  string `json:"contract,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// EntityService is the service block within an entity response.
type EntityService struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Tier string `json:"tier,omitempty"`
}

// OwnerInfo represents an owner entry.
type OwnerInfo struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// LinkInfo represents a link entry.
type LinkInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon,omitempty"`
}

// RelationInfo represents a relationship between entities as returned by the API.
type RelationInfo struct {
	Type       string `json:"type"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Via        string `json:"via,omitempty"`
}

// ManifestEntityResponse is the response from create/update manifest endpoints.
type ManifestEntityResponse struct {
	Success bool `json:"success"`
	Entity  struct {
		ID          int    `json:"id"`
		ServiceID   string `json:"serviceId"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		Lifecycle   string `json:"lifecycle"`
		Description string `json:"description"`
		Source      string `json:"source"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	} `json:"entity"`
}

// CreateEntityRequest is the request body for creating an entity via manifest.
type CreateEntityRequest struct {
	Content string `json:"content"`
	Source  string `json:"source"`
}

// entityGetResponse wraps the GET entity response.
type entityGetResponse struct {
	Entity Entity `json:"entity"`
}

// EntityListItem represents an entity in the list response.
type EntityListItem struct {
	Service     EntityService `json:"service"`
	Description string        `json:"description,omitempty"`
	Owner       []OwnerInfo   `json:"owner,omitempty"`
	Lifecycle   string        `json:"lifecycle,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
	CreatedAt   string        `json:"createdAt,omitempty"`
	UpdatedAt   string        `json:"updatedAt,omitempty"`
}

// entitiesListResponse wraps the list entities API response.
type entitiesListResponse struct {
	Entities []EntityListItem `json:"entities"`
	Page     struct {
		Total      int     `json:"total"`
		Limit      int     `json:"limit"`
		NextCursor *string `json:"nextCursor"`
	} `json:"page"`
}

// ListEntities retrieves all entities using cursor-based pagination.
func (c *Client) ListEntities(ctx context.Context) ([]EntityListItem, error) {
	var all []EntityListItem
	cursor := ""

	for {
		url := "/api/v1/entities?limit=100"
		if cursor != "" {
			url += "&cursor=" + cursor
		}

		body, err := c.Get(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("list entities: %w", err)
		}

		var resp entitiesListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("unmarshal entities list response: %w", err)
		}

		all = append(all, resp.Entities...)

		if resp.Page.NextCursor == nil || *resp.Page.NextCursor == "" {
			break
		}
		cursor = *resp.Page.NextCursor
	}

	return all, nil
}

// GetEntity retrieves an entity by service ID.
func (c *Client) GetEntity(ctx context.Context, serviceID string) (*Entity, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/entities/%s", serviceID))
	if err != nil {
		return nil, fmt.Errorf("get entity %s: %w", serviceID, err)
	}

	var resp entityGetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal entity response: %w", err)
	}

	return &resp.Entity, nil
}

// CreateEntity creates a new entity via manifest upload.
func (c *Client) CreateEntity(ctx context.Context, req CreateEntityRequest) (*ManifestEntityResponse, error) {
	body, err := c.Post(ctx, "/api/v1/manifests/entities", req)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}

	var resp ManifestEntityResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create entity response: %w", err)
	}

	return &resp, nil
}

// UpdateEntity updates an entity via manifest upload.
func (c *Client) UpdateEntity(ctx context.Context, serviceID string, req CreateEntityRequest) (*ManifestEntityResponse, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/manifests/entities/%s", serviceID), req)
	if err != nil {
		return nil, fmt.Errorf("update entity %s: %w", serviceID, err)
	}

	var resp ManifestEntityResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal update entity response: %w", err)
	}

	return &resp, nil
}

// DeleteEntity deletes an entity by service ID.
// NOTE: This endpoint is currently blocked per API audit (missing DELETE endpoint).
// Once implemented in the API, this will call DELETE /api/v1/manifests/entities/{serviceId}.
func (c *Client) DeleteEntity(ctx context.Context, serviceID string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/manifests/entities/%s", serviceID)); err != nil {
		return fmt.Errorf("delete entity %s: %w", serviceID, err)
	}
	return nil
}
