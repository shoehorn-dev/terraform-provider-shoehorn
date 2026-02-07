package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Team represents a Shoehorn team.
type Team struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id,omitempty"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name,omitempty"`
	Slug         string                 `json:"slug"`
	Source       string                 `json:"source,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IsActive     bool                   `json:"is_active,omitempty"`
	MemberCount  int                    `json:"member_count,omitempty"`
	Members      []TeamMember           `json:"members,omitempty"`
	ParentTeamID *string                `json:"parent_team_id,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
	UpdatedAt    string                 `json:"updated_at,omitempty"`
}

// TeamMember represents a team member in the API response.
type TeamMember struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
}

// AddMemberRequest is the request to add a member to a team.
type AddMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

// CreateTeamRequest is the request body for creating a team.
type CreateTeamRequest struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name,omitempty"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTeamRequest is the request body for updating a team.
type UpdateTeamRequest struct {
	Name          string                 `json:"name,omitempty"`
	DisplayName   string                 `json:"display_name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ParentTeamID  *string                `json:"parent_team_id,omitempty"`
	AddMembers    []AddMemberRequest     `json:"add_members,omitempty"`
	RemoveMembers []string               `json:"remove_members,omitempty"`
}

// teamResponse wraps a single team in the API response.
// The API returns members both in team.members and at the top level.
type teamResponse struct {
	Team    Team         `json:"team"`
	Members []TeamMember `json:"members,omitempty"`
}

// teamsResponse wraps a list of teams in the API response.
type teamsResponse struct {
	Teams []Team `json:"teams"`
	Total int    `json:"total"`
}

// GetTeam retrieves a team by ID.
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/api/v1/admin/teams/%s", id))
	if err != nil {
		return nil, fmt.Errorf("get team %s: %w", id, err)
	}

	var resp teamResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal team response: %w", err)
	}

	// Use top-level members if team.members is empty
	if len(resp.Team.Members) == 0 && len(resp.Members) > 0 {
		resp.Team.Members = resp.Members
	}

	return &resp.Team, nil
}

// ListTeams retrieves all teams.
func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	body, err := c.Get(ctx, "/api/v1/admin/teams")
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	var resp teamsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal teams response: %w", err)
	}

	return resp.Teams, nil
}

// CreateTeam creates a new team.
func (c *Client) CreateTeam(ctx context.Context, req CreateTeamRequest) (*Team, error) {
	body, err := c.Post(ctx, "/api/v1/admin/teams", req)
	if err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	var resp teamResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create team response: %w", err)
	}

	return &resp.Team, nil
}

// UpdateTeam updates an existing team.
func (c *Client) UpdateTeam(ctx context.Context, id string, req UpdateTeamRequest) (*Team, error) {
	body, err := c.Put(ctx, fmt.Sprintf("/api/v1/admin/teams/%s", id), req)
	if err != nil {
		return nil, fmt.Errorf("update team %s: %w", id, err)
	}

	var resp teamResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal update team response: %w", err)
	}

	return &resp.Team, nil
}

// DeleteTeam deletes a team by ID.
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	if err := c.Delete(ctx, fmt.Sprintf("/api/v1/admin/teams/%s", id)); err != nil {
		return fmt.Errorf("delete team %s: %w", id, err)
	}

	return nil
}
