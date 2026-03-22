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

func TestListRoles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/roles" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": []map[string]interface{}{
				{"user_id": "user-1", "email": "admin@example.com", "role": "admin"},
				{"user_id": "user-2", "email": "dev@example.com", "role": "viewer"},
			},
			"count": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	roles, err := c.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("ListRoles() error = %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("role count = %d, want 2", len(roles))
	}
	if roles[0].UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", roles[0].UserID, "user-1")
	}
	if roles[0].Role != "admin" {
		t.Errorf("Role = %q, want %q", roles[0].Role, "admin")
	}
	if roles[0].Email != "admin@example.com" {
		t.Errorf("Email = %q, want %q", roles[0].Email, "admin@example.com")
	}
}

func TestGetUserRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": []map[string]interface{}{
				{"user_id": "user-1", "email": "admin@example.com", "role": "admin"},
				{"user_id": "user-2", "email": "dev@example.com", "role": "viewer"},
			},
			"count": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	role, err := c.GetUserRole(context.Background(), "user-2", "viewer")
	if err != nil {
		t.Fatalf("GetUserRole() error = %v", err)
	}
	if role.UserID != "user-2" {
		t.Errorf("UserID = %q, want %q", role.UserID, "user-2")
	}
	if role.Role != "viewer" {
		t.Errorf("Role = %q, want %q", role.Role, "viewer")
	}
	if role.Email != "dev@example.com" {
		t.Errorf("Email = %q, want %q", role.Email, "dev@example.com")
	}
}

func TestGetUserRole_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": []map[string]interface{}{
				{"user_id": "user-1", "role": "admin"},
			},
			"count": 1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetUserRole(context.Background(), "user-99", "admin")
	if err == nil {
		t.Fatal("expected error for not found role, got nil")
	}
}

func TestAddUserRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/roles/users/user-1/roles" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req RoleRequest
		json.Unmarshal(body, &req)

		if req.Role != "admin" {
			t.Errorf("Role = %q, want %q", req.Role, "admin")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Role added"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.AddUserRole(context.Background(), "user-1", RoleRequest{Role: "admin"})
	if err != nil {
		t.Fatalf("AddUserRole() error = %v", err)
	}
}

func TestRemoveUserRole_Success(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.RemoveUserRole(context.Background(), "user-1", RoleRequest{Role: "admin"})
	if err != nil {
		t.Fatalf("RemoveUserRole() error = %v", err)
	}

	// The role MUST be included in the DELETE path to avoid removing all roles
	wantPath := "/api/v1/roles/users/user-1/roles/admin"
	if gotPath != wantPath {
		t.Errorf("DELETE path = %q, want %q (role must be in path)", gotPath, wantPath)
	}
}

func TestRemoveUserRole_SpecialCharsInRole(t *testing.T) {
	var gotRawPath, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawPath = r.URL.RawPath
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.RemoveUserRole(context.Background(), "user-1", RoleRequest{Role: "org/admin"})
	if err != nil {
		t.Fatalf("RemoveUserRole() error = %v", err)
	}

	// RawPath preserves the URL encoding; role containing "/" should be encoded as %2F
	wantRawPath := "/api/v1/roles/users/user-1/roles/org%2Fadmin"
	if gotRawPath != wantRawPath {
		t.Errorf("DELETE raw path = %q, want %q (role must be URL-encoded)", gotRawPath, wantRawPath)
	}

	// Decoded path should show the full role name
	wantDecodedPath := "/api/v1/roles/users/user-1/roles/org/admin"
	if gotPath != wantDecodedPath {
		t.Errorf("DELETE decoded path = %q, want %q", gotPath, wantDecodedPath)
	}
}

func TestRoleClient_Lifecycle_Integration(t *testing.T) {
	roles := make([]map[string]interface{}, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/roles/users/user-1/roles":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			role := map[string]interface{}{
				"user_id": "user-1",
				"email":   "user@example.com",
				"role":    req["role"],
			}
			roles = append(roles, role)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "Role added"})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/roles":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"roles": roles, "count": len(roles)})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/roles/users/user-1/roles/admin":
			// Remove the specific role
			newRoles := make([]map[string]interface{}, 0)
			for _, role := range roles {
				if role["role"] != "admin" {
					newRoles = append(newRoles, role)
				}
			}
			roles = newRoles
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// ADD ROLE
	err := c.AddUserRole(context.Background(), "user-1", RoleRequest{Role: "admin"})
	if err != nil {
		t.Fatalf("ADD failed: %v", err)
	}

	// READ
	role, err := c.GetUserRole(context.Background(), "user-1", "admin")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if role.Role != "admin" {
		t.Errorf("READ: Role = %q, want %q", role.Role, "admin")
	}

	// REMOVE ROLE
	err = c.RemoveUserRole(context.Background(), "user-1", RoleRequest{Role: "admin"})
	if err != nil {
		t.Fatalf("REMOVE failed: %v", err)
	}

	// VERIFY REMOVED
	_, err = c.GetUserRole(context.Background(), "user-1", "admin")
	if err == nil {
		t.Error("expected error after removal, got nil")
	}
}
