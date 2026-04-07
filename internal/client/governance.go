package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GovernanceAction represents a governance action item in the Shoehorn platform.
type GovernanceAction struct {
	ID             string `json:"id"`
	EntityID       string `json:"entity_id"`
	EntityName     string `json:"entity_name"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Priority       string `json:"priority"`        // critical, high, medium, low
	Status         string `json:"status"`           // open, in_progress, resolved, dismissed, wont_fix
	SourceType     string `json:"source_type"`      // scorecard, security, policy
	SourceID       string `json:"source_id"`
	AssignedTo     string `json:"assigned_to,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	SLADays        *int   `json:"sla_days,omitempty"`
	ResolutionNote string `json:"resolution_note,omitempty"`
	CreatedBy      string `json:"created_by,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// CreateGovernanceActionRequest is the request body for creating a governance action.
type CreateGovernanceActionRequest struct {
	EntityID    string  `json:"entity_id"`
	EntityName  string  `json:"entity_name,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Priority    string  `json:"priority"`
	SourceType  string  `json:"source_type"`
	SourceID    string  `json:"source_id,omitempty"`
	AssignedTo  *string `json:"assigned_to,omitempty"`
	SLADays     *int    `json:"sla_days,omitempty"`
}

// UpdateGovernanceActionRequest is the request body for updating a governance action.
// Pointer fields allow sending only changed fields via PATCH.
type UpdateGovernanceActionRequest struct {
	Status         *string `json:"status,omitempty"`
	Priority       *string `json:"priority,omitempty"`
	AssignedTo     *string `json:"assigned_to,omitempty"`
	DueDate        *string `json:"due_date,omitempty"`
	ResolutionNote *string `json:"resolution_note,omitempty"`
}

// GovernanceActionFilters defines optional query parameters for listing governance actions.
type GovernanceActionFilters struct {
	EntityID   string
	Status     string
	Priority   string
	SourceType string
	Overdue    *bool
}

// governanceActionListResponse wraps the list governance actions API response.
type governanceActionListResponse struct {
	Actions []GovernanceAction     `json:"actions"`
	Total   int                    `json:"total"`
	Summary map[string]interface{} `json:"summary"`
}

// governanceActionCreateResponse wraps the create governance action API response.
type governanceActionCreateResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ListGovernanceActions retrieves governance actions with optional filters.
func (c *Client) ListGovernanceActions(ctx context.Context, filters *GovernanceActionFilters) ([]GovernanceAction, int, error) {
	path := "/api/v1/governance/actions"

	if filters != nil {
		params := url.Values{}
		if filters.EntityID != "" {
			params.Set("entity_id", filters.EntityID)
		}
		if filters.Status != "" {
			params.Set("status", filters.Status)
		}
		if filters.Priority != "" {
			params.Set("priority", filters.Priority)
		}
		if filters.SourceType != "" {
			params.Set("source_type", filters.SourceType)
		}
		if filters.Overdue != nil {
			if *filters.Overdue {
				params.Set("overdue", "true")
			} else {
				params.Set("overdue", "false")
			}
		}
		if encoded := params.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, 0, fmt.Errorf("list governance actions: %w", err)
	}

	var resp governanceActionListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, fmt.Errorf("unmarshal governance actions response: %w", err)
	}

	return resp.Actions, resp.Total, nil
}

// GetGovernanceAction retrieves a governance action by ID.
func (c *Client) GetGovernanceAction(ctx context.Context, id string) (*GovernanceAction, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/governance/actions/%s", url.PathEscape(id)))
	if err != nil {
		return nil, fmt.Errorf("get governance action %s: %w", id, err)
	}

	var action GovernanceAction
	if err := json.Unmarshal(body, &action); err != nil {
		return nil, fmt.Errorf("unmarshal governance action response: %w", err)
	}

	return &action, nil
}

// CreateGovernanceAction creates a new governance action and returns the full object.
// The API returns only the ID on creation, so a follow-up GET is performed to
// retrieve the complete action.
func (c *Client) CreateGovernanceAction(ctx context.Context, req CreateGovernanceActionRequest) (*GovernanceAction, error) {
	body, err := c.Post(ctx, "/api/v1/governance/actions", req)
	if err != nil {
		return nil, fmt.Errorf("create governance action: %w", err)
	}

	var createResp governanceActionCreateResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("unmarshal create governance action response: %w", err)
	}

	if createResp.ID == "" {
		return nil, fmt.Errorf("create governance action: response missing id")
	}

	// Fetch the full action since create only returns the ID
	return c.GetGovernanceAction(ctx, createResp.ID)
}

// UpdateGovernanceAction updates an existing governance action using PATCH.
// Only non-nil fields in the request are sent to the API.
func (c *Client) UpdateGovernanceAction(ctx context.Context, id string, req UpdateGovernanceActionRequest) error {
	_, err := c.Patch(ctx, fmt.Sprintf("/api/v1/governance/actions/%s", url.PathEscape(id)), req)
	if err != nil {
		return fmt.Errorf("update governance action %s: %w", id, err)
	}
	return nil
}

// DeleteGovernanceAction deletes a governance action by ID.
// The API returns 204 No Content on success.
func (c *Client) DeleteGovernanceAction(ctx context.Context, id string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/governance/actions/%s", url.PathEscape(id))); err != nil {
		return fmt.Errorf("delete governance action %s: %w", id, err)
	}
	return nil
}
