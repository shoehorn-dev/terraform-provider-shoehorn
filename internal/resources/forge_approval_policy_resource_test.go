package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestForgeApprovalPolicyResource_Metadata(t *testing.T) {
	r := NewForgeApprovalPolicyResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_forge_approval_policy" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_forge_approval_policy")
	}
}

func TestForgeApprovalPolicyResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewForgeApprovalPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{
		"id", "name", "description", "enabled", "steps", "created_at", "updated_at",
	}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestForgeApprovalPolicyResource_Schema_NameIsRequired(t *testing.T) {
	r := NewForgeApprovalPolicyResource()
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

func TestForgeApprovalPolicyResource_Schema_StepsIsRequired(t *testing.T) {
	r := NewForgeApprovalPolicyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["steps"]
	if attr == nil {
		t.Fatal("steps attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("steps should be required")
	}
}

func TestForgeApprovalPolicyResource_Schema_IDIsComputed(t *testing.T) {
	r := NewForgeApprovalPolicyResource()
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

func TestForgeApprovalPolicyResource_Configure_WithValidClient(t *testing.T) {
	r := &ForgeApprovalPolicyResource{}
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

func TestForgeApprovalPolicyResource_Configure_WrongType(t *testing.T) {
	r := &ForgeApprovalPolicyResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapApprovalPolicyToState(t *testing.T) {
	policy := &client.ForgeApprovalPolicy{
		ID:          "pol-123",
		Name:        "Production Deploy",
		Description: "Requires manager approval",
		Enabled:     true,
		ApprovalChain: []client.ApprovalStep{
			{
				Name:          "Manager Review",
				Approvers:     []string{"user-1", "user-2"},
				RequiredCount: 0,
			},
			{
				Name:          "Security Review",
				Approvers:     []string{"sec-team"},
				RequiredCount: 1,
			},
		},
		CreatedAt: "2025-01-15T10:00:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &ForgeApprovalPolicyResourceModel{}
	diags := mapApprovalPolicyToState(context.Background(), policy, state)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != "pol-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "pol-123")
	}
	if state.Name.ValueString() != "Production Deploy" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "Production Deploy")
	}
	if state.Description.ValueString() != "Requires manager approval" {
		t.Errorf("Description = %q, want %q", state.Description.ValueString(), "Requires manager approval")
	}
	if state.Enabled.ValueBool() != true {
		t.Errorf("Enabled = %v, want %v", state.Enabled.ValueBool(), true)
	}
	if state.ApprovalChain.IsNull() {
		t.Fatal("ApprovalChain should not be null")
	}
	if len(state.ApprovalChain.Elements()) != 2 {
		t.Errorf("ApprovalChain length = %d, want 2", len(state.ApprovalChain.Elements()))
	}
}

func TestMapApprovalPolicyToState_EmptyOptionalFields(t *testing.T) {
	policy := &client.ForgeApprovalPolicy{
		ID:      "pol-456",
		Name:    "Simple Policy",
		Enabled: false,
		ApprovalChain: []client.ApprovalStep{
			{
				Name:      "Approval",
				Approvers: []string{"admin"},
			},
		},
	}

	state := &ForgeApprovalPolicyResourceModel{}
	diags := mapApprovalPolicyToState(context.Background(), policy, state)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != "pol-456" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "pol-456")
	}
	if !state.Description.IsNull() {
		t.Errorf("Description should be null when empty, got %q", state.Description.ValueString())
	}
	if state.Enabled.ValueBool() != false {
		t.Errorf("Enabled = %v, want %v", state.Enabled.ValueBool(), false)
	}
}
