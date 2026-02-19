package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListGroups_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/groups" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "g-1", "name": "platform-team", "path": "/platform-team", "memberCount": 5},
				{"id": "g-2", "name": "developers", "path": "/developers", "memberCount": 12},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	groups, err := c.ListGroups(context.Background())
	if err != nil {
		t.Fatalf("ListGroups() error = %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("group count = %d, want 2", len(groups))
	}
	if groups[0].Name != "platform-team" {
		t.Errorf("Name = %q, want %q", groups[0].Name, "platform-team")
	}
	if groups[1].MemberCount != 12 {
		t.Errorf("MemberCount = %d, want 12", groups[1].MemberCount)
	}
}

func TestListGroups_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	groups, err := c.ListGroups(context.Background())
	if err != nil {
		t.Fatalf("ListGroups() error = %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("group count = %d, want 0", len(groups))
	}
}

func TestGetGroupRoles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/groups/platform-team/roles" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": []map[string]interface{}{
				{"roleName": "tenant:admin", "bundleDisplayName": "Admin", "provider": "default"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	roles, err := c.GetGroupRoles(context.Background(), "platform-team")
	if err != nil {
		t.Fatalf("GetGroupRoles() error = %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("role count = %d, want 1", len(roles))
	}
	if roles[0].RoleName != "tenant:admin" {
		t.Errorf("RoleName = %q, want %q", roles[0].RoleName, "tenant:admin")
	}
	if roles[0].BundleDisplayName != "Admin" {
		t.Errorf("BundleDisplayName = %q, want %q", roles[0].BundleDisplayName, "Admin")
	}
}

func TestAssignGroupRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/groups/platform-team/roles" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req GroupRoleRequest
		json.Unmarshal(body, &req)
		if req.RoleName != "tenant:admin" {
			t.Errorf("RoleName = %q, want %q", req.RoleName, "tenant:admin")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Role assigned"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.AssignGroupRole(context.Background(), "platform-team", GroupRoleRequest{
		RoleName: "tenant:admin",
		Provider: "default",
	})
	if err != nil {
		t.Fatalf("AssignGroupRole() error = %v", err)
	}
}

func TestRemoveGroupRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/groups/platform-team/roles/tenant:admin" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.RemoveGroupRole(context.Background(), "platform-team", "tenant:admin")
	if err != nil {
		t.Fatalf("RemoveGroupRole() error = %v", err)
	}
}

func TestGroupRoleClient_Lifecycle_Integration(t *testing.T) {
	roles := make([]map[string]interface{}, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/groups/dev-team/roles":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			roles = append(roles, map[string]interface{}{
				"roleName": req["role_name"],
				"provider": req["provider"],
			})
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "Role assigned"})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/groups/dev-team/roles":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"roles": roles})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/groups/dev-team/roles/entity:viewer":
			roles = []map[string]interface{}{}
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// ASSIGN
	err := c.AssignGroupRole(context.Background(), "dev-team", GroupRoleRequest{
		RoleName: "entity:viewer",
		Provider: "default",
	})
	if err != nil {
		t.Fatalf("ASSIGN failed: %v", err)
	}

	// READ
	foundRoles, err := c.GetGroupRoles(context.Background(), "dev-team")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if len(foundRoles) != 1 {
		t.Fatalf("role count = %d, want 1", len(foundRoles))
	}
	if foundRoles[0].RoleName != "entity:viewer" {
		t.Errorf("RoleName = %q, want %q", foundRoles[0].RoleName, "entity:viewer")
	}

	// REMOVE
	err = c.RemoveGroupRole(context.Background(), "dev-team", "entity:viewer")
	if err != nil {
		t.Fatalf("REMOVE failed: %v", err)
	}

	// VERIFY REMOVED
	foundRoles, err = c.GetGroupRoles(context.Background(), "dev-team")
	if err != nil {
		t.Fatalf("READ after removal failed: %v", err)
	}
	if len(foundRoles) != 0 {
		t.Errorf("role count after removal = %d, want 0", len(foundRoles))
	}
}
