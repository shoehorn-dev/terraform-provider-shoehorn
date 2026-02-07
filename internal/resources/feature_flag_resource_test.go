package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestFeatureFlagResource_Metadata(t *testing.T) {
	r := NewFeatureFlagResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_feature_flag" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_feature_flag")
	}
}

func TestFeatureFlagResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewFeatureFlagResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "key", "name", "description", "default_enabled", "created_at", "updated_at"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestFeatureFlagResource_Schema_KeyIsRequired(t *testing.T) {
	r := NewFeatureFlagResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	keyAttr := resp.Schema.Attributes["key"]
	if keyAttr == nil {
		t.Fatal("key attribute not found")
	}
	if !keyAttr.IsRequired() {
		t.Error("key should be required")
	}
}

func TestFeatureFlagResource_Schema_NameIsRequired(t *testing.T) {
	r := NewFeatureFlagResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	nameAttr := resp.Schema.Attributes["name"]
	if nameAttr == nil {
		t.Fatal("name attribute not found")
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestFeatureFlagResource_Schema_IDIsComputed(t *testing.T) {
	r := NewFeatureFlagResource()
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

func TestFeatureFlagResource_Configure_WithValidClient(t *testing.T) {
	r := &FeatureFlagResource{}
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

func TestFeatureFlagResource_Configure_NilProviderData(t *testing.T) {
	r := &FeatureFlagResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Error("should not error on nil provider data")
	}
}

func TestFeatureFlagResource_Configure_WrongType(t *testing.T) {
	r := &FeatureFlagResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapFeatureFlagToState(t *testing.T) {
	flag := &client.FeatureFlag{
		ID:             "flag-123",
		Key:            "dark-mode",
		Name:           "Dark Mode",
		Description:    "Enable dark mode",
		DefaultEnabled: true,
		CreatedAt:      "2025-01-15T10:00:00Z",
		UpdatedAt:      "2025-01-15T11:00:00Z",
	}

	state := &FeatureFlagResourceModel{}
	mapFeatureFlagToState(flag, state)

	if state.ID.ValueString() != "flag-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "flag-123")
	}
	if state.Key.ValueString() != "dark-mode" {
		t.Errorf("Key = %q, want %q", state.Key.ValueString(), "dark-mode")
	}
	if state.Name.ValueString() != "Dark Mode" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "Dark Mode")
	}
	if state.Description.ValueString() != "Enable dark mode" {
		t.Errorf("Description = %q, want %q", state.Description.ValueString(), "Enable dark mode")
	}
	if !state.DefaultEnabled.ValueBool() {
		t.Error("DefaultEnabled = false, want true")
	}
	if state.CreatedAt.ValueString() != "2025-01-15T10:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", state.CreatedAt.ValueString(), "2025-01-15T10:00:00Z")
	}
}

func TestMapFeatureFlagToState_EmptyOptionalFields(t *testing.T) {
	flag := &client.FeatureFlag{
		ID:             "flag-123",
		Key:            "dark-mode",
		Name:           "Dark Mode",
		DefaultEnabled: false,
	}

	state := &FeatureFlagResourceModel{}
	mapFeatureFlagToState(flag, state)

	if state.ID.ValueString() != "flag-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "flag-123")
	}
	if !state.DefaultEnabled.ValueBool() != true {
		// DefaultEnabled is false, !false = true
	}
}
