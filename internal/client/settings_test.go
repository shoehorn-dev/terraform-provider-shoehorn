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
				"primary_color":             "#3b82f6",
				"secondary_color":           "#64748b",
				"accent_color":              "#8b5cf6",
				"logo_url":                  "https://example.com/logo.png",
				"logo_url_dark":             "https://example.com/logo-dark.png",
				"default_theme":             "dark",
				"platform_name":             "Acme Portal",
				"platform_name_color_light": "#2870a4",
				"platform_name_color_dark":  "#8ec9f0",
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
	if settings.Appearance.LogoURLDark != "https://example.com/logo-dark.png" {
		t.Errorf("LogoURLDark = %q, want %q", settings.Appearance.LogoURLDark, "https://example.com/logo-dark.png")
	}
	if settings.Appearance.PlatformNameColorLight != "#2870a4" {
		t.Errorf("PlatformNameColorLight = %q, want %q", settings.Appearance.PlatformNameColorLight, "#2870a4")
	}
	if settings.Appearance.PlatformNameColorDark != "#8ec9f0" {
		t.Errorf("PlatformNameColorDark = %q, want %q", settings.Appearance.PlatformNameColorDark, "#8ec9f0")
	}
}

func TestUpdateSettings_SendsDarkLogoAndColorsButNotRemovedFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		appearance := raw["appearance"]
		want := map[string]string{
			"logo_url_dark":             "https://example.com/logo-dark.png",
			"platform_name_color_light": "#2870a4",
			"platform_name_color_dark":  "#8ec9f0",
		}
		for key, value := range want {
			if appearance[key] != value {
				t.Errorf("appearance[%q] = %v, want %q", key, appearance[key], value)
			}
		}
		for _, removed := range []string{"platform_description", "company_name"} {
			if _, ok := appearance[removed]; ok {
				t.Errorf("appearance must not send %q, the platform removed it", removed)
			}
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "settings-1", "appearance": appearance})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.UpdateSettings(context.Background(), UpdateSettingsRequest{
		Appearance: AppearanceSettings{
			LogoURLDark:            "https://example.com/logo-dark.png",
			PlatformNameColorLight: "#2870a4",
			PlatformNameColorDark:  "#8ec9f0",
		},
	})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
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

func TestGetSettings_Security(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "settings-1",
			"appearance": map[string]interface{}{},
			"security": map[string]interface{}{
				"infrastructure_owner_team": "platform",
			},
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

	if settings.Security.InfrastructureOwnerTeam != "platform" {
		t.Errorf("InfrastructureOwnerTeam = %q, want %q", settings.Security.InfrastructureOwnerTeam, "platform")
	}
}

func TestUpdateSettings_Security(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req UpdateSettingsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req.Security == nil || req.Security.InfrastructureOwnerTeam != "checkout" {
			t.Errorf("request security = %+v, want InfrastructureOwnerTeam=checkout", req.Security)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         "settings-1",
			"appearance": map[string]interface{}{},
			"security": map[string]interface{}{
				"infrastructure_owner_team": "checkout",
			},
			"created_at": "2025-01-15T10:00:00Z",
			"updated_at": "2025-01-15T12:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	settings, err := c.UpdateSettings(context.Background(), UpdateSettingsRequest{
		Security: &SecuritySettings{InfrastructureOwnerTeam: "checkout"},
	})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if settings.Security.InfrastructureOwnerTeam != "checkout" {
		t.Errorf("InfrastructureOwnerTeam = %q, want %q", settings.Security.InfrastructureOwnerTeam, "checkout")
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
