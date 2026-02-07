package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestUserRoleResource_Metadata(t *testing.T) {
	r := NewUserRoleResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_user_role" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_user_role")
	}
}

func TestUserRoleResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewUserRoleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "user_id", "role", "email"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestUserRoleResource_Schema_UserIDIsRequired(t *testing.T) {
	r := NewUserRoleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["user_id"]
	if attr == nil {
		t.Fatal("user_id attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("user_id should be required")
	}
}

func TestUserRoleResource_Schema_RoleIsRequired(t *testing.T) {
	r := NewUserRoleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["role"]
	if attr == nil {
		t.Fatal("role attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("role should be required")
	}
}

func TestUserRoleResource_Schema_IDIsComputed(t *testing.T) {
	r := NewUserRoleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	idAttr := resp.Schema.Attributes["id"]
	if idAttr == nil {
		t.Fatal("id attribute not found")
	}
	if !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestUserRoleResource_Schema_EmailIsComputed(t *testing.T) {
	r := NewUserRoleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["email"]
	if attr == nil {
		t.Fatal("email attribute not found")
	}
	if !attr.IsComputed() {
		t.Error("email should be computed")
	}
}

func TestUserRoleResource_Configure_WithValidClient(t *testing.T) {
	r := &UserRoleResource{}
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

func TestUserRoleResource_Configure_NilProviderData(t *testing.T) {
	r := &UserRoleResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Error("should not error on nil provider data")
	}
}

func TestUserRoleResource_Configure_WrongType(t *testing.T) {
	r := &UserRoleResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
