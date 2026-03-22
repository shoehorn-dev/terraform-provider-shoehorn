package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Group represents an IdP group.
type Group struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	MemberCount int             `json:"memberCount"`
	SubGroups   []Group         `json:"subGroups,omitempty"`
	Roles       []GroupRoleInfo `json:"roles,omitempty"`
}

// GroupRoleInfo represents a role mapping for a group.
type GroupRoleInfo struct {
	RoleName           string `json:"roleName"`
	BundleDisplayName  string `json:"bundleDisplayName,omitempty"`
	BundleColor        string `json:"bundleColor,omitempty"`
	Provider           string `json:"provider,omitempty"`
}

// GroupRoleRequest is the request body for assigning a role to a group.
type GroupRoleRequest struct {
	RoleName    string `json:"role_name"`
	Provider    string `json:"provider,omitempty"`
	Description string `json:"description,omitempty"`
}

// groupListResponse wraps the list groups response.
type groupListResponse struct {
	Items []Group `json:"items"`
}

// groupRolesResponse wraps the group roles response.
type groupRolesResponse struct {
	Roles []GroupRoleInfo `json:"roles"`
}

// ListGroups retrieves all IdP groups.
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	body, err := c.Get(ctx, "/api/v1/groups")
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}

	var resp groupListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal groups response: %w", err)
	}

	return resp.Items, nil
}

// GetGroupRoles retrieves role mappings for a specific group.
func (c *Client) GetGroupRoles(ctx context.Context, groupName string) ([]GroupRoleInfo, error) {
	path := fmt.Sprintf("/api/v1/groups/%s/roles", url.PathEscape(groupName))
	body, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("get group roles for %q: %w", groupName, err)
	}

	var resp groupRolesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal group roles response: %w", err)
	}

	return resp.Roles, nil
}

// AssignGroupRole assigns a role to a group.
func (c *Client) AssignGroupRole(ctx context.Context, groupName string, req GroupRoleRequest) error {
	path := fmt.Sprintf("/api/v1/groups/%s/roles", url.PathEscape(groupName))
	_, err := c.Post(ctx, path, req)
	if err != nil {
		return fmt.Errorf("assign role %q to group %q: %w", req.RoleName, groupName, err)
	}
	return nil
}

// RemoveGroupRole removes a role from a group.
func (c *Client) RemoveGroupRole(ctx context.Context, groupName, roleName string) error {
	path := fmt.Sprintf("/api/v1/groups/%s/roles/%s", url.PathEscape(groupName), url.PathEscape(roleName))
	if err := c.Delete(ctx, path); err != nil {
		return fmt.Errorf("remove role %q from group %q: %w", roleName, groupName, err)
	}
	return nil
}
