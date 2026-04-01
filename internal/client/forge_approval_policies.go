package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ForgeApprovalPolicy represents a Shoehorn Forge approval policy.
type ForgeApprovalPolicy struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	Enabled          bool           `json:"enabled"`
	Priority         int            `json:"priority,omitempty"`
	ApprovalChain    []ApprovalStep `json:"approval_chain,omitempty"`
	AutoApproveAfter int            `json:"auto_approve_after,omitempty"`
	CreatedAt        string         `json:"created_at,omitempty"`
	UpdatedAt        string         `json:"updated_at,omitempty"`
}

// ApprovalStep represents a single step in an approval policy workflow.
type ApprovalStep struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	Approvers     []string `json:"approvers"`
	RequiredCount int      `json:"required_count,omitempty"`
	Timeout       int      `json:"timeout,omitempty"`
}

// CreateApprovalPolicyRequest is the request body for creating a forge approval policy.
type CreateApprovalPolicyRequest struct {
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	Enabled          bool           `json:"enabled"`
	Priority         int            `json:"priority,omitempty"`
	ApprovalChain    []ApprovalStep `json:"approval_chain,omitempty"`
	AutoApproveAfter int            `json:"auto_approve_after,omitempty"`
}

// UpdateApprovalPolicyRequest is the request body for updating a forge approval policy.
type UpdateApprovalPolicyRequest struct {
	Name             string         `json:"name,omitempty"`
	Description      string         `json:"description,omitempty"`
	Enabled          *bool          `json:"enabled,omitempty"`
	Priority         *int           `json:"priority,omitempty"`
	ApprovalChain    []ApprovalStep `json:"approval_chain,omitempty"`
	AutoApproveAfter *int           `json:"auto_approve_after,omitempty"`
}

// approvalPolicyResponse wraps a single approval policy response.
type approvalPolicyResponse struct {
	Policy ForgeApprovalPolicy `json:"policy"`
}

// approvalPolicyListResponse wraps the list response.
type approvalPolicyListResponse struct {
	Policies []ForgeApprovalPolicy `json:"policies"`
}

// ListApprovalPolicies retrieves all forge approval policies.
func (c *Client) ListApprovalPolicies(ctx context.Context) ([]ForgeApprovalPolicy, error) {
	body, err := c.Get(ctx, "/api/v1/forge/approval-policies")
	if err != nil {
		return nil, fmt.Errorf("list approval policies: %w", err)
	}

	var resp approvalPolicyListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal approval policies response: %w", err)
	}

	return resp.Policies, nil
}

// GetApprovalPolicy retrieves a forge approval policy by ID.
func (c *Client) GetApprovalPolicy(ctx context.Context, id string) (*ForgeApprovalPolicy, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/forge/approval-policies/%s", url.PathEscape(id)))
	if err != nil {
		return nil, fmt.Errorf("get approval policy %s: %w", id, err)
	}

	var resp approvalPolicyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal approval policy response: %w", err)
	}

	return &resp.Policy, nil
}

// CreateApprovalPolicy creates a new forge approval policy.
func (c *Client) CreateApprovalPolicy(ctx context.Context, req CreateApprovalPolicyRequest) (*ForgeApprovalPolicy, error) {
	body, err := c.Post(ctx, "/api/v1/forge/approval-policies", req)
	if err != nil {
		return nil, fmt.Errorf("create approval policy: %w", err)
	}

	var resp approvalPolicyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create approval policy response: %w", err)
	}

	return &resp.Policy, nil
}

// UpdateApprovalPolicy updates an existing forge approval policy.
func (c *Client) UpdateApprovalPolicy(ctx context.Context, id string, req UpdateApprovalPolicyRequest) (*ForgeApprovalPolicy, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/forge/approval-policies/%s", url.PathEscape(id)), req)
	if err != nil {
		return nil, fmt.Errorf("update approval policy %s: %w", id, err)
	}

	var resp approvalPolicyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal update approval policy response: %w", err)
	}

	return &resp.Policy, nil
}

// DeleteApprovalPolicy deletes a forge approval policy by ID.
func (c *Client) DeleteApprovalPolicy(ctx context.Context, id string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/forge/approval-policies/%s", url.PathEscape(id))); err != nil {
		return fmt.Errorf("delete approval policy %s: %w", id, err)
	}
	return nil
}
