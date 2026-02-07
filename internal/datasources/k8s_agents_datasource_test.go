package datasources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestK8sAgentsDataSource_Metadata(t *testing.T) {
	d := NewK8sAgentsDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_k8s_agents" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_k8s_agents")
	}
}

func TestK8sAgentsDataSource_Schema_HasAgentsAttribute(t *testing.T) {
	d := NewK8sAgentsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["agents"]; !ok {
		t.Error("schema missing 'agents' attribute")
	}
}

func TestK8sAgentsDataSource_Configure_WithValidClient(t *testing.T) {
	d := &K8sAgentsDataSource{}
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

func TestK8sAgentsDataSource_Configure_WrongType(t *testing.T) {
	d := &K8sAgentsDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
