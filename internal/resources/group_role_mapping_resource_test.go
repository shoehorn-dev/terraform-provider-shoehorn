package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	expectedAttrs := []string{"id", "group_name", "role_name", "auth_provider", "description"}
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

func TestGroupRoleMappingResource_Update_ReturnsError(t *testing.T) {
	r := NewGroupRoleMappingResource().(*GroupRoleMappingResource)
	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Update() should return an error since updates are not supported")
	}
}

func TestGroupRoleMappingResource_ImportState_ValidID(t *testing.T) {
	r := NewGroupRoleMappingResource().(*GroupRoleMappingResource)
	resp := &resource.ImportStateResponse{
		State: newGroupRoleMappingState(),
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "platform-team:tenant:admin"}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() unexpected error: %v", resp.Diagnostics)
	}

	var groupName, roleName string
	resp.State.GetAttribute(context.Background(), path.Root("group_name"), &groupName)
	resp.State.GetAttribute(context.Background(), path.Root("role_name"), &roleName)

	if groupName != "platform-team" {
		t.Errorf("group_name = %q, want %q", groupName, "platform-team")
	}
	if roleName != "tenant:admin" {
		t.Errorf("role_name = %q, want %q", roleName, "tenant:admin")
	}
}

func TestGroupRoleMappingResource_ImportState_InvalidID_NoColon(t *testing.T) {
	r := NewGroupRoleMappingResource().(*GroupRoleMappingResource)
	resp := &resource.ImportStateResponse{
		State: newGroupRoleMappingState(),
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "invalid-no-colon"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ImportState() should return error for ID without colon")
	}
}

func TestGroupRoleMappingResource_ImportState_InvalidID_Empty(t *testing.T) {
	r := NewGroupRoleMappingResource().(*GroupRoleMappingResource)
	resp := &resource.ImportStateResponse{
		State: newGroupRoleMappingState(),
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: ":"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ImportState() should return error for empty parts")
	}
}

func TestGroupRoleMappingResource_ImportState_EmptyString(t *testing.T) {
	r := NewGroupRoleMappingResource().(*GroupRoleMappingResource)
	resp := &resource.ImportStateResponse{
		State: newGroupRoleMappingState(),
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: ""}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ImportState() should return error for empty ID")
	}
}

// groupRoleMappingStateType is the tftypes representation of the resource schema.
var groupRoleMappingStateType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"group_name":  tftypes.String,
		"role_name":   tftypes.String,
		"auth_provider": tftypes.String,
		"description": tftypes.String,
	},
}

// getGroupRoleMappingSchema returns the schema for use in test state objects.
func getGroupRoleMappingSchema() schema.Schema {
	r := NewGroupRoleMappingResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	return resp.Schema
}

// newGroupRoleMappingState returns a tfsdk.State initialized for import tests.
func newGroupRoleMappingState() tfsdk.State {
	return tfsdk.State{
		Schema: getGroupRoleMappingSchema(),
		Raw:    tftypes.NewValue(groupRoleMappingStateType, nil),
	}
}
