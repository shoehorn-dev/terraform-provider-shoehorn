package datasources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestEntitiesDataSource_Metadata(t *testing.T) {
	d := NewEntitiesDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_entities" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_entities")
	}
}

func TestEntitiesDataSource_Schema_HasEntitiesAttribute(t *testing.T) {
	d := NewEntitiesDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["entities"]; !ok {
		t.Error("schema missing 'entities' attribute")
	}
}

func TestEntitiesDataSource_Configure_WithValidClient(t *testing.T) {
	d := &EntitiesDataSource{}
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

func TestEntitiesDataSource_Configure_WrongType(t *testing.T) {
	d := &EntitiesDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestEntitiesDataSource_Configure_NilProviderData(t *testing.T) {
	d := &EntitiesDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Error("should not error on nil provider data")
	}
}
