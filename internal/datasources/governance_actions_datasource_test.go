package datasources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestGovernanceActionsDataSource_Metadata(t *testing.T) {
	d := NewGovernanceActionsDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_governance_actions" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_governance_actions")
	}
}

func TestGovernanceActionsDataSource_Schema_HasExpectedAttributes(t *testing.T) {
	d := NewGovernanceActionsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"status", "priority", "entity_id", "source_type", "overdue", "total", "actions",
	}
	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("schema missing %q attribute", name)
		}
	}
}

func TestGovernanceActionsDataSource_Configure_WithValidClient(t *testing.T) {
	d := &GovernanceActionsDataSource{}
	c := client.NewClient("https://test.example.com", "key", 30*time.Second)

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: c,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}
	if d.client != c {
		t.Error("client not set correctly")
	}
}

func TestGovernanceActionsDataSource_Configure_WrongType(t *testing.T) {
	d := &GovernanceActionsDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
