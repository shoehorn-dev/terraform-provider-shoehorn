package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// UserRole represents a user-to-role assignment returned by the Shoehorn API.
type UserRole struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role"`
}

// RoleRequest is the JSON request body for adding a role to a user.
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

// GetUserRole retrieves a specific user's role assignment by listing all roles
// and filtering for the matching userID and role. Returns ErrNotFound if no match exists.
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

	return nil, fmt.Errorf("role %q for user %q: %w", role, userID, ErrNotFound)
}

// AddUserRole assigns a role to a user by POSTing to /api/v1/roles/users/{userID}/roles.
func (c *Client) AddUserRole(ctx context.Context, userID string, req RoleRequest) error {
	path := fmt.Sprintf("/api/v1/roles/users/%s/roles", url.PathEscape(userID))
	_, err := c.Post(ctx, path, req)
	if err != nil {
		return fmt.Errorf("add role %s to user %s: %w", req.Role, userID, err)
	}
	return nil
}

// RemoveUserRole removes a specific role from a user by issuing a DELETE to
// /api/v1/roles/users/{userID}/roles/{role}.
func (c *Client) RemoveUserRole(ctx context.Context, userID string, req RoleRequest) error {
	path := fmt.Sprintf("/api/v1/roles/users/%s/roles/%s", url.PathEscape(userID), url.PathEscape(req.Role))
	if err := c.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove role %s from user %s: %w", req.Role, userID, err)
	}
	return nil
}
