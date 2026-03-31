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

func TestListForgeMolds_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/forge/molds" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"molds": []map[string]interface{}{
				{
					"id": "mold-1", "slug": "k8s-deploy", "name": "Kubernetes Deploy",
					"version": "1.0.0", "visibility": "public", "category": "deployment",
					"published": true,
				},
				{
					"id": "mold-2", "slug": "aws-lambda", "name": "AWS Lambda",
					"version": "2.1.0", "visibility": "tenant", "category": "serverless",
					"published": false,
				},
			},
			"pagination": map[string]interface{}{"total": 2, "page": 1},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	molds, err := c.ListForgeMolds(context.Background())
	if err != nil {
		t.Fatalf("ListForgeMolds() error = %v", err)
	}
	if len(molds) != 2 {
		t.Fatalf("mold count = %d, want 2", len(molds))
	}
	if molds[0].Slug != "k8s-deploy" {
		t.Errorf("Slug = %q, want %q", molds[0].Slug, "k8s-deploy")
	}
	if molds[0].Name != "Kubernetes Deploy" {
		t.Errorf("Name = %q, want %q", molds[0].Name, "Kubernetes Deploy")
	}
	if molds[0].Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", molds[0].Version, "1.0.0")
	}
	if molds[0].Visibility != "public" {
		t.Errorf("Visibility = %q, want %q", molds[0].Visibility, "public")
	}
	if molds[0].Category != "deployment" {
		t.Errorf("Category = %q, want %q", molds[0].Category, "deployment")
	}
	if molds[1].Slug != "aws-lambda" {
		t.Errorf("Slug = %q, want %q", molds[1].Slug, "aws-lambda")
	}
}

func TestGetForgeMold_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/forge/molds/k8s-deploy" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mold": map[string]interface{}{
				"id": "mold-1", "slug": "k8s-deploy", "name": "Kubernetes Deploy",
				"description": "Deploy to Kubernetes clusters",
				"version": "1.0.0", "visibility": "public", "category": "deployment",
				"tags":      []string{"kubernetes", "deploy"},
				"icon":      "k8s-icon",
				"published": true,
				"actions": []map[string]interface{}{
					{"action": "deploy", "label": "Deploy", "description": "Deploy to cluster", "primary": true},
					{"action": "rollback", "label": "Rollback"},
				},
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"namespace": map[string]interface{}{"type": "string"},
					},
				},
				"defaults": map[string]interface{}{
					"namespace": "default",
				},
				"created_at": "2025-01-15T10:00:00Z",
				"updated_at": "2025-01-15T11:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	mold, err := c.GetForgeMold(context.Background(), "k8s-deploy")
	if err != nil {
		t.Fatalf("GetForgeMold() error = %v", err)
	}
	if mold.ID != "mold-1" {
		t.Errorf("ID = %q, want %q", mold.ID, "mold-1")
	}
	if mold.Slug != "k8s-deploy" {
		t.Errorf("Slug = %q, want %q", mold.Slug, "k8s-deploy")
	}
	if mold.Name != "Kubernetes Deploy" {
		t.Errorf("Name = %q, want %q", mold.Name, "Kubernetes Deploy")
	}
	if mold.Description != "Deploy to Kubernetes clusters" {
		t.Errorf("Description = %q, want %q", mold.Description, "Deploy to Kubernetes clusters")
	}
	if mold.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", mold.Version, "1.0.0")
	}
	if mold.Icon != "k8s-icon" {
		t.Errorf("Icon = %q, want %q", mold.Icon, "k8s-icon")
	}
	if !mold.Published {
		t.Error("Published = false, want true")
	}
	if len(mold.Tags) != 2 {
		t.Fatalf("Tags count = %d, want 2", len(mold.Tags))
	}
	if mold.Tags[0] != "kubernetes" {
		t.Errorf("Tags[0] = %q, want %q", mold.Tags[0], "kubernetes")
	}
	if len(mold.Actions) != 2 {
		t.Fatalf("Actions count = %d, want 2", len(mold.Actions))
	}
	if mold.Actions[0].Action != "deploy" {
		t.Errorf("Actions[0].Action = %q, want %q", mold.Actions[0].Action, "deploy")
	}
	if mold.Actions[0].Label != "Deploy" {
		t.Errorf("Actions[0].Label = %q, want %q", mold.Actions[0].Label, "Deploy")
	}
	if !mold.Actions[0].Primary {
		t.Error("Actions[0].Primary = false, want true")
	}
	if mold.Schema == nil {
		t.Error("Schema should not be nil")
	}
	if mold.Defaults == nil {
		t.Error("Defaults should not be nil")
	}
	if mold.Defaults["namespace"] != "default" {
		t.Errorf("Defaults[namespace] = %v, want %q", mold.Defaults["namespace"], "default")
	}
}

func TestGetForgeMold_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "mold not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetForgeMold(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetForgeMold() expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestCreateForgeMold_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/forge/molds" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateForgeMoldRequest
		json.Unmarshal(body, &req)

		if req.Slug != "k8s-deploy" {
			t.Errorf("Slug = %q, want %q", req.Slug, "k8s-deploy")
		}
		if req.Name != "Kubernetes Deploy" {
			t.Errorf("Name = %q, want %q", req.Name, "Kubernetes Deploy")
		}
		if req.Version != "1.0.0" {
			t.Errorf("Version = %q, want %q", req.Version, "1.0.0")
		}
		if req.Visibility != "public" {
			t.Errorf("Visibility = %q, want %q", req.Visibility, "public")
		}
		if req.Category != "deployment" {
			t.Errorf("Category = %q, want %q", req.Category, "deployment")
		}
		if len(req.Actions) != 1 {
			t.Fatalf("Actions count = %d, want 1", len(req.Actions))
		}
		if req.Actions[0].Action != "deploy" {
			t.Errorf("Actions[0].Action = %q, want %q", req.Actions[0].Action, "deploy")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mold": map[string]interface{}{
				"id": "mold-1", "slug": req.Slug, "name": req.Name,
				"description": req.Description,
				"version": req.Version, "visibility": req.Visibility,
				"category": req.Category, "published": false,
				"actions": []map[string]interface{}{
					{"action": "deploy", "label": "Deploy", "primary": true},
				},
				"created_at": "2025-01-15T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	mold, err := c.CreateForgeMold(context.Background(), CreateForgeMoldRequest{
		Slug:        "k8s-deploy",
		Name:        "Kubernetes Deploy",
		Description: "Deploy to Kubernetes clusters",
		Version:     "1.0.0",
		Visibility:  "public",
		Category:    "deployment",
		Actions: []ForgeMoldAction{
			{Action: "deploy", Label: "Deploy", Primary: true},
		},
	})
	if err != nil {
		t.Fatalf("CreateForgeMold() error = %v", err)
	}
	if mold.ID != "mold-1" {
		t.Errorf("ID = %q, want %q", mold.ID, "mold-1")
	}
	if mold.Slug != "k8s-deploy" {
		t.Errorf("Slug = %q, want %q", mold.Slug, "k8s-deploy")
	}
	if mold.Name != "Kubernetes Deploy" {
		t.Errorf("Name = %q, want %q", mold.Name, "Kubernetes Deploy")
	}
	if !mold.Actions[0].Primary {
		t.Error("Actions[0].Primary = false, want true")
	}
}

func TestUpdateForgeMold_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/forge/molds/k8s-deploy" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req UpdateForgeMoldRequest
		json.Unmarshal(body, &req)

		if req.Version != "1.1.0" {
			t.Errorf("Version = %q, want %q", req.Version, "1.1.0")
		}
		if req.Name != "Kubernetes Deploy v2" {
			t.Errorf("Name = %q, want %q", req.Name, "Kubernetes Deploy v2")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mold": map[string]interface{}{
				"id": "mold-1", "slug": "k8s-deploy", "name": "Kubernetes Deploy v2",
				"version": "1.1.0", "visibility": "public", "category": "deployment",
				"published":  false,
				"updated_at": "2025-01-15T12:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	mold, err := c.UpdateForgeMold(context.Background(), "k8s-deploy", UpdateForgeMoldRequest{
		Version: "1.1.0",
		Name:    "Kubernetes Deploy v2",
	})
	if err != nil {
		t.Fatalf("UpdateForgeMold() error = %v", err)
	}
	if mold.Name != "Kubernetes Deploy v2" {
		t.Errorf("Name = %q, want %q", mold.Name, "Kubernetes Deploy v2")
	}
	if mold.Version != "1.1.0" {
		t.Errorf("Version = %q, want %q", mold.Version, "1.1.0")
	}
}

func TestDeleteForgeMold_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/forge/molds/k8s-deploy" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("version") != "1.0.0" {
			t.Errorf("version query param = %q, want %q", r.URL.Query().Get("version"), "1.0.0")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteForgeMold(context.Background(), "k8s-deploy", "1.0.0")
	if err != nil {
		t.Fatalf("DeleteForgeMold() error = %v", err)
	}
}

func TestPublishForgeMold_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/forge/molds/k8s-deploy/publish" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)

		if req["version"] != "1.0.0" {
			t.Errorf("version = %q, want %q", req["version"], "1.0.0")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mold": map[string]interface{}{
				"id": "mold-1", "slug": "k8s-deploy", "name": "Kubernetes Deploy",
				"version": "1.0.0", "visibility": "public", "category": "deployment",
				"published":  true,
				"updated_at": "2025-01-15T13:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	mold, err := c.PublishForgeMold(context.Background(), "k8s-deploy", "1.0.0")
	if err != nil {
		t.Fatalf("PublishForgeMold() error = %v", err)
	}
	if !mold.Published {
		t.Error("Published = false, want true")
	}
	if mold.Slug != "k8s-deploy" {
		t.Errorf("Slug = %q, want %q", mold.Slug, "k8s-deploy")
	}
}

func TestListForgeMolds_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"molds":      []map[string]interface{}{},
			"pagination": map[string]interface{}{"total": 0},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	molds, err := c.ListForgeMolds(context.Background())
	if err != nil {
		t.Fatalf("ListForgeMolds() error = %v", err)
	}
	if len(molds) != 0 {
		t.Errorf("mold count = %d, want 0", len(molds))
	}
}

func TestCreateForgeMold_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"code":"INTERNAL","message":"internal server error"}`)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.CreateForgeMold(context.Background(), CreateForgeMoldRequest{
		Slug:       "test-mold",
		Name:       "Test Mold",
		Version:    "1.0.0",
		Visibility: "public",
		Category:   "deployment",
		Actions: []ForgeMoldAction{
			{Action: "deploy", Label: "Deploy", Primary: true},
		},
	})
	if err == nil {
		t.Fatal("CreateForgeMold() expected error for 500 response, got nil")
	}
}

func TestDeleteForgeMold_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "NOT_FOUND", "message": "mold not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteForgeMold(context.Background(), "nonexistent", "1.0.0")
	if err == nil {
		t.Fatal("DeleteForgeMold() expected error for 404, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestForgeMold_Lifecycle(t *testing.T) {
	molds := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// CREATE
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/forge/molds":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			slug := req["slug"].(string)
			mold := map[string]interface{}{
				"id": "mold-1", "slug": slug, "name": req["name"],
				"description": req["description"], "version": req["version"],
				"visibility": req["visibility"], "category": req["category"],
				"actions": req["actions"], "published": false,
				"created_at": "2025-01-15T10:00:00Z",
			}
			if tags, ok := req["tags"]; ok {
				mold["tags"] = tags
			}
			if schema, ok := req["schema"]; ok {
				mold["schema"] = schema
			}
			if defaults, ok := req["defaults"]; ok {
				mold["defaults"] = defaults
			}
			molds[slug] = mold

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"mold": mold})

		// PUBLISH
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/forge/molds/k8s-deploy/publish":
			if m, ok := molds["k8s-deploy"]; ok {
				m["published"] = true
				m["updated_at"] = "2025-01-15T12:00:00Z"
				molds["k8s-deploy"] = m
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"mold": m})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "mold not found"})
			}

		// GET
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/forge/molds/k8s-deploy":
			if m, ok := molds["k8s-deploy"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"mold": m})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "mold not found"})
			}

		// LIST
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/forge/molds":
			moldList := make([]map[string]interface{}, 0, len(molds))
			for _, m := range molds {
				moldList = append(moldList, m)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"molds":      moldList,
				"pagination": map[string]interface{}{"total": len(moldList)},
			})

		// UPDATE
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/forge/molds/k8s-deploy":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			if m, ok := molds["k8s-deploy"]; ok {
				if name, ok := req["name"].(string); ok && name != "" {
					m["name"] = name
				}
				if version, ok := req["version"].(string); ok && version != "" {
					m["version"] = version
				}
				if desc, ok := req["description"].(string); ok {
					m["description"] = desc
				}
				m["updated_at"] = "2025-01-15T11:00:00Z"
				molds["k8s-deploy"] = m
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"mold": m})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "mold not found"})
			}

		// DELETE
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/forge/molds/k8s-deploy":
			delete(molds, "k8s-deploy")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"not found"}`)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	created, err := c.CreateForgeMold(context.Background(), CreateForgeMoldRequest{
		Slug:        "k8s-deploy",
		Name:        "Kubernetes Deploy",
		Description: "Deploy to K8s",
		Version:     "1.0.0",
		Visibility:  "public",
		Category:    "deployment",
		Tags:        []string{"kubernetes"},
		Actions: []ForgeMoldAction{
			{Action: "deploy", Label: "Deploy", Primary: true},
		},
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if created.ID != "mold-1" {
		t.Errorf("CREATE: ID = %q, want %q", created.ID, "mold-1")
	}
	if created.Slug != "k8s-deploy" {
		t.Errorf("CREATE: Slug = %q, want %q", created.Slug, "k8s-deploy")
	}
	if created.Published {
		t.Error("CREATE: Published = true, want false")
	}

	// READ
	read, err := c.GetForgeMold(context.Background(), "k8s-deploy")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if read.Name != "Kubernetes Deploy" {
		t.Errorf("READ: Name = %q, want %q", read.Name, "Kubernetes Deploy")
	}

	// LIST
	list, err := c.ListForgeMolds(context.Background())
	if err != nil {
		t.Fatalf("LIST failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("LIST: count = %d, want 1", len(list))
	}

	// UPDATE
	updated, err := c.UpdateForgeMold(context.Background(), "k8s-deploy", UpdateForgeMoldRequest{
		Version: "1.1.0",
		Name:    "Kubernetes Deploy v2",
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updated.Name != "Kubernetes Deploy v2" {
		t.Errorf("UPDATE: Name = %q, want %q", updated.Name, "Kubernetes Deploy v2")
	}
	if updated.Version != "1.1.0" {
		t.Errorf("UPDATE: Version = %q, want %q", updated.Version, "1.1.0")
	}

	// PUBLISH
	published, err := c.PublishForgeMold(context.Background(), "k8s-deploy", "1.1.0")
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	if !published.Published {
		t.Error("PUBLISH: Published = false, want true")
	}

	// DELETE
	err = c.DeleteForgeMold(context.Background(), "k8s-deploy", "1.1.0")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// Verify deleted - GET should now fail
	_, err = c.GetForgeMold(context.Background(), "k8s-deploy")
	if err == nil {
		t.Fatal("GET after DELETE: expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("GET after DELETE: expected not found error, got: %v", err)
	}
}
