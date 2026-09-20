package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestPlatformPolicyResource_Metadata(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_platform_policy" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_platform_policy")
	}
}

func TestPlatformPolicyResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "key", "name", "description", "category", "enabled", "enforcement", "system", "created_at", "updated_at"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestPlatformPolicyResource_Schema_KeyIsRequired(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["key"]
	if attr == nil {
		t.Fatal("key attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("key should be required")
	}
}

func TestPlatformPolicyResource_Schema_EnabledIsRequired(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["enabled"]
	if attr == nil {
		t.Fatal("enabled attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("enabled should be required")
	}
}

// Enforcement changes nothing, so a configuration no longer has to carry it.
func TestPlatformPolicyResource_Schema_EnforcementIsOptionalComputedAndDeprecated(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["enforcement"].(schema.StringAttribute)
	if !ok {
		t.Fatal("enforcement attribute not found")
	}
	if attr.IsRequired() {
		t.Error("enforcement should no longer be required")
	}
	if !attr.IsOptional() {
		t.Error("enforcement should be optional")
	}
	if !attr.IsComputed() {
		t.Error("enforcement should be computed, so omitting it keeps the API's value")
	}
	if attr.DeprecationMessage == "" {
		t.Error("enforcement should be deprecated")
	}
}

// The key is checked during plan, so an unmanageable policy never reaches apply.
func TestPlatformPolicyResource_Schema_KeyRefusesPoliciesTerraformCannotManage(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["key"].(schema.StringAttribute)
	if !ok {
		t.Fatal("key attribute not found")
	}
	if len(attr.Validators) == 0 {
		t.Fatal("key has no validators")
	}

	check := func(key string) validator.StringResponse {
		out := validator.StringResponse{}
		for _, v := range attr.Validators {
			v.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("key"),
				ConfigValue: types.StringValue(key),
			}, &out)
		}
		return out
	}

	if !check("tenant-isolation").Diagnostics.HasError() {
		t.Error("an always-on policy passed validation")
	}
	if !check("stale-entity-cleanup").Diagnostics.HasError() {
		t.Error("a removed policy passed validation")
	}
	if out := check("governance-auto-actions"); out.Diagnostics.HasError() {
		t.Errorf("a setting failed validation: %v", out.Diagnostics)
	}
}

func TestPlatformPolicyResource_Schema_IDIsComputed(t *testing.T) {
	r := NewPlatformPolicyResource()
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

func TestPlatformPolicyResource_Schema_SystemIsComputed(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["system"]
	if attr == nil {
		t.Fatal("system attribute not found")
	}
	if !attr.IsComputed() {
		t.Error("system should be computed")
	}
}

func TestPlatformPolicyResource_Configure_WithValidClient(t *testing.T) {
	r := &PlatformPolicyResource{}
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

func TestPlatformPolicyResource_Configure_WrongType(t *testing.T) {
	r := &PlatformPolicyResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapPolicyToState(t *testing.T) {
	policy := &client.PlatformPolicy{
		ID:          "pol-123",
		Key:         "tenant-isolation",
		Name:        "Tenant Isolation",
		Description: "Enforces tenant isolation",
		Category:    "security",
		Enabled:     true,
		Enforcement: "block",
		System:      true,
		CreatedAt:   "2025-01-15T10:00:00Z",
		UpdatedAt:   "2025-01-15T11:00:00Z",
	}

	state := &PlatformPolicyResourceModel{}
	mapPolicyToState(policy, state)

	if state.ID.ValueString() != "pol-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "pol-123")
	}
	if state.Key.ValueString() != "tenant-isolation" {
		t.Errorf("Key = %q, want %q", state.Key.ValueString(), "tenant-isolation")
	}
	if state.Name.ValueString() != "Tenant Isolation" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "Tenant Isolation")
	}
	if !state.Enabled.ValueBool() {
		t.Error("Enabled = false, want true")
	}
	if state.Enforcement.ValueString() != "block" {
		t.Errorf("Enforcement = %q, want %q", state.Enforcement.ValueString(), "block")
	}
	if !state.System.ValueBool() {
		t.Error("System = false, want true")
	}
	if state.Category.ValueString() != "security" {
		t.Errorf("Category = %q, want %q", state.Category.ValueString(), "security")
	}
}

func TestMapPolicyToState_EmptyOptionalFields(t *testing.T) {
	policy := &client.PlatformPolicy{
		ID:      "pol-123",
		Key:     "test-policy",
		Name:    "Test Policy",
		Enabled: false,
		System:  false,
	}

	state := &PlatformPolicyResourceModel{}
	mapPolicyToState(policy, state)

	if state.ID.ValueString() != "pol-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "pol-123")
	}
	if state.Enabled.ValueBool() {
		t.Error("Enabled = true, want false")
	}
	if state.System.ValueBool() {
		t.Error("System = true, want false")
	}
}

func TestMapPolicyToState_ClearsOptionalFields(t *testing.T) {
	state := &PlatformPolicyResourceModel{
		Description: types.StringValue("Old description"),
		Category:    types.StringValue("security"),
		Enforcement: types.StringValue("block"),
	}

	policy := &client.PlatformPolicy{
		ID:      "pol-123",
		Key:     "test-policy",
		Name:    "Test Policy",
		Enabled: true,
		// Description, Category, Enforcement are empty — cleared
	}

	mapPolicyToState(policy, state)

	if !state.Description.IsNull() {
		t.Errorf("Description should be null when API returns empty, got %q", state.Description.ValueString())
	}
	if !state.Category.IsNull() {
		t.Errorf("Category should be null when API returns empty, got %q", state.Category.ValueString())
	}
	// Enforcement is Required, so it should always be set (even if empty string)
	if state.Enforcement.IsNull() {
		t.Error("Enforcement should not be null (Required field)")
	}
}
