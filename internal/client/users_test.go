package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListDirectoryUsers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/users" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id":        "u-1",
					"username":  "alice",
					"firstName": "Alice",
					"lastName":  "Smith",
					"email":     "alice@example.com",
					"enabled":   true,
					"bundles": []map[string]interface{}{
						{"id": "b-1", "name": "admin", "displayName": "Admin", "color": "#3b82f6"},
					},
				},
				{
					"id":       "u-2",
					"username": "bob",
					"email":    "bob@example.com",
					"enabled":  false,
				},
			},
			"provider": "zitadel",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	users, err := c.ListDirectoryUsers(context.Background())
	if err != nil {
		t.Fatalf("ListDirectoryUsers() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("user count = %d, want 2", len(users))
	}
	if users[0].Username != "alice" {
		t.Errorf("Username = %q, want %q", users[0].Username, "alice")
	}
	if users[0].Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", users[0].Email, "alice@example.com")
	}
	if !users[0].Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if len(users[0].Bundles) != 1 {
		t.Errorf("bundle count = %d, want 1", len(users[0].Bundles))
	}
	if users[0].Bundles[0].DisplayName != "Admin" {
		t.Errorf("BundleDisplayName = %q, want %q", users[0].Bundles[0].DisplayName, "Admin")
	}
	if users[1].Username != "bob" {
		t.Errorf("Username = %q, want %q", users[1].Username, "bob")
	}
	if users[1].Enabled {
		t.Errorf("Enabled = true, want false")
	}
}

func TestListDirectoryUsers_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	users, err := c.ListDirectoryUsers(context.Background())
	if err != nil {
		t.Fatalf("ListDirectoryUsers() error = %v", err)
	}
	if len(users) != 0 {
		t.Errorf("user count = %d, want 0", len(users))
	}
}

func TestGetDirectoryUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/users/u-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "u-1",
			"username":  "alice",
			"firstName": "Alice",
			"lastName":  "Smith",
			"email":     "alice@example.com",
			"enabled":   true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	user, err := c.GetDirectoryUser(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("GetDirectoryUser() error = %v", err)
	}
	if user.ID != "u-1" {
		t.Errorf("ID = %q, want %q", user.ID, "u-1")
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want %q", user.Username, "alice")
	}
	if user.FirstName != "Alice" {
		t.Errorf("FirstName = %q, want %q", user.FirstName, "Alice")
	}
}

func TestGetDirectoryUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "User not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetDirectoryUser(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found user, got nil")
	}
}
