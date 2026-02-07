package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestTenantSettingsResource_Metadata(t *testing.T) {
	r := NewTenantSettingsResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_tenant_settings" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_tenant_settings")
	}
}

func TestTenantSettingsResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewTenantSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{
		"id", "primary_color", "secondary_color", "accent_color",
		"logo_url", "favicon_url", "default_theme",
		"platform_name", "platform_description", "company_name",
		"created_at", "updated_at",
	}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestTenantSettingsResource_Schema_IDIsComputed(t *testing.T) {
	r := NewTenantSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	idAttr := resp.Schema.Attributes["id"]
	if idAttr == nil {
		t.Fatal("id attribute not found")
	}
	if !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestTenantSettingsResource_Schema_AllAppearanceFieldsOptional(t *testing.T) {
	r := NewTenantSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	optionalFields := []string{
		"primary_color", "secondary_color", "accent_color",
		"logo_url", "favicon_url", "default_theme",
		"platform_name", "platform_description", "company_name",
	}
	for _, name := range optionalFields {
		attr := resp.Schema.Attributes[name]
		if attr == nil {
			t.Errorf("%s attribute not found", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("%s should be optional", name)
		}
	}
}

func TestTenantSettingsResource_Configure_WithValidClient(t *testing.T) {
	r := &TenantSettingsResource{}
	c := client.NewClient("https://test.example.com", "key", 30*time.Second)

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: c,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}
	if r.client != c {
		t.Error("client not set correctly")
	}
}

func TestTenantSettingsResource_Configure_WrongType(t *testing.T) {
	r := &TenantSettingsResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapSettingsToState(t *testing.T) {
	settings := &client.TenantSettings{
		ID:       "settings-1",
		TenantID: "tenant-123",
		Appearance: client.AppearanceSettings{
			PrimaryColor:        "#3b82f6",
			SecondaryColor:      "#64748b",
			AccentColor:         "#8b5cf6",
			LogoURL:             "https://example.com/logo.png",
			DefaultTheme:        "dark",
			PlatformName:        "Acme Portal",
			PlatformDescription: "Internal Developer Platform",
			CompanyName:         "Acme Corp",
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &TenantSettingsResourceModel{}
	mapSettingsToState(settings, state)

	if state.ID.ValueString() != "settings-1" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "settings-1")
	}
	if state.PrimaryColor.ValueString() != "#3b82f6" {
		t.Errorf("PrimaryColor = %q, want %q", state.PrimaryColor.ValueString(), "#3b82f6")
	}
	if state.PlatformName.ValueString() != "Acme Portal" {
		t.Errorf("PlatformName = %q, want %q", state.PlatformName.ValueString(), "Acme Portal")
	}
	if state.DefaultTheme.ValueString() != "dark" {
		t.Errorf("DefaultTheme = %q, want %q", state.DefaultTheme.ValueString(), "dark")
	}
	if state.CompanyName.ValueString() != "Acme Corp" {
		t.Errorf("CompanyName = %q, want %q", state.CompanyName.ValueString(), "Acme Corp")
	}
	if state.CreatedAt.ValueString() != "2025-01-15T10:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", state.CreatedAt.ValueString(), "2025-01-15T10:00:00Z")
	}
}

func TestBuildAppearanceFromModel(t *testing.T) {
	model := &TenantSettingsResourceModel{
		PrimaryColor: types.StringValue("#3b82f6"),
		PlatformName: types.StringValue("Acme Portal"),
		CompanyName:  types.StringValue("Acme Corp"),
		DefaultTheme: types.StringValue("dark"),
	}

	appearance := buildAppearanceFromModel(model)

	if appearance.PrimaryColor != "#3b82f6" {
		t.Errorf("PrimaryColor = %q, want %q", appearance.PrimaryColor, "#3b82f6")
	}
	if appearance.PlatformName != "Acme Portal" {
		t.Errorf("PlatformName = %q, want %q", appearance.PlatformName, "Acme Portal")
	}
	if appearance.CompanyName != "Acme Corp" {
		t.Errorf("CompanyName = %q, want %q", appearance.CompanyName, "Acme Corp")
	}
	if appearance.DefaultTheme != "dark" {
		t.Errorf("DefaultTheme = %q, want %q", appearance.DefaultTheme, "dark")
	}
}
