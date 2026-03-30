package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestMarketplaceInstallationResource_Metadata(t *testing.T) {
	r := NewMarketplaceInstallationResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_marketplace_installation" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_marketplace_installation")
	}
}

func TestMarketplaceInstallationResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewMarketplaceInstallationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{
		"id", "slug", "enabled", "config_json", "kind",
		"version", "sync_status", "installed_by", "created_at", "updated_at",
	}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestMarketplaceInstallationResource_Schema_SlugIsRequired(t *testing.T) {
	r := NewMarketplaceInstallationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["slug"]
	if attr == nil {
		t.Fatal("slug attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("slug should be required")
	}
}

func TestMarketplaceInstallationResource_Schema_ConfigJSONIsSensitive(t *testing.T) {
	r := NewMarketplaceInstallationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["config_json"]
	if attr == nil {
		t.Fatal("config_json attribute not found")
	}
	if !attr.IsSensitive() {
		t.Error("config_json should be sensitive")
	}
}

func TestMarketplaceInstallationResource_Schema_IDIsComputed(t *testing.T) {
	r := NewMarketplaceInstallationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["id"]
	if attr == nil {
		t.Fatal("id attribute not found")
	}
	if !attr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestMarketplaceInstallationResource_Configure_WithValidClient(t *testing.T) {
	r := &MarketplaceInstallationResource{}
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

func TestMarketplaceInstallationResource_Configure_WrongType(t *testing.T) {
	r := &MarketplaceInstallationResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapMarketplaceInstallationToState(t *testing.T) {
	installation := &client.MarketplaceInstallation{
		ID:          "inst-123",
		Slug:        "my-addon",
		Enabled:     true,
		Kind:        "integration",
		Version:     "2.1.0",
		SyncStatus:  "synced",
		InstalledBy: "admin@example.com",
		CreatedAt:   "2025-01-15T10:00:00Z",
		UpdatedAt:   "2025-01-15T11:00:00Z",
	}

	state := &MarketplaceInstallationResourceModel{}
	mapMarketplaceInstallationToState(installation, state)

	if state.ID.ValueString() != "inst-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "inst-123")
	}
	if state.Slug.ValueString() != "my-addon" {
		t.Errorf("Slug = %q, want %q", state.Slug.ValueString(), "my-addon")
	}
	if state.Enabled.ValueBool() != true {
		t.Errorf("Enabled = %v, want %v", state.Enabled.ValueBool(), true)
	}
	if state.Kind.ValueString() != "integration" {
		t.Errorf("Kind = %q, want %q", state.Kind.ValueString(), "integration")
	}
	if state.Version.ValueString() != "2.1.0" {
		t.Errorf("Version = %q, want %q", state.Version.ValueString(), "2.1.0")
	}
	if state.SyncStatus.ValueString() != "synced" {
		t.Errorf("SyncStatus = %q, want %q", state.SyncStatus.ValueString(), "synced")
	}
	if state.InstalledBy.ValueString() != "admin@example.com" {
		t.Errorf("InstalledBy = %q, want %q", state.InstalledBy.ValueString(), "admin@example.com")
	}
}

func TestMapMarketplaceInstallationToState_EmptyOptionalFields(t *testing.T) {
	installation := &client.MarketplaceInstallation{
		ID:      "inst-456",
		Slug:    "bare-addon",
		Enabled: false,
	}

	state := &MarketplaceInstallationResourceModel{}
	mapMarketplaceInstallationToState(installation, state)

	if state.ID.ValueString() != "inst-456" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "inst-456")
	}
	if !state.Kind.IsNull() {
		t.Errorf("Kind should be null when empty, got %q", state.Kind.ValueString())
	}
	if !state.Version.IsNull() {
		t.Errorf("Version should be null when empty, got %q", state.Version.ValueString())
	}
	if !state.SyncStatus.IsNull() {
		t.Errorf("SyncStatus should be null when empty, got %q", state.SyncStatus.ValueString())
	}
	if !state.InstalledBy.IsNull() {
		t.Errorf("InstalledBy should be null when empty, got %q", state.InstalledBy.ValueString())
	}
}

func TestMapMarketplaceInstallationToState_ClearsOptionalFields(t *testing.T) {
	state := &MarketplaceInstallationResourceModel{
		Kind:        types.StringValue("old-kind"),
		InstalledBy: types.StringValue("old-user"),
	}

	installation := &client.MarketplaceInstallation{
		ID:      "inst-789",
		Slug:    "cleared-addon",
		Enabled: true,
		// Kind and InstalledBy are empty — cleared
	}

	mapMarketplaceInstallationToState(installation, state)

	if !state.Kind.IsNull() {
		t.Errorf("Kind should be null when API returns empty, got %q", state.Kind.ValueString())
	}
	if !state.InstalledBy.IsNull() {
		t.Errorf("InstalledBy should be null when API returns empty, got %q", state.InstalledBy.ValueString())
	}
}
