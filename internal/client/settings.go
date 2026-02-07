package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// TenantSettings represents Shoehorn tenant settings.
type TenantSettings struct {
	ID         string             `json:"id,omitempty"`
	TenantID   string             `json:"tenant_id,omitempty"`
	Appearance AppearanceSettings `json:"appearance"`
	CreatedAt  string             `json:"created_at,omitempty"`
	UpdatedAt  string             `json:"updated_at,omitempty"`
}

// AppearanceSettings represents the appearance configuration.
type AppearanceSettings struct {
	PrimaryColor        string `json:"primary_color,omitempty"`
	SecondaryColor      string `json:"secondary_color,omitempty"`
	AccentColor         string `json:"accent_color,omitempty"`
	LogoURL             string `json:"logo_url,omitempty"`
	FaviconURL          string `json:"favicon_url,omitempty"`
	DefaultTheme        string `json:"default_theme,omitempty"`
	PlatformName        string `json:"platform_name,omitempty"`
	PlatformDescription string `json:"platform_description,omitempty"`
	CompanyName         string `json:"company_name,omitempty"`
}

// UpdateSettingsRequest is the request body for updating tenant settings.
type UpdateSettingsRequest struct {
	Appearance AppearanceSettings `json:"appearance"`
}

// GetSettings retrieves the tenant settings.
func (c *Client) GetSettings(ctx context.Context) (*TenantSettings, error) {
	body, err := c.Get(ctx, "/api/v1/admin/settings")
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	var settings TenantSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal settings response: %w", err)
	}

	return &settings, nil
}

// UpdateSettings updates the tenant settings (upsert).
func (c *Client) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) (*TenantSettings, error) {
	body, err := c.Put(ctx, "/api/v1/admin/settings", req)
	if err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}

	var settings TenantSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal update settings response: %w", err)
	}

	return &settings, nil
}
