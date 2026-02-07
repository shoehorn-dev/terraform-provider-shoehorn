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

func TestListAPIKeys_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/api-keys" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"id": "key-1", "name": "CI Key", "key_prefix": "shp_svc_abc12345",
					"scopes": []string{"entities:read", "catalog:read"},
					"created_at": "2025-01-15T10:00:00Z",
				},
				{
					"id": "key-2", "name": "Admin Key", "key_prefix": "shp_svc_def67890",
					"scopes": []string{"admin:write"},
					"created_at": "2025-01-15T11:00:00Z",
				},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	keys, err := c.ListAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("ListAPIKeys() error = %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("key count = %d, want 2", len(keys))
	}
	if keys[0].Name != "CI Key" {
		t.Errorf("Name = %q, want %q", keys[0].Name, "CI Key")
	}
	if keys[0].KeyPrefix != "shp_svc_abc12345" {
		t.Errorf("KeyPrefix = %q, want %q", keys[0].KeyPrefix, "shp_svc_abc12345")
	}
}

func TestGetAPIKey_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]interface{}{
				{"id": "key-1", "name": "CI Key", "key_prefix": "shp_svc_abc", "scopes": []string{"entities:read"}},
				{"id": "key-2", "name": "Admin Key", "key_prefix": "shp_svc_def", "scopes": []string{"admin:write"}},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	apiKey, err := c.GetAPIKey(context.Background(), "key-2")
	if err != nil {
		t.Fatalf("GetAPIKey() error = %v", err)
	}
	if apiKey.Name != "Admin Key" {
		t.Errorf("Name = %q, want %q", apiKey.Name, "Admin Key")
	}
}

func TestGetAPIKey_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"keys": []interface{}{}, "total": 0})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetAPIKey(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found key, got nil")
	}
}

func TestCreateAPIKey_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/admin/api-keys" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateAPIKeyRequest
		json.Unmarshal(body, &req)

		if req.Name != "New Key" {
			t.Errorf("Name = %q, want %q", req.Name, "New Key")
		}
		if len(req.Scopes) != 2 {
			t.Errorf("Scopes count = %d, want 2", len(req.Scopes))
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"key": map[string]interface{}{
				"id": "key-new", "name": "New Key", "key_prefix": "shp_svc_new12345",
				"scopes": req.Scopes, "description": req.Description,
				"created_at": "2025-01-15T10:00:00Z", "updated_at": "2025-01-15T10:00:00Z",
			},
			"raw_key": "shp_svc_new12345_full_secret_key_that_is_72_chars_long_padpadpadpadpadpadpad",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resp, err := c.CreateAPIKey(context.Background(), CreateAPIKeyRequest{
		Name:        "New Key",
		Description: "A new API key",
		Scopes:      []string{"entities:read", "catalog:read"},
	})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	if resp.Key.ID != "key-new" {
		t.Errorf("Key.ID = %q, want %q", resp.Key.ID, "key-new")
	}
	if resp.Key.Name != "New Key" {
		t.Errorf("Key.Name = %q, want %q", resp.Key.Name, "New Key")
	}
	if resp.RawKey == "" {
		t.Error("RawKey should not be empty on creation")
	}
}

func TestCreateAPIKey_WithExpiry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req CreateAPIKeyRequest
		json.Unmarshal(body, &req)

		if req.ExpiresInDays == nil || *req.ExpiresInDays != 90 {
			t.Errorf("ExpiresInDays = %v, want 90", req.ExpiresInDays)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"key": map[string]interface{}{
				"id": "key-exp", "name": "Expiring Key", "key_prefix": "shp_svc_exp",
				"scopes": req.Scopes, "expires_at": "2025-04-15T10:00:00Z",
				"created_at": "2025-01-15T10:00:00Z",
			},
			"raw_key": "shp_svc_exp_secret_key",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	days := 90
	resp, err := c.CreateAPIKey(context.Background(), CreateAPIKeyRequest{
		Name:          "Expiring Key",
		Scopes:        []string{"entities:read"},
		ExpiresInDays: &days,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	if resp.Key.ExpiresAt != "2025-04-15T10:00:00Z" {
		t.Errorf("ExpiresAt = %q, want %q", resp.Key.ExpiresAt, "2025-04-15T10:00:00Z")
	}
}

func TestRevokeAPIKey_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/admin/api-keys/key-1/revoke" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "API key revoked"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.RevokeAPIKey(context.Background(), "key-1")
	if err != nil {
		t.Fatalf("RevokeAPIKey() error = %v", err)
	}
}

func TestAPIKeyClient_Lifecycle_Integration(t *testing.T) {
	keys := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/api-keys":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			key := map[string]interface{}{
				"id": "key-123", "name": req["name"], "key_prefix": "shp_svc_test",
				"scopes": req["scopes"], "description": req["description"],
				"created_at": "2025-01-15T10:00:00Z", "updated_at": "2025-01-15T10:00:00Z",
			}
			keys["key-123"] = key

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"key":     key,
				"raw_key": "shp_svc_test_secret_key_for_testing",
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/admin/api-keys":
			keyList := make([]map[string]interface{}, 0)
			for _, k := range keys {
				keyList = append(keyList, k)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"keys": keyList, "total": len(keyList)})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/admin/api-keys/key-123/revoke":
			if key, ok := keys["key-123"]; ok {
				key["revoked_at"] = "2025-01-15T12:00:00Z"
				keys["key-123"] = key
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "API key revoked"})

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	createResp, err := c.CreateAPIKey(context.Background(), CreateAPIKeyRequest{
		Name: "Test Key", Description: "For testing", Scopes: []string{"entities:read"},
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if createResp.RawKey == "" {
		t.Error("CREATE: RawKey should not be empty")
	}
	if createResp.Key.ID != "key-123" {
		t.Errorf("CREATE: ID = %q, want %q", createResp.Key.ID, "key-123")
	}

	// READ
	readKey, err := c.GetAPIKey(context.Background(), "key-123")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if readKey.Name != "Test Key" {
		t.Errorf("READ: Name = %q, want %q", readKey.Name, "Test Key")
	}

	// REVOKE
	err = c.RevokeAPIKey(context.Background(), "key-123")
	if err != nil {
		t.Fatalf("REVOKE failed: %v", err)
	}
}
