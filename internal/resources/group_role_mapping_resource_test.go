package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestGroupRoleMappingResource_Metadata(t *testing.T) {
	r := NewGroupRoleMappingResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_group_role_mapping" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_group_role_mapping")
	}
}

func TestGroupRoleMappingResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewGroupRoleMappingResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "group_name", "role_name", "provider", "description"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestGroupRoleMappingResource_Schema_GroupNameIsRequired(t *testing.T) {
	r := NewGroupRoleMappingResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["group_name"]
	if !ok {
		t.Fatal("group_name attribute not found")
	}
	strAttr, ok := attr.(interface{ IsRequired() bool })
	if !ok {
		t.Fatal("group_name does not implement IsRequired")
	}
	if !strAttr.IsRequired() {
		t.Error("group_name should be required")
	}
}

func TestGroupRoleMappingResource_Schema_RoleNameIsRequired(t *testing.T) {
	r := NewGroupRoleMappingResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["role_name"]
	if !ok {
		t.Fatal("role_name attribute not found")
	}
	strAttr, ok := attr.(interface{ IsRequired() bool })
	if !ok {
		t.Fatal("role_name does not implement IsRequired")
	}
	if !strAttr.IsRequired() {
		t.Error("role_name should be required")
	}
}

func TestGroupRoleMappingResource_Configure_SetsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	c := client.NewClient(server.URL, "key", 30*time.Second)
	r := NewGroupRoleMappingResource().(*GroupRoleMappingResource)

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: c}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() unexpected error: %v", resp.Diagnostics)
	}
	if r.client == nil {
		t.Error("Configure() did not set client")
	}
}

func TestGroupRoleMappingResource_Description(t *testing.T) {
	r := NewGroupRoleMappingResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	if resp.Schema.Description == "" {
		t.Error("schema description should not be empty")
	}
}
