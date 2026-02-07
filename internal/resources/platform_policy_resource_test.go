package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func TestPlatformPolicyResource_Schema_EnforcementIsRequired(t *testing.T) {
	r := NewPlatformPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["enforcement"]
	if attr == nil {
		t.Fatal("enforcement attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("enforcement should be required")
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
