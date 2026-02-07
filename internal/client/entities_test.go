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

func TestListEntities_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"entities": []map[string]interface{}{
				{
					"service": map[string]interface{}{
						"id":   "svc-1",
						"name": "Service One",
						"type": "service",
						"tier": "tier1",
					},
					"description": "First service",
					"lifecycle":   "production",
				},
				{
					"service": map[string]interface{}{
						"id":   "svc-2",
						"name": "Service Two",
						"type": "library",
					},
					"description": "Second service",
					"lifecycle":   "experimental",
				},
			},
			"page": map[string]interface{}{
				"total": 2,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)
	entities, err := c.ListEntities(context.Background())
	if err != nil {
		t.Fatalf("ListEntities() error = %v", err)
	}

	if len(entities) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(entities))
	}
	if entities[0].Service.ID != "svc-1" {
		t.Errorf("entities[0].Service.ID = %q, want %q", entities[0].Service.ID, "svc-1")
	}
	if entities[0].Service.Name != "Service One" {
		t.Errorf("entities[0].Service.Name = %q, want %q", entities[0].Service.Name, "Service One")
	}
	if entities[1].Service.Type != "library" {
		t.Errorf("entities[1].Service.Type = %q, want %q", entities[1].Service.Type, "library")
	}
}

func TestGetEntity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/entities/my-service" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"entity": map[string]interface{}{
				"service": map[string]interface{}{
					"id":   "my-service",
					"name": "My Service",
					"type": "service",
					"tier": "tier1",
				},
				"description": "A test service",
				"lifecycle":   "production",
				"owner": []map[string]interface{}{
					{"type": "team", "id": "platform"},
				},
				"tags": []string{"go", "api"},
				"links": []map[string]interface{}{
					{"name": "Repo", "url": "https://github.com/example/repo", "icon": "github"},
				},
				"createdAt": "2025-01-15T10:30:00Z",
				"updatedAt": "2025-01-15T11:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)
	entity, err := c.GetEntity(context.Background(), "my-service")
	if err != nil {
		t.Fatalf("GetEntity() error = %v", err)
	}

	if entity.Service.ID != "my-service" {
		t.Errorf("Service.ID = %q, want %q", entity.Service.ID, "my-service")
	}
	if entity.Service.Name != "My Service" {
		t.Errorf("Service.Name = %q, want %q", entity.Service.Name, "My Service")
	}
	if entity.Service.Type != "service" {
		t.Errorf("Service.Type = %q, want %q", entity.Service.Type, "service")
	}
	if entity.Service.Tier != "tier1" {
		t.Errorf("Service.Tier = %q, want %q", entity.Service.Tier, "tier1")
	}
	if entity.Description != "A test service" {
		t.Errorf("Description = %q, want %q", entity.Description, "A test service")
	}
	if entity.Lifecycle != "production" {
		t.Errorf("Lifecycle = %q, want %q", entity.Lifecycle, "production")
	}
	if len(entity.Owner) != 1 {
		t.Fatalf("Owner count = %d, want 1", len(entity.Owner))
	}
	if entity.Owner[0].Type != "team" || entity.Owner[0].ID != "platform" {
		t.Errorf("Owner[0] = %+v, want {Type:team, ID:platform}", entity.Owner[0])
	}
	if len(entity.Tags) != 2 {
		t.Fatalf("Tags count = %d, want 2", len(entity.Tags))
	}
	if entity.Tags[0] != "go" || entity.Tags[1] != "api" {
		t.Errorf("Tags = %v, want [go api]", entity.Tags)
	}
	if len(entity.Links) != 1 {
		t.Fatalf("Links count = %d, want 1", len(entity.Links))
	}
	if entity.Links[0].Name != "Repo" {
		t.Errorf("Links[0].Name = %q, want %q", entity.Links[0].Name, "Repo")
	}
	if entity.Links[0].URL != "https://github.com/example/repo" {
		t.Errorf("Links[0].URL = %q, want %q", entity.Links[0].URL, "https://github.com/example/repo")
	}
	if entity.CreatedAt != "2025-01-15T10:30:00Z" {
		t.Errorf("CreatedAt = %q, want %q", entity.CreatedAt, "2025-01-15T10:30:00Z")
	}
	if entity.UpdatedAt != "2025-01-15T11:00:00Z" {
		t.Errorf("UpdatedAt = %q, want %q", entity.UpdatedAt, "2025-01-15T11:00:00Z")
	}
}

func TestGetEntity_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Entity not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)
	_, err := c.GetEntity(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found entity, got nil")
	}
}

func TestCreateEntity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/manifests/entities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateEntityRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		if req.Source != "terraform" {
			t.Errorf("Source = %q, want %q", req.Source, "terraform")
		}
		if req.Content == "" {
			t.Error("Content should not be empty")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"entity": map[string]interface{}{
				"id":          1,
				"serviceId":   "my-new-service",
				"name":        "My New Service",
				"type":        "service",
				"lifecycle":   "experimental",
				"description": "A new service",
				"source":      "terraform",
				"createdAt":   "2025-01-15T10:30:00Z",
				"updatedAt":   "2025-01-15T10:30:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)

	manifestYAML := `apiVersion: shoehorn.io/v1alpha1
kind: Entity
metadata:
  name: my-new-service
spec:
  type: service
  lifecycle: experimental
  description: A new service`

	resp, err := c.CreateEntity(context.Background(), CreateEntityRequest{
		Content: manifestYAML,
		Source:  "terraform",
	})
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}

	if !resp.Success {
		t.Error("Success = false, want true")
	}
	if resp.Entity.ServiceID != "my-new-service" {
		t.Errorf("Entity.ServiceID = %q, want %q", resp.Entity.ServiceID, "my-new-service")
	}
	if resp.Entity.Name != "My New Service" {
		t.Errorf("Entity.Name = %q, want %q", resp.Entity.Name, "My New Service")
	}
	if resp.Entity.Type != "service" {
		t.Errorf("Entity.Type = %q, want %q", resp.Entity.Type, "service")
	}
	if resp.Entity.Lifecycle != "experimental" {
		t.Errorf("Entity.Lifecycle = %q, want %q", resp.Entity.Lifecycle, "experimental")
	}
	if resp.Entity.Source != "terraform" {
		t.Errorf("Entity.Source = %q, want %q", resp.Entity.Source, "terraform")
	}
}

func TestUpdateEntity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/manifests/entities/my-service" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateEntityRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		if req.Source != "terraform" {
			t.Errorf("Source = %q, want %q", req.Source, "terraform")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"entity": map[string]interface{}{
				"id":          1,
				"serviceId":   "my-service",
				"name":        "My Service Updated",
				"type":        "service",
				"lifecycle":   "production",
				"description": "Updated description",
				"source":      "terraform",
				"createdAt":   "2025-01-15T10:30:00Z",
				"updatedAt":   "2025-01-15T12:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)

	manifestYAML := `apiVersion: shoehorn.io/v1alpha1
kind: Entity
metadata:
  name: my-service
spec:
  type: service
  lifecycle: production
  description: Updated description`

	resp, err := c.UpdateEntity(context.Background(), "my-service", CreateEntityRequest{
		Content: manifestYAML,
		Source:  "terraform",
	})
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v", err)
	}

	if !resp.Success {
		t.Error("Success = false, want true")
	}
	if resp.Entity.ServiceID != "my-service" {
		t.Errorf("Entity.ServiceID = %q, want %q", resp.Entity.ServiceID, "my-service")
	}
	if resp.Entity.Description != "Updated description" {
		t.Errorf("Entity.Description = %q, want %q", resp.Entity.Description, "Updated description")
	}
	if resp.Entity.UpdatedAt != "2025-01-15T12:00:00Z" {
		t.Errorf("Entity.UpdatedAt = %q, want %q", resp.Entity.UpdatedAt, "2025-01-15T12:00:00Z")
	}
}

func TestDeleteEntity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/manifests/entities/my-service" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)
	err := c.DeleteEntity(context.Background(), "my-service")
	if err != nil {
		t.Fatalf("DeleteEntity() error = %v", err)
	}
}

func TestDeleteEntity_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Entity not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)
	err := c.DeleteEntity(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found entity, got nil")
	}
}

func TestEntityClient_CRUD_Integration(t *testing.T) {
	entities := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/manifests/entities":
			body, _ := io.ReadAll(r.Body)
			var req CreateEntityRequest
			json.Unmarshal(body, &req)

			entities["my-service"] = map[string]interface{}{
				"service": map[string]interface{}{
					"id":   "my-service",
					"name": "My Service",
					"type": "service",
				},
				"description": "A test service",
				"lifecycle":   "experimental",
				"createdAt":   "2025-01-15T10:30:00Z",
				"updatedAt":   "2025-01-15T10:30:00Z",
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"entity": map[string]interface{}{
					"id":          1,
					"serviceId":   "my-service",
					"name":        "My Service",
					"type":        "service",
					"lifecycle":   "experimental",
					"description": "A test service",
					"source":      req.Source,
					"createdAt":   "2025-01-15T10:30:00Z",
					"updatedAt":   "2025-01-15T10:30:00Z",
				},
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/entities/my-service":
			if entity, ok := entities["my-service"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"entity": entity})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "Entity not found"})
			}

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/manifests/entities/my-service":
			entity := entities["my-service"]
			entity["description"] = "Updated description"
			entity["lifecycle"] = "production"
			entity["updatedAt"] = "2025-01-15T12:00:00Z"
			entities["my-service"] = entity

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"entity": map[string]interface{}{
					"id":          1,
					"serviceId":   "my-service",
					"name":        "My Service",
					"type":        "service",
					"lifecycle":   "production",
					"description": "Updated description",
					"source":      "terraform",
					"createdAt":   "2025-01-15T10:30:00Z",
					"updatedAt":   "2025-01-15T12:00:00Z",
				},
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/manifests/entities/my-service":
			delete(entities, "my-service")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", 30*time.Second)

	// CREATE
	createResp, err := c.CreateEntity(context.Background(), CreateEntityRequest{
		Content: "apiVersion: shoehorn.io/v1alpha1\nkind: Entity\nmetadata:\n  name: my-service",
		Source:  "terraform",
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if !createResp.Success {
		t.Error("CREATE: Success = false, want true")
	}
	if createResp.Entity.ServiceID != "my-service" {
		t.Errorf("CREATE: ServiceID = %q, want %q", createResp.Entity.ServiceID, "my-service")
	}

	// READ
	readEntity, err := c.GetEntity(context.Background(), "my-service")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if readEntity.Service.ID != "my-service" {
		t.Errorf("READ: Service.ID = %q, want %q", readEntity.Service.ID, "my-service")
	}
	if readEntity.Description != "A test service" {
		t.Errorf("READ: Description = %q, want %q", readEntity.Description, "A test service")
	}

	// UPDATE
	updateResp, err := c.UpdateEntity(context.Background(), "my-service", CreateEntityRequest{
		Content: "apiVersion: shoehorn.io/v1alpha1\nkind: Entity\nmetadata:\n  name: my-service\nspec:\n  lifecycle: production",
		Source:  "terraform",
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updateResp.Entity.Lifecycle != "production" {
		t.Errorf("UPDATE: Lifecycle = %q, want %q", updateResp.Entity.Lifecycle, "production")
	}

	// DELETE
	err = c.DeleteEntity(context.Background(), "my-service")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// Verify deleted
	_, err = c.GetEntity(context.Background(), "my-service")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
