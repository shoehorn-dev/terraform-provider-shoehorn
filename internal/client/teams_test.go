package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetTeam_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/admin/teams/team-123" {
			t.Errorf("path = %q, want /api/v1/admin/teams/team-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"team": map[string]interface{}{
				"id":           "team-123",
				"name":         "Platform",
				"display_name": "Platform Engineering",
				"slug":         "platform",
				"description":  "Core platform team",
				"source":       "manual",
				"is_active":    true,
				"member_count": 5,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	team, err := c.GetTeam(context.Background(), "team-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.ID != "team-123" {
		t.Errorf("ID = %q, want %q", team.ID, "team-123")
	}
	if team.Name != "Platform" {
		t.Errorf("Name = %q, want %q", team.Name, "Platform")
	}
	if team.DisplayName != "Platform Engineering" {
		t.Errorf("DisplayName = %q, want %q", team.DisplayName, "Platform Engineering")
	}
	if team.Slug != "platform" {
		t.Errorf("Slug = %q, want %q", team.Slug, "platform")
	}
	if team.Description != "Core platform team" {
		t.Errorf("Description = %q, want %q", team.Description, "Core platform team")
	}
	if team.MemberCount != 5 {
		t.Errorf("MemberCount = %d, want 5", team.MemberCount)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Team not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetTeam(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing team, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		// Error may be wrapped
		t.Logf("error type: %T, value: %v", err, err)
	} else {
		if apiErr.StatusCode != 404 {
			t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
		}
	}
}

func TestListTeams_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/admin/teams" {
			t.Errorf("path = %q, want /api/v1/admin/teams", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"teams": []map[string]interface{}{
				{"id": "team-1", "name": "Team A", "slug": "team-a"},
				{"id": "team-2", "name": "Team B", "slug": "team-b"},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	teams, err := c.ListTeams(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(teams) != 2 {
		t.Fatalf("len(teams) = %d, want 2", len(teams))
	}
	if teams[0].ID != "team-1" {
		t.Errorf("teams[0].ID = %q, want %q", teams[0].ID, "team-1")
	}
	if teams[1].Name != "Team B" {
		t.Errorf("teams[1].Name = %q, want %q", teams[1].Name, "Team B")
	}
}

func TestCreateTeam_Success(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/admin/teams" {
			t.Errorf("path = %q, want /api/v1/admin/teams", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"team": map[string]interface{}{
				"id":           "new-team-id",
				"name":         "Platform",
				"display_name": "Platform Engineering",
				"slug":         "platform",
				"description":  "Core platform team",
				"source":       "manual",
				"is_active":    true,
				"member_count": 0,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	team, err := c.CreateTeam(context.Background(), CreateTeamRequest{
		Name:        "Platform",
		DisplayName: "Platform Engineering",
		Slug:        "platform",
		Description: "Core platform team",
		Metadata:    map[string]interface{}{"cost_center": "engineering"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.ID != "new-team-id" {
		t.Errorf("ID = %q, want %q", team.ID, "new-team-id")
	}
	if team.Name != "Platform" {
		t.Errorf("Name = %q, want %q", team.Name, "Platform")
	}

	// Verify request body
	if gotBody["name"] != "Platform" {
		t.Errorf("request name = %v, want %q", gotBody["name"], "Platform")
	}
	if gotBody["slug"] != "platform" {
		t.Errorf("request slug = %v, want %q", gotBody["slug"], "platform")
	}
}

func TestUpdateTeam_Success(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		if r.URL.Path != "/api/v1/admin/teams/team-123" {
			t.Errorf("path = %q, want /api/v1/admin/teams/team-123", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"team": map[string]interface{}{
				"id":          "team-123",
				"name":        "Updated Platform",
				"slug":        "platform",
				"description": "Updated description",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	team, err := c.UpdateTeam(context.Background(), "team-123", UpdateTeamRequest{
		Name:        "Updated Platform",
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.Name != "Updated Platform" {
		t.Errorf("Name = %q, want %q", team.Name, "Updated Platform")
	}

	if gotBody["name"] != "Updated Platform" {
		t.Errorf("request name = %v, want %q", gotBody["name"], "Updated Platform")
	}
}

func TestDeleteTeam_Success(t *testing.T) {
	var gotMethod string
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteTeam(context.Background(), "team-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/v1/admin/teams/team-456" {
		t.Errorf("path = %q, want /api/v1/admin/teams/team-456", gotPath)
	}
}

func TestDeleteTeam_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Team not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteTeam(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateTeam_WithMetadata(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"team": map[string]interface{}{
				"id":       "new-id",
				"name":     "Test",
				"slug":     "test",
				"metadata": map[string]interface{}{"cost_center": "eng", "level": float64(3)},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	team, err := c.CreateTeam(context.Background(), CreateTeamRequest{
		Name: "Test",
		Slug: "test",
		Metadata: map[string]interface{}{
			"cost_center": "eng",
			"level":       3,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.Metadata == nil {
		t.Fatal("Metadata is nil")
	}
	if team.Metadata["cost_center"] != "eng" {
		t.Errorf("Metadata[cost_center] = %v, want %q", team.Metadata["cost_center"], "eng")
	}

	// Verify metadata sent in request
	metadata, ok := gotBody["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("request metadata not found")
	}
	if metadata["cost_center"] != "eng" {
		t.Errorf("request metadata[cost_center] = %v, want %q", metadata["cost_center"], "eng")
	}
}
