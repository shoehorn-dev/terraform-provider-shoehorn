package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestIntegrationResource_Metadata(t *testing.T) {
	r := NewIntegrationResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_integration" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_integration")
	}
}

func TestIntegrationResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewIntegrationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "name", "type", "status", "config_json", "team_id", "created_at", "updated_at"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestIntegrationResource_Schema_NameIsRequired(t *testing.T) {
	r := NewIntegrationResource()
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

func TestIntegrationResource_Schema_TypeIsRequired(t *testing.T) {
	r := NewIntegrationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["type"]
	if attr == nil {
		t.Fatal("type attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("type should be required")
	}
}

func TestIntegrationResource_Schema_ConfigJSONIsRequired(t *testing.T) {
	r := NewIntegrationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["config_json"]
	if attr == nil {
		t.Fatal("config_json attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("config_json should be required")
	}
	if !attr.IsSensitive() {
		t.Error("config_json should be sensitive")
	}
}

func TestIntegrationResource_Schema_IDIsComputed(t *testing.T) {
	r := NewIntegrationResource()
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

func TestIntegrationResource_Configure_WithValidClient(t *testing.T) {
	r := &IntegrationResource{}
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

func TestIntegrationResource_Configure_WrongType(t *testing.T) {
	r := &IntegrationResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapIntegrationToState(t *testing.T) {
	integration := &client.Integration{
		ID:        1,
		Name:      "GitHub Prod",
		Type:      "github",
		Status:    "active",
		TeamID:    "team-1",
		CreatedAt: "2025-01-15T10:00:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &IntegrationResourceModel{}
	mapIntegrationToState(integration, state)

	if state.ID.ValueString() != "1" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "1")
	}
	if state.Name.ValueString() != "GitHub Prod" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "GitHub Prod")
	}
	if state.Type.ValueString() != "github" {
		t.Errorf("Type = %q, want %q", state.Type.ValueString(), "github")
	}
	if state.Status.ValueString() != "active" {
		t.Errorf("Status = %q, want %q", state.Status.ValueString(), "active")
	}
	if state.TeamID.ValueString() != "team-1" {
		t.Errorf("TeamID = %q, want %q", state.TeamID.ValueString(), "team-1")
	}
}

func TestMapIntegrationToState_EmptyOptionalFields(t *testing.T) {
	integration := &client.Integration{
		ID:   2,
		Name: "Slack",
		Type: "slack",
	}

	state := &IntegrationResourceModel{}
	mapIntegrationToState(integration, state)

	if state.ID.ValueString() != "2" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "2")
	}
	if state.Name.ValueString() != "Slack" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "Slack")
	}
}
