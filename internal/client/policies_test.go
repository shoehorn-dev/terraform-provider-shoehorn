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

func TestListPolicies_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/policies" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policies": []map[string]interface{}{
				{
					"id": "pol-1", "key": "tenant-isolation", "name": "Tenant Isolation",
					"category": "security", "enabled": true, "enforcement": "block", "system": true,
				},
				{
					"id": "pol-2", "key": "entity-metadata-validation", "name": "Entity Metadata Validation",
					"category": "governance", "enabled": false, "enforcement": "warn", "system": false,
				},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	policies, err := c.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListPolicies() error = %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("policy count = %d, want 2", len(policies))
	}
	if policies[0].Key != "tenant-isolation" {
		t.Errorf("Key = %q, want %q", policies[0].Key, "tenant-isolation")
	}
	if !policies[0].System {
		t.Error("System = false, want true")
	}
}

func TestGetPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policies": []map[string]interface{}{
				{"id": "pol-1", "key": "tenant-isolation", "name": "Tenant Isolation", "enabled": true, "enforcement": "block"},
				{"id": "pol-2", "key": "entity-metadata-validation", "name": "Entity Metadata Validation", "enabled": false, "enforcement": "warn"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	policy, err := c.GetPolicy(context.Background(), "entity-metadata-validation")
	if err != nil {
		t.Fatalf("GetPolicy() error = %v", err)
	}
	if policy.ID != "pol-2" {
		t.Errorf("ID = %q, want %q", policy.ID, "pol-2")
	}
	if policy.Name != "Entity Metadata Validation" {
		t.Errorf("Name = %q, want %q", policy.Name, "Entity Metadata Validation")
	}
}

func TestGetPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"policies": []interface{}{}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found policy, got nil")
	}
}

func TestUpdatePolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/admin/policies/pol-2" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req UpdatePolicyRequest
		json.Unmarshal(body, &req)

		if req.Enabled == nil || !*req.Enabled {
			t.Error("Enabled should be true")
		}
		if req.Enforcement != "block" {
			t.Errorf("Enforcement = %q, want %q", req.Enforcement, "block")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "pol-2", "key": "entity-metadata-validation",
			"name": "Entity Metadata Validation", "category": "governance",
			"enabled": true, "enforcement": "block",
			"updated_at": "2025-01-15T11:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	enabled := true
	policy, err := c.UpdatePolicy(context.Background(), "pol-2", UpdatePolicyRequest{
		Enabled:     &enabled,
		Enforcement: "block",
	})
	if err != nil {
		t.Fatalf("UpdatePolicy() error = %v", err)
	}
	if !policy.Enabled {
		t.Error("Enabled = false, want true")
	}
	if policy.Enforcement != "block" {
		t.Errorf("Enforcement = %q, want %q", policy.Enforcement, "block")
	}
}
