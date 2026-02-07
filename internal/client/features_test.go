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

func TestGetFeatureFlag_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/features" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"flags": []map[string]interface{}{
				{
					"id": "flag-1", "key": "dark-mode", "name": "Dark Mode",
					"description": "Enable dark mode", "default_enabled": true,
					"override_count": 2, "created_at": "2025-01-15T10:00:00Z",
					"updated_at": "2025-01-15T10:00:00Z",
				},
				{
					"id": "flag-2", "key": "beta-features", "name": "Beta Features",
					"default_enabled": false,
				},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	flag, err := c.GetFeatureFlag(context.Background(), "dark-mode")
	if err != nil {
		t.Fatalf("GetFeatureFlag() error = %v", err)
	}

	if flag.Key != "dark-mode" {
		t.Errorf("Key = %q, want %q", flag.Key, "dark-mode")
	}
	if flag.Name != "Dark Mode" {
		t.Errorf("Name = %q, want %q", flag.Name, "Dark Mode")
	}
	if !flag.DefaultEnabled {
		t.Error("DefaultEnabled = false, want true")
	}
	if flag.Description != "Enable dark mode" {
		t.Errorf("Description = %q, want %q", flag.Description, "Enable dark mode")
	}
}

func TestGetFeatureFlag_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"flags": []interface{}{}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetFeatureFlag(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found flag, got nil")
	}
}

func TestListFeatureFlags_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"flags": []map[string]interface{}{
				{"id": "flag-1", "key": "dark-mode", "name": "Dark Mode", "default_enabled": true},
				{"id": "flag-2", "key": "beta", "name": "Beta", "default_enabled": false},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	flags, err := c.ListFeatureFlags(context.Background())
	if err != nil {
		t.Fatalf("ListFeatureFlags() error = %v", err)
	}
	if len(flags) != 2 {
		t.Errorf("flag count = %d, want 2", len(flags))
	}
}

func TestCreateFeatureFlag_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/admin/features" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateFeatureFlagRequest
		json.Unmarshal(body, &req)

		if req.Key != "new-flag" {
			t.Errorf("Key = %q, want %q", req.Key, "new-flag")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "flag-new", "key": req.Key, "name": req.Name,
			"description": req.Description, "default_enabled": req.DefaultEnabled,
			"created_at": "2025-01-15T10:00:00Z", "updated_at": "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	flag, err := c.CreateFeatureFlag(context.Background(), CreateFeatureFlagRequest{
		Key:            "new-flag",
		Name:           "New Flag",
		Description:    "A new feature flag",
		DefaultEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateFeatureFlag() error = %v", err)
	}

	if flag.Key != "new-flag" {
		t.Errorf("Key = %q, want %q", flag.Key, "new-flag")
	}
	if flag.Name != "New Flag" {
		t.Errorf("Name = %q, want %q", flag.Name, "New Flag")
	}
	if !flag.DefaultEnabled {
		t.Error("DefaultEnabled = false, want true")
	}
}

func TestUpdateFeatureFlag_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/admin/features/dark-mode" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "flag-1", "key": "dark-mode", "name": "Updated Name",
			"description": "Updated desc", "default_enabled": false,
			"created_at": "2025-01-15T10:00:00Z", "updated_at": "2025-01-15T12:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	enabled := false
	flag, err := c.UpdateFeatureFlag(context.Background(), "dark-mode", UpdateFeatureFlagRequest{
		Name:           "Updated Name",
		DefaultEnabled: &enabled,
	})
	if err != nil {
		t.Fatalf("UpdateFeatureFlag() error = %v", err)
	}

	if flag.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", flag.Name, "Updated Name")
	}
	if flag.DefaultEnabled {
		t.Error("DefaultEnabled = true, want false")
	}
}

func TestDeleteFeatureFlag_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/admin/features/old-flag" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteFeatureFlag(context.Background(), "old-flag")
	if err != nil {
		t.Fatalf("DeleteFeatureFlag() error = %v", err)
	}
}

func TestFeatureFlagClient_CRUD_Integration(t *testing.T) {
	flags := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/features":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			flag := map[string]interface{}{
				"id": "flag-123", "key": req["key"], "name": req["name"],
				"description": req["description"], "default_enabled": req["default_enabled"],
				"created_at": "2025-01-15T10:00:00Z", "updated_at": "2025-01-15T10:00:00Z",
			}
			flags[req["key"].(string)] = flag

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(flag)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/admin/features":
			flagList := make([]map[string]interface{}, 0)
			for _, f := range flags {
				flagList = append(flagList, f)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"flags": flagList})

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/admin/features/test-flag":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			flag := flags["test-flag"]
			if name, ok := req["name"].(string); ok && name != "" {
				flag["name"] = name
			}
			if enabled, ok := req["default_enabled"].(bool); ok {
				flag["default_enabled"] = enabled
			}
			flag["updated_at"] = "2025-01-15T12:00:00Z"
			flags["test-flag"] = flag

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(flag)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/admin/features/test-flag":
			delete(flags, "test-flag")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	flag, err := c.CreateFeatureFlag(context.Background(), CreateFeatureFlagRequest{
		Key: "test-flag", Name: "Test Flag", Description: "A test flag", DefaultEnabled: true,
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if flag.Key != "test-flag" {
		t.Errorf("CREATE: Key = %q, want %q", flag.Key, "test-flag")
	}

	// READ
	readFlag, err := c.GetFeatureFlag(context.Background(), "test-flag")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if readFlag.Name != "Test Flag" {
		t.Errorf("READ: Name = %q, want %q", readFlag.Name, "Test Flag")
	}

	// UPDATE
	enabled := false
	updatedFlag, err := c.UpdateFeatureFlag(context.Background(), "test-flag", UpdateFeatureFlagRequest{
		Name: "Updated Flag", DefaultEnabled: &enabled,
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updatedFlag.Name != "Updated Flag" {
		t.Errorf("UPDATE: Name = %q, want %q", updatedFlag.Name, "Updated Flag")
	}

	// DELETE
	err = c.DeleteFeatureFlag(context.Background(), "test-flag")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// Verify deleted
	_, err = c.GetFeatureFlag(context.Background(), "test-flag")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
