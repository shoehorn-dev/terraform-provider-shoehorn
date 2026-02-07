package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestTeamResource_Metadata(t *testing.T) {
	r := NewTeamResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_team" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_team")
	}
}

func TestTeamResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewTeamResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	requiredAttrs := []string{"id", "name", "slug", "description", "metadata", "members", "display_name", "is_active", "member_count", "created_at", "updated_at"}
	for _, name := range requiredAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestTeamResource_Schema_NameIsRequired(t *testing.T) {
	r := NewTeamResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	nameAttr := resp.Schema.Attributes["name"]
	if nameAttr == nil {
		t.Fatal("name attribute not found")
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestTeamResource_Schema_SlugIsRequired(t *testing.T) {
	r := NewTeamResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	slugAttr := resp.Schema.Attributes["slug"]
	if slugAttr == nil {
		t.Fatal("slug attribute not found")
	}
	if !slugAttr.IsRequired() {
		t.Error("slug should be required")
	}
}

func TestTeamResource_Schema_IDIsComputed(t *testing.T) {
	r := NewTeamResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	idAttr := resp.Schema.Attributes["id"]
	if idAttr == nil {
		t.Fatal("id attribute not found")
	}
	if !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestTeamResource_Configure_WithValidClient(t *testing.T) {
	r := &TeamResource{}
	c := client.NewClient("https://test.example.com", "key", 30*time.Second)

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: c,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}
	if r.client != c {
		t.Error("client not set correctly")
	}
}

func TestTeamResource_Configure_NilProviderData(t *testing.T) {
	r := &TeamResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Error("should not error on nil provider data")
	}
}

func TestTeamResource_Configure_WrongType(t *testing.T) {
	r := &TeamResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

// TestTeamClient_CRUD_Integration tests the full CRUD lifecycle using a mock server.
func TestTeamClient_CRUD_Integration(t *testing.T) {
	teams := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/teams":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			team := map[string]interface{}{
				"id":           "generated-id-123",
				"name":         req["name"],
				"display_name": req["display_name"],
				"slug":         req["slug"],
				"description":  req["description"],
				"source":       "manual",
				"is_active":    true,
				"member_count": float64(0),
				"created_at":   "2025-01-15T10:30:00Z",
				"updated_at":   "2025-01-15T10:30:00Z",
			}
			if req["metadata"] != nil {
				team["metadata"] = req["metadata"]
			}
			teams["generated-id-123"] = team

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"team": team})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/admin/teams/generated-id-123":
			if team, ok := teams["generated-id-123"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"team": team})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "Team not found"})
			}

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/admin/teams/generated-id-123":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			team := teams["generated-id-123"]
			if name, ok := req["name"].(string); ok && name != "" {
				team["name"] = name
			}
			if desc, ok := req["description"].(string); ok {
				team["description"] = desc
			}
			if dn, ok := req["display_name"].(string); ok {
				team["display_name"] = dn
			}
			team["updated_at"] = "2025-01-15T11:00:00Z"
			teams["generated-id-123"] = team

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"team": team})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/admin/teams/generated-id-123":
			delete(teams, "generated-id-123")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	team, err := c.CreateTeam(context.Background(), client.CreateTeamRequest{
		Name:        "Platform",
		DisplayName: "Platform Engineering",
		Slug:        "platform",
		Description: "Core platform team",
		Metadata:    map[string]interface{}{"cost_center": "engineering"},
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if team.ID != "generated-id-123" {
		t.Errorf("CREATE: ID = %q, want %q", team.ID, "generated-id-123")
	}
	if team.Name != "Platform" {
		t.Errorf("CREATE: Name = %q, want %q", team.Name, "Platform")
	}

	// READ
	readTeam, err := c.GetTeam(context.Background(), "generated-id-123")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if readTeam.Name != "Platform" {
		t.Errorf("READ: Name = %q, want %q", readTeam.Name, "Platform")
	}
	if readTeam.Description != "Core platform team" {
		t.Errorf("READ: Description = %q, want %q", readTeam.Description, "Core platform team")
	}

	// UPDATE
	updatedTeam, err := c.UpdateTeam(context.Background(), "generated-id-123", client.UpdateTeamRequest{
		Name:        "Updated Platform",
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updatedTeam.Name != "Updated Platform" {
		t.Errorf("UPDATE: Name = %q, want %q", updatedTeam.Name, "Updated Platform")
	}

	// DELETE
	err = c.DeleteTeam(context.Background(), "generated-id-123")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// Verify deleted (should 404)
	_, err = c.GetTeam(context.Background(), "generated-id-123")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestMapTeamToState(t *testing.T) {
	team := &client.Team{
		ID:          "team-123",
		Name:        "Platform",
		DisplayName: "Platform Engineering",
		Slug:        "platform",
		Description: "Core platform team",
		IsActive:    true,
		MemberCount: 5,
		Metadata:    map[string]interface{}{"cost_center": "eng"},
		CreatedAt:   "2025-01-15T10:30:00Z",
		UpdatedAt:   "2025-01-15T11:00:00Z",
	}

	state := &TeamResourceModel{}
	mapTeamToState(team, state)

	if state.ID.ValueString() != "team-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "team-123")
	}
	if state.Name.ValueString() != "Platform" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "Platform")
	}
	if state.DisplayName.ValueString() != "Platform Engineering" {
		t.Errorf("DisplayName = %q, want %q", state.DisplayName.ValueString(), "Platform Engineering")
	}
	if state.Slug.ValueString() != "platform" {
		t.Errorf("Slug = %q, want %q", state.Slug.ValueString(), "platform")
	}
	if state.Description.ValueString() != "Core platform team" {
		t.Errorf("Description = %q, want %q", state.Description.ValueString(), "Core platform team")
	}
	if !state.IsActive.ValueBool() {
		t.Error("IsActive should be true")
	}
	if state.MemberCount.ValueInt64() != 5 {
		t.Errorf("MemberCount = %d, want 5", state.MemberCount.ValueInt64())
	}
	if state.CreatedAt.ValueString() != "2025-01-15T10:30:00Z" {
		t.Errorf("CreatedAt = %q, want %q", state.CreatedAt.ValueString(), "2025-01-15T10:30:00Z")
	}
	if state.UpdatedAt.ValueString() != "2025-01-15T11:00:00Z" {
		t.Errorf("UpdatedAt = %q, want %q", state.UpdatedAt.ValueString(), "2025-01-15T11:00:00Z")
	}

	// Verify metadata is JSON
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(state.Metadata.ValueString()), &metadata); err != nil {
		t.Fatalf("metadata not valid JSON: %v", err)
	}
	if metadata["cost_center"] != "eng" {
		t.Errorf("metadata[cost_center] = %v, want %q", metadata["cost_center"], "eng")
	}
}

func TestTeamResource_Schema_MembersIsOptional(t *testing.T) {
	r := NewTeamResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	memAttr := resp.Schema.Attributes["members"]
	if memAttr == nil {
		t.Fatal("members attribute not found")
	}
	if !memAttr.IsOptional() {
		t.Error("members should be optional")
	}
}

func TestMapTeamToState_WithMembers(t *testing.T) {
	team := &client.Team{
		ID:          "team-123",
		Name:        "Platform",
		Slug:        "platform",
		IsActive:    true,
		MemberCount: 2,
		Members: []client.TeamMember{
			{ID: "m1", TeamID: "team-123", UserID: "alice@example.com", Role: "manager"},
			{ID: "m2", TeamID: "team-123", UserID: "bob@example.com", Role: "member"},
		},
	}

	state := &TeamResourceModel{}
	mapTeamToState(team, state)

	if state.Members.IsNull() {
		t.Fatal("Members should not be null")
	}

	var members []tfMemberEntry
	if err := json.Unmarshal([]byte(state.Members.ValueString()), &members); err != nil {
		t.Fatalf("Failed to unmarshal members: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("Expected 2 members, got %d", len(members))
	}
	if members[0].UserID != "alice@example.com" {
		t.Errorf("members[0].UserID = %q, want %q", members[0].UserID, "alice@example.com")
	}
	if members[0].Role != "manager" {
		t.Errorf("members[0].Role = %q, want %q", members[0].Role, "manager")
	}
}

func TestComputeMemberDiff_AddMembers(t *testing.T) {
	state := types.StringValue(`[]`)
	plan := types.StringValue(`[{"user_id":"alice@example.com","role":"manager"}]`)

	adds, removes := computeMemberDiff(state, plan)

	if len(adds) != 1 {
		t.Fatalf("Expected 1 add, got %d", len(adds))
	}
	if adds[0].UserID != "alice@example.com" {
		t.Errorf("add[0].UserID = %q, want %q", adds[0].UserID, "alice@example.com")
	}
	if adds[0].Role != "manager" {
		t.Errorf("add[0].Role = %q, want %q", adds[0].Role, "manager")
	}
	if len(removes) != 0 {
		t.Errorf("Expected 0 removes, got %d", len(removes))
	}
}

func TestComputeMemberDiff_RemoveMembers(t *testing.T) {
	state := types.StringValue(`[{"user_id":"alice@example.com","role":"admin"},{"user_id":"bob@example.com","role":"member"}]`)
	plan := types.StringValue(`[{"user_id":"alice@example.com","role":"admin"}]`)

	adds, removes := computeMemberDiff(state, plan)

	if len(adds) != 0 {
		t.Errorf("Expected 0 adds, got %d", len(adds))
	}
	if len(removes) != 1 {
		t.Fatalf("Expected 1 remove, got %d", len(removes))
	}
	if removes[0] != "bob@example.com" {
		t.Errorf("removes[0] = %q, want %q", removes[0], "bob@example.com")
	}
}

func TestComputeMemberDiff_RoleChange(t *testing.T) {
	state := types.StringValue(`[{"user_id":"alice@example.com","role":"member"}]`)
	plan := types.StringValue(`[{"user_id":"alice@example.com","role":"manager"}]`)

	adds, removes := computeMemberDiff(state, plan)

	if len(adds) != 1 {
		t.Fatalf("Expected 1 add (role change), got %d", len(adds))
	}
	if adds[0].Role != "manager" {
		t.Errorf("add[0].Role = %q, want %q", adds[0].Role, "manager")
	}
	if len(removes) != 0 {
		t.Errorf("Expected 0 removes, got %d", len(removes))
	}
}

func TestComputeMemberDiff_NullToMembers(t *testing.T) {
	state := types.StringNull()
	plan := types.StringValue(`[{"user_id":"alice@example.com","role":"admin"}]`)

	adds, removes := computeMemberDiff(state, plan)

	if len(adds) != 1 {
		t.Fatalf("Expected 1 add, got %d", len(adds))
	}
	if len(removes) != 0 {
		t.Errorf("Expected 0 removes, got %d", len(removes))
	}
}

func TestComputeMemberDiff_MembersToNull(t *testing.T) {
	state := types.StringValue(`[{"user_id":"alice@example.com","role":"admin"}]`)
	plan := types.StringNull()

	adds, removes := computeMemberDiff(state, plan)

	if len(adds) != 0 {
		t.Errorf("Expected 0 adds, got %d", len(adds))
	}
	if len(removes) != 1 {
		t.Fatalf("Expected 1 remove, got %d", len(removes))
	}
}

func TestMembersEquivalent(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "same order",
			a:    `[{"user_id":"alice","role":"admin"},{"user_id":"bob","role":"member"}]`,
			b:    `[{"user_id":"alice","role":"admin"},{"user_id":"bob","role":"member"}]`,
			want: true,
		},
		{
			name: "different order",
			a:    `[{"user_id":"bob","role":"member"},{"user_id":"alice","role":"admin"}]`,
			b:    `[{"user_id":"alice","role":"admin"},{"user_id":"bob","role":"member"}]`,
			want: true,
		},
		{
			name: "different role",
			a:    `[{"user_id":"alice","role":"admin"}]`,
			b:    `[{"user_id":"alice","role":"member"}]`,
			want: false,
		},
		{
			name: "different count",
			a:    `[{"user_id":"alice","role":"admin"}]`,
			b:    `[{"user_id":"alice","role":"admin"},{"user_id":"bob","role":"member"}]`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := membersEquivalent(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("membersEquivalent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapTeamToState_EmptyOptionalFields(t *testing.T) {
	team := &client.Team{
		ID:   "team-123",
		Name: "Platform",
		Slug: "platform",
	}

	state := &TeamResourceModel{}
	mapTeamToState(team, state)

	if state.ID.ValueString() != "team-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "team-123")
	}

	// Optional fields should not crash
	if state.Metadata.IsNull() || state.Metadata.ValueString() != "" {
		// Metadata should be empty or null - not crash
	}
}
