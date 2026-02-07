package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestAPIKeyResource_Metadata(t *testing.T) {
	r := NewAPIKeyResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_api_key" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_api_key")
	}
}

func TestAPIKeyResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewAPIKeyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "name", "description", "scopes", "expires_in_days", "key_prefix", "raw_key", "expires_at", "created_at"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestAPIKeyResource_Schema_NameIsRequired(t *testing.T) {
	r := NewAPIKeyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	nameAttr := resp.Schema.Attributes["name"]
	if nameAttr == nil {
		t.Fatal("name attribute not found")
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestAPIKeyResource_Schema_ScopesIsRequired(t *testing.T) {
	r := NewAPIKeyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	scopesAttr := resp.Schema.Attributes["scopes"]
	if scopesAttr == nil {
		t.Fatal("scopes attribute not found")
	}
	if !scopesAttr.IsRequired() {
		t.Error("scopes should be required")
	}
}

func TestAPIKeyResource_Schema_RawKeyIsSensitive(t *testing.T) {
	r := NewAPIKeyResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	rawKeyAttr := resp.Schema.Attributes["raw_key"]
	if rawKeyAttr == nil {
		t.Fatal("raw_key attribute not found")
	}
	if !rawKeyAttr.IsSensitive() {
		t.Error("raw_key should be sensitive")
	}
}

func TestAPIKeyResource_Schema_IDIsComputed(t *testing.T) {
	r := NewAPIKeyResource()
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

func TestAPIKeyResource_Configure_WithValidClient(t *testing.T) {
	r := &APIKeyResource{}
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

func TestAPIKeyResource_Configure_WrongType(t *testing.T) {
	r := &APIKeyResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
