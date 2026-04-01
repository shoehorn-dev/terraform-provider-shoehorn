package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestForgeMoldResource_Metadata(t *testing.T) {
	r := NewForgeMoldResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_forge_mold" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_forge_mold")
	}
}

func TestForgeMoldResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewForgeMoldResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{
		"id", "slug", "name", "description", "version", "visibility",
		"tags", "icon", "category", "schema_json", "defaults_json",
		"actions", "published", "created_at", "updated_at",
	}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestForgeMoldResource_Schema_SlugIsRequired(t *testing.T) {
	r := NewForgeMoldResource()
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

func TestForgeMoldResource_Schema_NameIsRequired(t *testing.T) {
	r := NewForgeMoldResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["name"]
	if attr == nil {
		t.Fatal("name attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestForgeMoldResource_Schema_VersionIsRequired(t *testing.T) {
	r := NewForgeMoldResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["version"]
	if attr == nil {
		t.Fatal("version attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("version should be required")
	}
}

func TestForgeMoldResource_Schema_VisibilityIsRequired(t *testing.T) {
	r := NewForgeMoldResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["visibility"]
	if attr == nil {
		t.Fatal("visibility attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("visibility should be required")
	}
}

func TestForgeMoldResource_Schema_CategoryIsRequired(t *testing.T) {
	r := NewForgeMoldResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["category"]
	if attr == nil {
		t.Fatal("category attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("category should be required")
	}
}

func TestForgeMoldResource_Schema_IDIsComputed(t *testing.T) {
	r := NewForgeMoldResource()
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

func TestForgeMoldResource_Configure_WithValidClient(t *testing.T) {
	r := &ForgeMoldResource{}
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

func TestForgeMoldResource_Configure_WrongType(t *testing.T) {
	r := &ForgeMoldResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapForgeMoldToState(t *testing.T) {
	mold := &client.ForgeMold{
		ID:          "mold-123",
		Slug:        "my-mold",
		Name:        "My Mold",
		Description: "A test mold",
		Version:     "1.0.0",
		Visibility:  "public",
		Category:    "infrastructure",
		Tags:        []string{"aws", "terraform"},
		Icon:        "cloud",
		Published:   true,
		Schema:      map[string]interface{}{"type": "object"},
		Defaults:    map[string]interface{}{"region": "us-east-1"},
		Actions: []client.ForgeMoldAction{
			{
				Action:      "deploy",
				Label:       "Deploy",
				Description: "Deploy the mold",
				Primary:     true,
			},
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &ForgeMoldResourceModel{}
	mapForgeMoldToState(mold, state)

	if state.ID.ValueString() != "mold-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "mold-123")
	}
	if state.Slug.ValueString() != "my-mold" {
		t.Errorf("Slug = %q, want %q", state.Slug.ValueString(), "my-mold")
	}
	if state.Name.ValueString() != "My Mold" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "My Mold")
	}
	if state.Description.ValueString() != "A test mold" {
		t.Errorf("Description = %q, want %q", state.Description.ValueString(), "A test mold")
	}
	if state.Version.ValueString() != "1.0.0" {
		t.Errorf("Version = %q, want %q", state.Version.ValueString(), "1.0.0")
	}
	if state.Visibility.ValueString() != "public" {
		t.Errorf("Visibility = %q, want %q", state.Visibility.ValueString(), "public")
	}
	if state.Category.ValueString() != "infrastructure" {
		t.Errorf("Category = %q, want %q", state.Category.ValueString(), "infrastructure")
	}
	if state.Published.ValueBool() != true {
		t.Errorf("Published = %v, want %v", state.Published.ValueBool(), true)
	}
	if state.Icon.ValueString() != "cloud" {
		t.Errorf("Icon = %q, want %q", state.Icon.ValueString(), "cloud")
	}
	if state.Tags.IsNull() {
		t.Error("Tags should not be null when mold has tags")
	}
	if state.SchemaJSON.IsNull() {
		t.Error("SchemaJSON should not be null when mold has schema")
	}
	if state.DefaultsJSON.IsNull() {
		t.Error("DefaultsJSON should not be null when mold has defaults")
	}
	if len(state.Actions) != 1 {
		t.Fatalf("Actions length = %d, want 1", len(state.Actions))
	}
	if state.Actions[0].Action.ValueString() != "deploy" {
		t.Errorf("Actions[0].Action = %q, want %q", state.Actions[0].Action.ValueString(), "deploy")
	}
	if state.Actions[0].Label.ValueString() != "Deploy" {
		t.Errorf("Actions[0].Label = %q, want %q", state.Actions[0].Label.ValueString(), "Deploy")
	}
	if state.Actions[0].Primary.ValueBool() != true {
		t.Errorf("Actions[0].Primary = %v, want %v", state.Actions[0].Primary.ValueBool(), true)
	}
}

func TestMapForgeMoldToState_EmptyOptionalFields(t *testing.T) {
	mold := &client.ForgeMold{
		ID:         "mold-456",
		Slug:       "bare-mold",
		Name:       "Bare Mold",
		Version:    "0.1.0",
		Visibility: "private",
		Category:   "app",
		Actions:    []client.ForgeMoldAction{},
	}

	state := &ForgeMoldResourceModel{}
	mapForgeMoldToState(mold, state)

	if state.ID.ValueString() != "mold-456" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "mold-456")
	}
	if !state.Description.IsNull() {
		t.Errorf("Description should be null when empty, got %q", state.Description.ValueString())
	}
	if !state.Icon.IsNull() {
		t.Errorf("Icon should be null when empty, got %q", state.Icon.ValueString())
	}
	if state.Tags.IsNull() != true {
		t.Error("Tags should be null when mold has no tags")
	}
	if !state.SchemaJSON.IsNull() {
		t.Error("SchemaJSON should be null when mold has no schema")
	}
	if !state.DefaultsJSON.IsNull() {
		t.Error("DefaultsJSON should be null when mold has no defaults")
	}
	if len(state.Actions) != 0 {
		t.Errorf("Actions length = %d, want 0", len(state.Actions))
	}
}

func TestMapForgeMoldToState_ClearsOptionalFields(t *testing.T) {
	state := &ForgeMoldResourceModel{
		Description: types.StringValue("old description"),
		Icon:        types.StringValue("old-icon"),
	}

	mold := &client.ForgeMold{
		ID:         "mold-789",
		Slug:       "cleared-mold",
		Name:       "Cleared Mold",
		Version:    "2.0.0",
		Visibility: "tenant",
		Category:   "data",
		Actions:    []client.ForgeMoldAction{},
		// Description and Icon are empty — cleared
	}

	mapForgeMoldToState(mold, state)

	if !state.Description.IsNull() {
		t.Errorf("Description should be null when API returns empty, got %q", state.Description.ValueString())
	}
	if !state.Icon.IsNull() {
		t.Errorf("Icon should be null when API returns empty, got %q", state.Icon.ValueString())
	}
}
