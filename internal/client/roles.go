package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// UserRole represents a user role assignment.
type UserRole struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role"`
}

// RoleRequest is the request body for adding/removing a role.
type RoleRequest struct {
	Role string `json:"role"`
}

// roleListResponse wraps the list roles response.
type roleListResponse struct {
	Roles []UserRole `json:"roles"`
	Count int        `json:"count"`
}

// ListRoles retrieves all role assignments.
func (c *Client) ListRoles(ctx context.Context) ([]UserRole, error) {
	body, err := c.Get(ctx, "/api/v1/roles")
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}

	var resp roleListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal roles response: %w", err)
	}

	return resp.Roles, nil
}

// GetUserRole retrieves a specific user's role assignment.
func (c *Client) GetUserRole(ctx context.Context, userID, role string) (*UserRole, error) {
	roles, err := c.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range roles {
		if r.UserID == userID && r.Role == role {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("role %q for user %q not found", role, userID)
}

// AddUserRole adds a role to a user.
func (c *Client) AddUserRole(ctx context.Context, userID string, req RoleRequest) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/v1/roles/users/%s/roles", userID), req)
	if err != nil {
		return fmt.Errorf("add role %s to user %s: %w", req.Role, userID, err)
	}
	return nil
}

// RemoveUserRole removes a role from a user.
func (c *Client) RemoveUserRole(ctx context.Context, userID string, req RoleRequest) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/roles/users/%s/roles", userID)); err != nil {
		return fmt.Errorf("remove role %s from user %s: %w", req.Role, userID, err)
	}
	return nil
}
