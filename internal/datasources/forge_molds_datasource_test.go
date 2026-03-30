package datasources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestForgeMoldsDataSource_Metadata(t *testing.T) {
	d := NewForgeMoldsDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_forge_molds" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_forge_molds")
	}
}

func TestForgeMoldsDataSource_Schema_HasExpectedAttributes(t *testing.T) {
	d := NewForgeMoldsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	expectedAttrs := []string{"molds"}
	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("schema missing %q attribute", name)
		}
	}
}

func TestForgeMoldsDataSource_Configure_WithValidClient(t *testing.T) {
	d := &ForgeMoldsDataSource{}
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

func TestForgeMoldsDataSource_Configure_WrongType(t *testing.T) {
	d := &ForgeMoldsDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
