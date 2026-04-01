package datasources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestGitOpsResourcesDataSource_Metadata(t *testing.T) {
	d := NewGitOpsResourcesDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_gitops_resources" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_gitops_resources")
	}
}

func TestGitOpsResourcesDataSource_Schema_HasExpectedAttributes(t *testing.T) {
	d := NewGitOpsResourcesDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"cluster_id", "tool", "sync_status", "health_status", "total", "resources",
	}
	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("schema missing %q attribute", name)
		}
	}
}

func TestGitOpsResourcesDataSource_Configure_WithValidClient(t *testing.T) {
	d := &GitOpsResourcesDataSource{}
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

func TestGitOpsResourcesDataSource_Configure_WrongType(t *testing.T) {
	d := &GitOpsResourcesDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
