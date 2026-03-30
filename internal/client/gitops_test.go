package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListGitOpsResources_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/operations/gitops" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"id": "res-1", "cluster_id": "prod-east", "tool": "argocd",
					"namespace": "argocd", "name": "my-app", "kind": "Application",
					"sync_status": "Synced", "health_status": "Healthy",
					"source_url": "https://github.com/org/repo", "revision": "abc123",
					"target_revision": "main", "auto_sync": true, "suspended": false,
					"entity_id": "ent-1", "entity_name": "my-service", "owner_team": "platform",
					"last_synced_at": "2025-06-01T10:00:00Z",
					"created_at":     "2025-01-15T10:00:00Z", "updated_at": "2025-06-01T10:00:00Z",
				},
				{
					"id": "res-2", "cluster_id": "prod-east", "tool": "fluxcd",
					"namespace": "flux-system", "name": "helm-release-1", "kind": "HelmRelease",
					"sync_status": "OutOfSync", "health_status": "Progressing",
					"chart_name": "my-chart", "chart_version": "1.2.3",
					"auto_sync": false, "suspended": false,
					"created_at": "2025-02-01T10:00:00Z", "updated_at": "2025-06-01T10:00:00Z",
				},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resources, total, err := c.ListGitOpsResources(context.Background(), ListGitOpsResourcesParams{})
	if err != nil {
		t.Fatalf("ListGitOpsResources() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(resources) != 2 {
		t.Fatalf("resource count = %d, want 2", len(resources))
	}

	// Verify first resource (ArgoCD Application)
	if resources[0].ID != "res-1" {
		t.Errorf("resources[0].ID = %q, want %q", resources[0].ID, "res-1")
	}
	if resources[0].ClusterID != "prod-east" {
		t.Errorf("resources[0].ClusterID = %q, want %q", resources[0].ClusterID, "prod-east")
	}
	if resources[0].Tool != "argocd" {
		t.Errorf("resources[0].Tool = %q, want %q", resources[0].Tool, "argocd")
	}
	if resources[0].Name != "my-app" {
		t.Errorf("resources[0].Name = %q, want %q", resources[0].Name, "my-app")
	}
	if resources[0].Kind != "Application" {
		t.Errorf("resources[0].Kind = %q, want %q", resources[0].Kind, "Application")
	}
	if resources[0].SyncStatus != "Synced" {
		t.Errorf("resources[0].SyncStatus = %q, want %q", resources[0].SyncStatus, "Synced")
	}
	if resources[0].HealthStatus != "Healthy" {
		t.Errorf("resources[0].HealthStatus = %q, want %q", resources[0].HealthStatus, "Healthy")
	}
	if resources[0].SourceURL != "https://github.com/org/repo" {
		t.Errorf("resources[0].SourceURL = %q, want %q", resources[0].SourceURL, "https://github.com/org/repo")
	}
	if !resources[0].AutoSync {
		t.Error("resources[0].AutoSync = false, want true")
	}
	if resources[0].EntityName != "my-service" {
		t.Errorf("resources[0].EntityName = %q, want %q", resources[0].EntityName, "my-service")
	}
	if resources[0].OwnerTeam != "platform" {
		t.Errorf("resources[0].OwnerTeam = %q, want %q", resources[0].OwnerTeam, "platform")
	}

	// Verify second resource (FluxCD HelmRelease)
	if resources[1].Tool != "fluxcd" {
		t.Errorf("resources[1].Tool = %q, want %q", resources[1].Tool, "fluxcd")
	}
	if resources[1].Kind != "HelmRelease" {
		t.Errorf("resources[1].Kind = %q, want %q", resources[1].Kind, "HelmRelease")
	}
	if resources[1].ChartName != "my-chart" {
		t.Errorf("resources[1].ChartName = %q, want %q", resources[1].ChartName, "my-chart")
	}
	if resources[1].ChartVersion != "1.2.3" {
		t.Errorf("resources[1].ChartVersion = %q, want %q", resources[1].ChartVersion, "1.2.3")
	}
}

func TestListGitOpsResources_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/operations/gitops" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if v := query.Get("cluster_id"); v != "prod-east" {
			t.Errorf("cluster_id = %q, want %q", v, "prod-east")
		}
		if v := query.Get("tool"); v != "argocd" {
			t.Errorf("tool = %q, want %q", v, "argocd")
		}
		if v := query.Get("sync_status"); v != "OutOfSync" {
			t.Errorf("sync_status = %q, want %q", v, "OutOfSync")
		}
		if v := query.Get("health_status"); v != "Degraded" {
			t.Errorf("health_status = %q, want %q", v, "Degraded")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"id": "res-3", "cluster_id": "prod-east", "tool": "argocd",
					"namespace": "argocd", "name": "failing-app", "kind": "Application",
					"sync_status": "OutOfSync", "health_status": "Degraded",
					"auto_sync": true, "suspended": false,
				},
			},
			"total": 1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resources, total, err := c.ListGitOpsResources(context.Background(), ListGitOpsResourcesParams{
		ClusterID:    "prod-east",
		Tool:         "argocd",
		SyncStatus:   "OutOfSync",
		HealthStatus: "Degraded",
	})
	if err != nil {
		t.Fatalf("ListGitOpsResources() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(resources) != 1 {
		t.Fatalf("resource count = %d, want 1", len(resources))
	}
	if resources[0].ID != "res-3" {
		t.Errorf("resources[0].ID = %q, want %q", resources[0].ID, "res-3")
	}
	if resources[0].SyncStatus != "OutOfSync" {
		t.Errorf("resources[0].SyncStatus = %q, want %q", resources[0].SyncStatus, "OutOfSync")
	}
	if resources[0].HealthStatus != "Degraded" {
		t.Errorf("resources[0].HealthStatus = %q, want %q", resources[0].HealthStatus, "Degraded")
	}
}

func TestListGitOpsResources_PartialFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Only cluster_id should be set
		if v := query.Get("cluster_id"); v != "staging" {
			t.Errorf("cluster_id = %q, want %q", v, "staging")
		}
		// Empty filters should not appear as query params
		if query.Has("tool") {
			t.Errorf("tool param should not be present, got %q", query.Get("tool"))
		}
		if query.Has("sync_status") {
			t.Errorf("sync_status param should not be present, got %q", query.Get("sync_status"))
		}
		if query.Has("health_status") {
			t.Errorf("health_status param should not be present, got %q", query.Get("health_status"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": []map[string]interface{}{},
			"total":     0,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resources, total, err := c.ListGitOpsResources(context.Background(), ListGitOpsResourcesParams{
		ClusterID: "staging",
	})
	if err != nil {
		t.Fatalf("ListGitOpsResources() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(resources) != 0 {
		t.Errorf("resource count = %d, want 0", len(resources))
	}
}

func TestGetGitOpsResource_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/operations/gitops/res-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resource": map[string]interface{}{
				"id": "res-1", "cluster_id": "prod-east", "tool": "argocd",
				"namespace": "argocd", "name": "my-app", "kind": "Application",
				"sync_status": "Synced", "health_status": "Healthy",
				"source_url": "https://github.com/org/repo", "revision": "abc123",
				"target_revision": "main", "auto_sync": true, "suspended": false,
				"entity_id": "ent-1", "entity_name": "my-service", "owner_team": "platform",
				"last_synced_at": "2025-06-01T10:00:00Z",
				"created_at":     "2025-01-15T10:00:00Z", "updated_at": "2025-06-01T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resource, err := c.GetGitOpsResource(context.Background(), "res-1")
	if err != nil {
		t.Fatalf("GetGitOpsResource() error = %v", err)
	}
	if resource.ID != "res-1" {
		t.Errorf("ID = %q, want %q", resource.ID, "res-1")
	}
	if resource.ClusterID != "prod-east" {
		t.Errorf("ClusterID = %q, want %q", resource.ClusterID, "prod-east")
	}
	if resource.Tool != "argocd" {
		t.Errorf("Tool = %q, want %q", resource.Tool, "argocd")
	}
	if resource.Namespace != "argocd" {
		t.Errorf("Namespace = %q, want %q", resource.Namespace, "argocd")
	}
	if resource.Name != "my-app" {
		t.Errorf("Name = %q, want %q", resource.Name, "my-app")
	}
	if resource.Kind != "Application" {
		t.Errorf("Kind = %q, want %q", resource.Kind, "Application")
	}
	if resource.SyncStatus != "Synced" {
		t.Errorf("SyncStatus = %q, want %q", resource.SyncStatus, "Synced")
	}
	if resource.HealthStatus != "Healthy" {
		t.Errorf("HealthStatus = %q, want %q", resource.HealthStatus, "Healthy")
	}
	if !resource.AutoSync {
		t.Error("AutoSync = false, want true")
	}
	if resource.Suspended {
		t.Error("Suspended = true, want false")
	}
	if resource.TargetRevision != "main" {
		t.Errorf("TargetRevision = %q, want %q", resource.TargetRevision, "main")
	}
}

func TestGetGitOpsResource_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "NOT_FOUND", "message": "GitOps resource not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetGitOpsResource(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found resource, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound(err) to be true, got false; err = %v", err)
	}
}

func TestGetGitOpsStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/operations/gitops/stats" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total":       42,
			"synced":      35,
			"out_of_sync": 3,
			"failed":      1,
			"suspended":   2,
			"unknown":     1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	stats, err := c.GetGitOpsStats(context.Background())
	if err != nil {
		t.Fatalf("GetGitOpsStats() error = %v", err)
	}
	if stats.Total != 42 {
		t.Errorf("Total = %d, want 42", stats.Total)
	}
	if stats.Synced != 35 {
		t.Errorf("Synced = %d, want 35", stats.Synced)
	}
	if stats.OutOfSync != 3 {
		t.Errorf("OutOfSync = %d, want 3", stats.OutOfSync)
	}
	if stats.Failed != 1 {
		t.Errorf("Failed = %d, want 1", stats.Failed)
	}
	if stats.Suspended != 2 {
		t.Errorf("Suspended = %d, want 2", stats.Suspended)
	}
	if stats.Unknown != 1 {
		t.Errorf("Unknown = %d, want 1", stats.Unknown)
	}
}

func TestListGitOpsResources_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": []map[string]interface{}{},
			"total":     0,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resources, total, err := c.ListGitOpsResources(context.Background(), ListGitOpsResourcesParams{})
	if err != nil {
		t.Fatalf("ListGitOpsResources() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(resources) != 0 {
		t.Errorf("resource count = %d, want 0", len(resources))
	}
}

func TestGetGitOpsResource_FlatResponse(t *testing.T) {
	// Test the fallback parsing path: the API returns the resource at the top level
	// instead of wrapped in {"resource": {...}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/operations/gitops/res-flat" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "res-flat", "cluster_id": "staging", "tool": "fluxcd",
			"namespace": "flux-system", "name": "flat-app", "kind": "Kustomization",
			"sync_status": "Synced", "health_status": "Healthy",
			"auto_sync": true, "suspended": false,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resource, err := c.GetGitOpsResource(context.Background(), "res-flat")
	if err != nil {
		t.Fatalf("GetGitOpsResource() error = %v", err)
	}
	if resource.ID != "res-flat" {
		t.Errorf("ID = %q, want %q", resource.ID, "res-flat")
	}
	if resource.ClusterID != "staging" {
		t.Errorf("ClusterID = %q, want %q", resource.ClusterID, "staging")
	}
	if resource.Tool != "fluxcd" {
		t.Errorf("Tool = %q, want %q", resource.Tool, "fluxcd")
	}
	if resource.Name != "flat-app" {
		t.Errorf("Name = %q, want %q", resource.Name, "flat-app")
	}
	if resource.Kind != "Kustomization" {
		t.Errorf("Kind = %q, want %q", resource.Kind, "Kustomization")
	}
	if resource.SyncStatus != "Synced" {
		t.Errorf("SyncStatus = %q, want %q", resource.SyncStatus, "Synced")
	}
}

func TestGetGitOpsStats_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"code":"INTERNAL","message":"internal server error"}`)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetGitOpsStats(context.Background())
	if err == nil {
		t.Fatal("GetGitOpsStats() expected error for 500 response, got nil")
	}
}
