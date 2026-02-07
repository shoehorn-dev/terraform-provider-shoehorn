package datasources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestAPIKeysDataSource_Metadata(t *testing.T) {
	d := NewAPIKeysDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_api_keys" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_api_keys")
	}
}

func TestAPIKeysDataSource_Schema_HasAPIKeysAttribute(t *testing.T) {
	d := NewAPIKeysDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["api_keys"]; !ok {
		t.Error("schema missing 'api_keys' attribute")
	}
}

func TestAPIKeysDataSource_Configure_WithValidClient(t *testing.T) {
	d := &APIKeysDataSource{}
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

func TestAPIKeysDataSource_Configure_WrongType(t *testing.T) {
	d := &APIKeysDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
