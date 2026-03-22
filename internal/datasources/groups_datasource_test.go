package datasources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestGroupsDataSource_Metadata(t *testing.T) {
	d := NewGroupsDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_groups" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_groups")
	}
}

func TestGroupsDataSource_Schema_HasGroupsAttribute(t *testing.T) {
	d := NewGroupsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["groups"]; !ok {
		t.Error("schema missing 'groups' attribute")
	}
}

func TestGroupsDataSource_Schema_Description(t *testing.T) {
	d := NewGroupsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Schema.Description == "" {
		t.Error("schema description should not be empty")
	}
}

func TestGroupsDataSource_Configure_SetsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	c := client.NewClient(server.URL, "key", 30*time.Second)
	d := NewGroupsDataSource().(*GroupsDataSource)

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: c}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() unexpected error: %v", resp.Diagnostics)
	}
	if d.client == nil {
		t.Error("Configure() did not set client")
	}
}
