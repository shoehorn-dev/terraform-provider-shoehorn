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

func TestGetSettings_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/settings" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "settings-1",
			"tenant_id": "tenant-123",
			"appearance": map[string]interface{}{
				"primary_color":        "#3b82f6",
				"secondary_color":      "#64748b",
				"accent_color":         "#8b5cf6",
				"logo_url":             "https://example.com/logo.png",
				"default_theme":        "dark",
				"platform_name":        "Acme Portal",
				"platform_description": "Internal Developer Platform",
				"company_name":         "Acme Corp",
			},
			"created_at": "2025-01-15T10:00:00Z",
			"updated_at": "2025-01-15T11:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	settings, err := c.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}

	if settings.ID != "settings-1" {
		t.Errorf("ID = %q, want %q", settings.ID, "settings-1")
	}
	if settings.Appearance.PrimaryColor != "#3b82f6" {
		t.Errorf("PrimaryColor = %q, want %q", settings.Appearance.PrimaryColor, "#3b82f6")
	}
	if settings.Appearance.PlatformName != "Acme Portal" {
		t.Errorf("PlatformName = %q, want %q", settings.Appearance.PlatformName, "Acme Portal")
	}
	if settings.Appearance.DefaultTheme != "dark" {
		t.Errorf("DefaultTheme = %q, want %q", settings.Appearance.DefaultTheme, "dark")
	}
	if settings.Appearance.CompanyName != "Acme Corp" {
		t.Errorf("CompanyName = %q, want %q", settings.Appearance.CompanyName, "Acme Corp")
	}
}

func TestGetSettings_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "settings-1",
			"appearance": map[string]interface{}{},
			"created_at": "2025-01-15T10:00:00Z",
			"updated_at": "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	settings, err := c.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}

	if settings.Appearance.PrimaryColor != "" {
		t.Errorf("PrimaryColor = %q, want empty", settings.Appearance.PrimaryColor)
	}
}

func TestUpdateSettings_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/admin/settings" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req UpdateSettingsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}

		if req.Appearance.PlatformName != "Updated Portal" {
			t.Errorf("PlatformName = %q, want %q", req.Appearance.PlatformName, "Updated Portal")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "settings-1",
			"tenant_id": "tenant-123",
			"appearance": map[string]interface{}{
				"primary_color": "#3b82f6",
				"platform_name": "Updated Portal",
				"company_name":  "Acme Corp",
			},
			"created_at": "2025-01-15T10:00:00Z",
			"updated_at": "2025-01-15T12:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	settings, err := c.UpdateSettings(context.Background(), UpdateSettingsRequest{
		Appearance: AppearanceSettings{
			PrimaryColor: "#3b82f6",
			PlatformName: "Updated Portal",
			CompanyName:  "Acme Corp",
		},
	})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	if settings.Appearance.PlatformName != "Updated Portal" {
		t.Errorf("PlatformName = %q, want %q", settings.Appearance.PlatformName, "Updated Portal")
	}
	if settings.UpdatedAt != "2025-01-15T12:00:00Z" {
		t.Errorf("UpdatedAt = %q, want %q", settings.UpdatedAt, "2025-01-15T12:00:00Z")
	}
}
