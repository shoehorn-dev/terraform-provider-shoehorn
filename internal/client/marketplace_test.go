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

func TestListMarketplaceItems_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/marketplace" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"slug": "slack-notifier", "kind": "addon", "name": "Slack Notifier", "version": "1.2.0", "tier": "free", "verified": true},
				{"slug": "dashboard-widget", "kind": "widget", "name": "Dashboard Widget", "version": "2.0.1", "tier": "pro"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	items, err := c.ListMarketplaceItems(context.Background(), "", "")
	if err != nil {
		t.Fatalf("ListMarketplaceItems() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("item count = %d, want 2", len(items))
	}
	if items[0].Slug != "slack-notifier" {
		t.Errorf("Slug = %q, want %q", items[0].Slug, "slack-notifier")
	}
	if items[0].Kind != "addon" {
		t.Errorf("Kind = %q, want %q", items[0].Kind, "addon")
	}
	if items[0].Name != "Slack Notifier" {
		t.Errorf("Name = %q, want %q", items[0].Name, "Slack Notifier")
	}
	if items[0].Version != "1.2.0" {
		t.Errorf("Version = %q, want %q", items[0].Version, "1.2.0")
	}
	if items[0].Tier != "free" {
		t.Errorf("Tier = %q, want %q", items[0].Tier, "free")
	}
	if !items[0].Verified {
		t.Error("Verified = false, want true")
	}
}

func TestListMarketplaceItems_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/marketplace" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("kind") != "addon" {
			t.Errorf("kind param = %q, want %q", r.URL.Query().Get("kind"), "addon")
		}
		if r.URL.Query().Get("category") != "notifications" {
			t.Errorf("category param = %q, want %q", r.URL.Query().Get("category"), "notifications")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"slug": "slack-notifier", "kind": "addon", "name": "Slack Notifier", "version": "1.2.0", "category": "notifications"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	items, err := c.ListMarketplaceItems(context.Background(), "addon", "notifications")
	if err != nil {
		t.Fatalf("ListMarketplaceItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("item count = %d, want 1", len(items))
	}
	if items[0].Category != "notifications" {
		t.Errorf("Category = %q, want %q", items[0].Category, "notifications")
	}
}

func TestListMarketplaceInstallations_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/marketplace/installed" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"installations": []map[string]interface{}{
				{"id": "inst-001", "item_slug": "slack-notifier", "item_kind": "addon", "item_version": "1.2.0", "enabled": true},
				{"id": "inst-002", "item_slug": "dashboard-widget", "item_kind": "widget", "item_version": "2.0.1", "enabled": false},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	installations, err := c.ListMarketplaceInstallations(context.Background())
	if err != nil {
		t.Fatalf("ListMarketplaceInstallations() error = %v", err)
	}
	if len(installations) != 2 {
		t.Fatalf("installation count = %d, want 2", len(installations))
	}
	if installations[0].ID != "inst-001" {
		t.Errorf("ID = %q, want %q", installations[0].ID, "inst-001")
	}
	if installations[0].Slug != "slack-notifier" {
		t.Errorf("Slug = %q, want %q", installations[0].Slug, "slack-notifier")
	}
	if !installations[0].Enabled {
		t.Error("Enabled = false, want true")
	}
	if installations[1].Enabled {
		t.Error("installations[1].Enabled = true, want false")
	}
}

func TestGetMarketplaceInstallation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/marketplace/installed/slack-notifier" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"installation": map[string]interface{}{
				"id": "inst-001", "item_slug": "slack-notifier", "item_kind": "addon",
				"item_version": "1.2.0", "enabled": true,
				"config":       map[string]interface{}{"webhook_url": "https://hooks.slack.com/xxx"},
				"sync_status":  "synced",
				"installed_by": "user@example.com",
				"created_at":   "2025-06-01T10:00:00Z",
				"updated_at":   "2025-06-01T12:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	inst, err := c.GetMarketplaceInstallation(context.Background(), "slack-notifier")
	if err != nil {
		t.Fatalf("GetMarketplaceInstallation() error = %v", err)
	}
	if inst.ID != "inst-001" {
		t.Errorf("ID = %q, want %q", inst.ID, "inst-001")
	}
	if inst.Slug != "slack-notifier" {
		t.Errorf("Slug = %q, want %q", inst.Slug, "slack-notifier")
	}
	if inst.Kind != "addon" {
		t.Errorf("Kind = %q, want %q", inst.Kind, "addon")
	}
	if inst.Version != "1.2.0" {
		t.Errorf("Version = %q, want %q", inst.Version, "1.2.0")
	}
	if !inst.Enabled {
		t.Error("Enabled = false, want true")
	}
	if inst.Config["webhook_url"] != "https://hooks.slack.com/xxx" {
		t.Errorf("Config[webhook_url] = %v, want %q", inst.Config["webhook_url"], "https://hooks.slack.com/xxx")
	}
	if inst.SyncStatus != "synced" {
		t.Errorf("SyncStatus = %q, want %q", inst.SyncStatus, "synced")
	}
	if inst.InstalledBy != "user@example.com" {
		t.Errorf("InstalledBy = %q, want %q", inst.InstalledBy, "user@example.com")
	}
}

func TestInstallMarketplaceItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/marketplace/install" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		if req["slug"] != "slack-notifier" {
			t.Errorf("slug = %q, want %q", req["slug"], "slack-notifier")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"installation": map[string]interface{}{
				"id": "inst-001", "item_slug": "slack-notifier", "item_kind": "addon",
				"item_version": "1.2.0", "enabled": true,
				"created_at": "2025-06-01T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	inst, err := c.InstallMarketplaceItem(context.Background(), "slack-notifier")
	if err != nil {
		t.Fatalf("InstallMarketplaceItem() error = %v", err)
	}
	if inst.ID != "inst-001" {
		t.Errorf("ID = %q, want %q", inst.ID, "inst-001")
	}
	if inst.Slug != "slack-notifier" {
		t.Errorf("Slug = %q, want %q", inst.Slug, "slack-notifier")
	}
	if !inst.Enabled {
		t.Error("Enabled = false, want true")
	}
}

func TestUninstallMarketplaceItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/marketplace/slack-notifier/uninstall" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.UninstallMarketplaceItem(context.Background(), "slack-notifier")
	if err != nil {
		t.Fatalf("UninstallMarketplaceItem() error = %v", err)
	}
}

func TestEnableMarketplaceItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/marketplace/slack-notifier/enable" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.EnableMarketplaceItem(context.Background(), "slack-notifier")
	if err != nil {
		t.Fatalf("EnableMarketplaceItem() error = %v", err)
	}
}

func TestDisableMarketplaceItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/marketplace/slack-notifier/disable" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DisableMarketplaceItem(context.Background(), "slack-notifier")
	if err != nil {
		t.Fatalf("DisableMarketplaceItem() error = %v", err)
	}
}

func TestUpdateMarketplaceItemConfig_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/marketplace/slack-notifier/config" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		configRaw, ok := req["config"]
		if !ok {
			t.Error("request body missing 'config' key")
		}
		config, ok := configRaw.(map[string]interface{})
		if !ok {
			t.Error("config is not a map")
		}
		if config["webhook_url"] != "https://hooks.slack.com/new" {
			t.Errorf("config[webhook_url] = %v, want %q", config["webhook_url"], "https://hooks.slack.com/new")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"installation": map[string]interface{}{
				"id": "inst-001", "item_slug": "slack-notifier", "item_kind": "addon",
				"item_version": "1.2.0", "enabled": true,
				"config":     map[string]interface{}{"webhook_url": "https://hooks.slack.com/new"},
				"updated_at": "2025-06-01T14:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	inst, err := c.UpdateMarketplaceItemConfig(context.Background(), "slack-notifier", map[string]interface{}{
		"webhook_url": "https://hooks.slack.com/new",
	})
	if err != nil {
		t.Fatalf("UpdateMarketplaceItemConfig() error = %v", err)
	}
	if inst.ID != "inst-001" {
		t.Errorf("ID = %q, want %q", inst.ID, "inst-001")
	}
	if inst.Config["webhook_url"] != "https://hooks.slack.com/new" {
		t.Errorf("Config[webhook_url] = %v, want %q", inst.Config["webhook_url"], "https://hooks.slack.com/new")
	}
}

func TestListMarketplaceItems_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	items, err := c.ListMarketplaceItems(context.Background(), "", "")
	if err != nil {
		t.Fatalf("ListMarketplaceItems() error = %v", err)
	}
	if len(items) != 0 {
		t.Errorf("item count = %d, want 0", len(items))
	}
}

func TestGetMarketplaceInstallation_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "NOT_FOUND", "message": "marketplace installation not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetMarketplaceInstallation(context.Background(), "nonexistent-addon")
	if err == nil {
		t.Fatal("GetMarketplaceInstallation() expected error for 404, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestInstallMarketplaceItem_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "ALREADY_EXISTS", "message": "marketplace item already installed",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.InstallMarketplaceItem(context.Background(), "slack-notifier")
	if err == nil {
		t.Fatal("InstallMarketplaceItem() expected error for 409, got nil")
	}
	if !IsAlreadyExists(err) {
		t.Errorf("expected already-exists error, got: %v", err)
	}
}

func TestMarketplace_Lifecycle(t *testing.T) {
	installations := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Install
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/marketplace/install":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			slug := req["slug"].(string)
			inst := map[string]interface{}{
				"id": "inst-001", "item_slug": slug, "item_kind": "addon",
				"item_version": "1.0.0", "enabled": true,
				"config":     map[string]interface{}{},
				"created_at": "2025-06-01T10:00:00Z",
			}
			installations[slug] = inst

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"installation": inst})

		// Get installed item
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/marketplace/installed/test-addon":
			if inst, ok := installations["test-addon"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"installation": inst})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "not found"})
			}

		// Enable
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/marketplace/test-addon/enable":
			if inst, ok := installations["test-addon"]; ok {
				inst["enabled"] = true
				installations["test-addon"] = inst
			}
			w.WriteHeader(http.StatusOK)

		// Disable
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/marketplace/test-addon/disable":
			if inst, ok := installations["test-addon"]; ok {
				inst["enabled"] = false
				installations["test-addon"] = inst
			}
			w.WriteHeader(http.StatusOK)

		// Update config
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/marketplace/test-addon/config":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			if inst, ok := installations["test-addon"]; ok {
				inst["config"] = req["config"]
				inst["updated_at"] = "2025-06-01T12:00:00Z"
				installations["test-addon"] = inst
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"installation": inst})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

		// Uninstall
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/marketplace/test-addon/uninstall":
			delete(installations, "test-addon")
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"not found"}`)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// INSTALL
	inst, err := c.InstallMarketplaceItem(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("INSTALL failed: %v", err)
	}
	if inst.ID != "inst-001" {
		t.Errorf("INSTALL: ID = %q, want %q", inst.ID, "inst-001")
	}
	if inst.Slug != "test-addon" {
		t.Errorf("INSTALL: Slug = %q, want %q", inst.Slug, "test-addon")
	}

	// GET
	got, err := c.GetMarketplaceInstallation(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if got.Slug != "test-addon" {
		t.Errorf("GET: Slug = %q, want %q", got.Slug, "test-addon")
	}
	if !got.Enabled {
		t.Error("GET: Enabled = false, want true")
	}

	// DISABLE
	err = c.DisableMarketplaceItem(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("DISABLE failed: %v", err)
	}

	// Verify disabled
	got, err = c.GetMarketplaceInstallation(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("GET after DISABLE failed: %v", err)
	}
	if got.Enabled {
		t.Error("GET after DISABLE: Enabled = true, want false")
	}

	// ENABLE
	err = c.EnableMarketplaceItem(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("ENABLE failed: %v", err)
	}

	// Verify enabled
	got, err = c.GetMarketplaceInstallation(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("GET after ENABLE failed: %v", err)
	}
	if !got.Enabled {
		t.Error("GET after ENABLE: Enabled = false, want true")
	}

	// UPDATE CONFIG
	updated, err := c.UpdateMarketplaceItemConfig(context.Background(), "test-addon", map[string]interface{}{
		"webhook_url": "https://hooks.slack.com/test",
		"channel":     "#alerts",
	})
	if err != nil {
		t.Fatalf("UPDATE CONFIG failed: %v", err)
	}
	if updated.Config["webhook_url"] != "https://hooks.slack.com/test" {
		t.Errorf("UPDATE CONFIG: Config[webhook_url] = %v, want %q", updated.Config["webhook_url"], "https://hooks.slack.com/test")
	}
	if updated.Config["channel"] != "#alerts" {
		t.Errorf("UPDATE CONFIG: Config[channel] = %v, want %q", updated.Config["channel"], "#alerts")
	}

	// UNINSTALL
	err = c.UninstallMarketplaceItem(context.Background(), "test-addon")
	if err != nil {
		t.Fatalf("UNINSTALL failed: %v", err)
	}

	// Verify uninstalled (should get 404)
	_, err = c.GetMarketplaceInstallation(context.Background(), "test-addon")
	if err == nil {
		t.Fatal("GET after UNINSTALL: expected error, got nil")
	}
}
