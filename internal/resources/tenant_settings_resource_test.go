package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	var diags diag.Diagnostics
	mapSettingsToState(context.Background(), settings, state, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

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

func TestTenantSettingsResource_Schema_HasAnnouncementAttribute(t *testing.T) {
	r := NewTenantSettingsResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	if _, ok := attrs["announcement"]; !ok {
		t.Error("schema missing announcement attribute")
	}
}

func TestMapSettingsToState_WithAnnouncement(t *testing.T) {
	settings := &client.TenantSettings{
		ID:       "settings-1",
		TenantID: "tenant-123",
		Appearance: client.AppearanceSettings{
			PrimaryColor: "#3b82f6",
		},
		Announcement: client.AnnouncementSettings{
			Enabled:   true,
			Message:   "Maintenance scheduled",
			Type:      "warning",
			Pinned:    false,
			LinkURL:   "https://status.example.com",
			LinkText:  "View Status",
			UpdatedAt: "2025-01-15T12:00:00Z",
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &TenantSettingsResourceModel{}
	var diags diag.Diagnostics
	mapSettingsToState(context.Background(), settings, state, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.Announcement.IsNull() {
		t.Fatal("announcement should not be null")
	}

	var announcement AnnouncementSettingsModel
	d := state.Announcement.As(context.Background(), &announcement, basetypes.ObjectAsOptions{})
	if d.HasError() {
		t.Fatalf("failed to parse announcement: %v", d)
	}

	if !announcement.Enabled.ValueBool() {
		t.Error("announcement.Enabled should be true")
	}
	if announcement.Message.ValueString() != "Maintenance scheduled" {
		t.Errorf("announcement.Message = %q, want %q", announcement.Message.ValueString(), "Maintenance scheduled")
	}
	if announcement.Type.ValueString() != "warning" {
		t.Errorf("announcement.Type = %q, want %q", announcement.Type.ValueString(), "warning")
	}
	if announcement.LinkURL.ValueString() != "https://status.example.com" {
		t.Errorf("announcement.LinkURL = %q, want %q", announcement.LinkURL.ValueString(), "https://status.example.com")
	}
}

func TestMapSettingsToState_WithoutAnnouncement(t *testing.T) {
	settings := &client.TenantSettings{
		ID:       "settings-1",
		TenantID: "tenant-123",
		Appearance: client.AppearanceSettings{
			PrimaryColor: "#3b82f6",
		},
		Announcement: client.AnnouncementSettings{
			Enabled: false,
			Message: "",
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &TenantSettingsResourceModel{}
	var diags diag.Diagnostics
	mapSettingsToState(context.Background(), settings, state, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Announcement should be null when disabled and no message
	if !state.Announcement.IsNull() {
		t.Error("announcement should be null when disabled with no message")
	}
}

func TestBuildAnnouncementFromModel_WithAnnouncement(t *testing.T) {
	announcementAttrs := map[string]attr.Value{
		"enabled":    types.BoolValue(true),
		"message":    types.StringValue("Maintenance window"),
		"type":       types.StringValue("warning"),
		"pinned":     types.BoolValue(false),
		"link_url":   types.StringValue("https://status.example.com"),
		"link_text":  types.StringValue("View Status"),
		"updated_at": types.StringValue("2025-01-15T12:00:00Z"),
	}

	announcementTypes := map[string]attr.Type{
		"enabled":    types.BoolType,
		"message":    types.StringType,
		"type":       types.StringType,
		"pinned":     types.BoolType,
		"link_url":   types.StringType,
		"link_text":  types.StringType,
		"updated_at": types.StringType,
	}

	announcementObj, diags := types.ObjectValue(announcementTypes, announcementAttrs)
	if diags.HasError() {
		t.Fatalf("failed to create announcement object: %v", diags)
	}

	model := &TenantSettingsResourceModel{
		Announcement: announcementObj,
	}

	var d diag.Diagnostics
	announcement := buildAnnouncementFromModel(context.Background(), model, &d)
	if d.HasError() {
		t.Fatalf("buildAnnouncementFromModel failed: %v", d)
	}

	if announcement == nil {
		t.Fatal("announcement should not be nil")
	}

	if !announcement.Enabled {
		t.Error("announcement.Enabled should be true")
	}
	if announcement.Message != "Maintenance window" {
		t.Errorf("announcement.Message = %q, want %q", announcement.Message, "Maintenance window")
	}
	if announcement.Type != "warning" {
		t.Errorf("announcement.Type = %q, want %q", announcement.Type, "warning")
	}
	if announcement.Pinned {
		t.Error("announcement.Pinned should be false")
	}
	if announcement.LinkURL != "https://status.example.com" {
		t.Errorf("announcement.LinkURL = %q, want %q", announcement.LinkURL, "https://status.example.com")
	}
	if announcement.LinkText != "View Status" {
		t.Errorf("announcement.LinkText = %q, want %q", announcement.LinkText, "View Status")
	}
}

func TestBuildAnnouncementFromModel_NullAnnouncement(t *testing.T) {
	model := &TenantSettingsResourceModel{
		Announcement: types.ObjectNull(map[string]attr.Type{
			"enabled":    types.BoolType,
			"message":    types.StringType,
			"type":       types.StringType,
			"pinned":     types.BoolType,
			"link_url":   types.StringType,
			"link_text":  types.StringType,
			"updated_at": types.StringType,
		}),
	}

	var diags diag.Diagnostics
	announcement := buildAnnouncementFromModel(context.Background(), model, &diags)

	if announcement != nil {
		t.Error("announcement should be nil for null model")
	}
	if diags.HasError() {
		t.Errorf("unexpected diagnostics: %v", diags)
	}
}
