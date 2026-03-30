package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GitOpsResource represents a GitOps resource (Application, Kustomization, HelmRelease, etc.)
// pushed by a K8s agent. These are read-only in the API.
type GitOpsResource struct {
	ID             string `json:"id"`
	ClusterID      string `json:"cluster_id"`
	Tool           string `json:"tool"`
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	SyncStatus     string `json:"sync_status"`
	HealthStatus   string `json:"health_status"`
	SourceURL      string `json:"source_url,omitempty"`
	Revision       string `json:"revision,omitempty"`
	TargetRevision string `json:"target_revision,omitempty"`
	AutoSync       bool   `json:"auto_sync"`
	Suspended      bool   `json:"suspended"`
	ChartName      string `json:"chart_name,omitempty"`
	ChartVersion   string `json:"chart_version,omitempty"`
	EntityID       string `json:"entity_id,omitempty"`
	EntityName     string `json:"entity_name,omitempty"`
	OwnerTeam      string `json:"owner_team,omitempty"`
	LastSyncedAt   string `json:"last_synced_at,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// GitOpsStats represents aggregate statistics for GitOps resources.
type GitOpsStats struct {
	Total     int `json:"total"`
	Synced    int `json:"synced"`
	OutOfSync int `json:"out_of_sync"`
	Failed    int `json:"failed"`
	Suspended int `json:"suspended"`
	Unknown   int `json:"unknown"`
}

// ListGitOpsResourcesParams contains optional filter parameters for listing GitOps resources.
type ListGitOpsResourcesParams struct {
	ClusterID    string
	Tool         string
	SyncStatus   string
	HealthStatus string
}

// gitOpsListResponse wraps the list response from /api/v1/operations/gitops.
type gitOpsListResponse struct {
	Resources []GitOpsResource `json:"resources"`
	Total     int              `json:"total"`
}

// gitOpsResourceResponse wraps a single resource response.
type gitOpsResourceResponse struct {
	Resource GitOpsResource `json:"resource"`
}

// ListGitOpsResources retrieves GitOps resources with optional filters.
func (c *Client) ListGitOpsResources(ctx context.Context, params ListGitOpsResourcesParams) ([]GitOpsResource, int, error) {
	path := "/api/v1/operations/gitops"

	q := url.Values{}
	if params.ClusterID != "" {
		q.Set("cluster_id", params.ClusterID)
	}
	if params.Tool != "" {
		q.Set("tool", params.Tool)
	}
	if params.SyncStatus != "" {
		q.Set("sync_status", params.SyncStatus)
	}
	if params.HealthStatus != "" {
		q.Set("health_status", params.HealthStatus)
	}
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, 0, fmt.Errorf("list gitops resources: %w", err)
	}

	var resp gitOpsListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, fmt.Errorf("unmarshal gitops resources response: %w", err)
	}

	return resp.Resources, resp.Total, nil
}

// GetGitOpsResource retrieves a single GitOps resource by ID.
func (c *Client) GetGitOpsResource(ctx context.Context, id string) (*GitOpsResource, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/operations/gitops/%s", url.PathEscape(id)))
	if err != nil {
		return nil, fmt.Errorf("get gitops resource %s: %w", id, err)
	}

	// Try wrapped response first ({"resource": {...}})
	var wrapped gitOpsResourceResponse
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("unmarshal gitops resource response: %w", err)
	}

	// If wrapped response had data, return it
	if wrapped.Resource.ID != "" {
		return &wrapped.Resource, nil
	}

	// Otherwise try direct response (resource at top level)
	var resource GitOpsResource
	if err := json.Unmarshal(body, &resource); err != nil {
		return nil, fmt.Errorf("unmarshal gitops resource response: %w", err)
	}

	return &resource, nil
}

// GetGitOpsStats retrieves aggregate GitOps statistics.
func (c *Client) GetGitOpsStats(ctx context.Context) (*GitOpsStats, error) {
	body, err := c.Get(ctx, "/api/v1/operations/gitops/stats")
	if err != nil {
		return nil, fmt.Errorf("get gitops stats: %w", err)
	}

	var stats GitOpsStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal gitops stats response: %w", err)
	}

	return &stats, nil
}
