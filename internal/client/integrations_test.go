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

func TestListIntegrations_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/integrations/configs" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integrations": []map[string]interface{}{
				{"id": 1, "name": "GitHub Prod", "type": "github", "status": "active"},
				{"id": 2, "name": "Slack Alerts", "type": "slack", "status": "active"},
			},
			"count": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integrations, err := c.ListIntegrations(context.Background())
	if err != nil {
		t.Fatalf("ListIntegrations() error = %v", err)
	}
	if len(integrations) != 2 {
		t.Errorf("integration count = %d, want 2", len(integrations))
	}
	if integrations[0].Name != "GitHub Prod" {
		t.Errorf("Name = %q, want %q", integrations[0].Name, "GitHub Prod")
	}
	if integrations[0].Type != "github" {
		t.Errorf("Type = %q, want %q", integrations[0].Type, "github")
	}
}

func TestGetIntegration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/integrations/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integration": map[string]interface{}{
				"id": 1, "name": "GitHub Prod", "type": "github", "status": "active",
				"created_at": "2025-01-15T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integration, err := c.GetIntegration(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetIntegration() error = %v", err)
	}
	if integration.Name != "GitHub Prod" {
		t.Errorf("Name = %q, want %q", integration.Name, "GitHub Prod")
	}
	if integration.Status != "active" {
		t.Errorf("Status = %q, want %q", integration.Status, "active")
	}
}

func TestCreateIntegration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/integrations" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateIntegrationRequest
		json.Unmarshal(body, &req)

		if req.Name != "GitHub Prod" {
			t.Errorf("Name = %q, want %q", req.Name, "GitHub Prod")
		}
		if req.Type != "github" {
			t.Errorf("Type = %q, want %q", req.Type, "github")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integration": map[string]interface{}{
				"id": 1, "name": req.Name, "type": req.Type,
				"status": "pending", "config": req.Config,
				"created_at": "2025-01-15T10:00:00Z",
			},
			"message": "Integration created successfully",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integration, err := c.CreateIntegration(context.Background(), CreateIntegrationRequest{
		Name: "GitHub Prod",
		Type: "github",
		Config: map[string]interface{}{
			"token":        "ghp_xxxx",
			"organization": "myorg",
		},
	})
	if err != nil {
		t.Fatalf("CreateIntegration() error = %v", err)
	}
	if integration.ID != 1 {
		t.Errorf("ID = %d, want 1", integration.ID)
	}
	if integration.Name != "GitHub Prod" {
		t.Errorf("Name = %q, want %q", integration.Name, "GitHub Prod")
	}
}

func TestUpdateIntegration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/integrations/1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integration": map[string]interface{}{
				"id": 1, "name": "GitHub Updated", "type": "github", "status": "active",
				"updated_at": "2025-01-15T11:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integration, err := c.UpdateIntegration(context.Background(), 1, UpdateIntegrationRequest{
		Name: "GitHub Updated",
	})
	if err != nil {
		t.Fatalf("UpdateIntegration() error = %v", err)
	}
	if integration.Name != "GitHub Updated" {
		t.Errorf("Name = %q, want %q", integration.Name, "GitHub Updated")
	}
}

func TestDeleteIntegration_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/integrations/1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Integration deleted successfully"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteIntegration(context.Background(), 1)
	if err != nil {
		t.Fatalf("DeleteIntegration() error = %v", err)
	}
}

func TestGetIntegrationsStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/integrations" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integrations": []map[string]interface{}{
				{"type": "orgdata", "provider": "zitadel", "status": "connected", "config": map[string]interface{}{"url": "https://zitadel.example.com"}, "metadata": map[string]interface{}{"version": "2.0"}},
				{"type": "github", "provider": "github", "status": "connected", "config": map[string]interface{}{"org": "myorg"}, "metadata": map[string]interface{}{"repos": 42}},
				{"type": "authentication", "provider": "zitadel", "status": "connected"},
			},
			"total":        3,
			"healthy":      3,
			"last_updated": "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integrations, total, healthy, err := c.GetIntegrationsStatus(context.Background())
	if err != nil {
		t.Fatalf("GetIntegrationsStatus() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if healthy != 3 {
		t.Errorf("healthy = %d, want 3", healthy)
	}
	if len(integrations) != 3 {
		t.Fatalf("integration count = %d, want 3", len(integrations))
	}
	if integrations[0].Type != "orgdata" {
		t.Errorf("integrations[0].Type = %q, want %q", integrations[0].Type, "orgdata")
	}
	if integrations[0].Provider != "zitadel" {
		t.Errorf("integrations[0].Provider = %q, want %q", integrations[0].Provider, "zitadel")
	}
	if integrations[0].Status != "connected" {
		t.Errorf("integrations[0].Status = %q, want %q", integrations[0].Status, "connected")
	}
	if integrations[1].Type != "github" {
		t.Errorf("integrations[1].Type = %q, want %q", integrations[1].Type, "github")
	}
}

func TestGetIntegrationsStatus_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integrations": []map[string]interface{}{},
			"total":        0,
			"healthy":      0,
			"last_updated": "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integrations, total, healthy, err := c.GetIntegrationsStatus(context.Background())
	if err != nil {
		t.Fatalf("GetIntegrationsStatus() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if healthy != 0 {
		t.Errorf("healthy = %d, want 0", healthy)
	}
	if len(integrations) != 0 {
		t.Errorf("integration count = %d, want 0", len(integrations))
	}
}

func TestGetIntegrationsStatus_WithConfigAndMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"integrations": []map[string]interface{}{
				{
					"type":     "github",
					"provider": "github",
					"status":   "connected",
					"config":   map[string]interface{}{"org": "myorg", "token_set": true},
					"metadata": map[string]interface{}{"repos_count": 42, "last_sync": "2025-01-15T10:00:00Z"},
				},
			},
			"total":   1,
			"healthy": 1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	integrations, _, _, err := c.GetIntegrationsStatus(context.Background())
	if err != nil {
		t.Fatalf("GetIntegrationsStatus() error = %v", err)
	}
	if len(integrations) != 1 {
		t.Fatalf("integration count = %d, want 1", len(integrations))
	}
	if integrations[0].Config == nil {
		t.Error("Config should not be nil")
	}
	if integrations[0].Config["org"] != "myorg" {
		t.Errorf("Config[org] = %v, want %q", integrations[0].Config["org"], "myorg")
	}
	if integrations[0].Metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestIntegrationClient_Lifecycle_Integration(t *testing.T) {
	integrations := make(map[int]map[string]interface{})
	nextID := 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/integrations":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			integration := map[string]interface{}{
				"id": nextID, "name": req["name"], "type": req["type"],
				"status": "pending", "config": req["config"],
				"created_at": "2025-01-15T10:00:00Z",
			}
			integrations[nextID] = integration
			nextID++

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"integration": integration})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/integrations/1":
			if i, ok := integrations[1]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"integration": i})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": "NOT_FOUND"}})
			}

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/integrations/1":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			if i, ok := integrations[1]; ok {
				if name, ok := req["name"].(string); ok {
					i["name"] = name
				}
				i["updated_at"] = "2025-01-15T11:00:00Z"
				integrations[1] = i
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"integration": i})
			}

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/integrations/1":
			delete(integrations, 1)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"message": "deleted"})

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error":{"code":"NOT_FOUND"}}`)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	created, err := c.CreateIntegration(context.Background(), CreateIntegrationRequest{
		Name: "Test GitHub", Type: "github",
		Config: map[string]interface{}{"token": "ghp_test", "organization": "testorg"},
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("CREATE: ID = %d, want 1", created.ID)
	}

	// READ
	read, err := c.GetIntegration(context.Background(), 1)
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if read.Name != "Test GitHub" {
		t.Errorf("READ: Name = %q, want %q", read.Name, "Test GitHub")
	}

	// UPDATE
	updated, err := c.UpdateIntegration(context.Background(), 1, UpdateIntegrationRequest{Name: "Updated GitHub"})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updated.Name != "Updated GitHub" {
		t.Errorf("UPDATE: Name = %q, want %q", updated.Name, "Updated GitHub")
	}

	// DELETE
	err = c.DeleteIntegration(context.Background(), 1)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
}
