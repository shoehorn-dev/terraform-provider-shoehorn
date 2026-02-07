package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestK8sAgentResource_Metadata(t *testing.T) {
	r := NewK8sAgentResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_k8s_agent" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_k8s_agent")
	}
}

func TestK8sAgentResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewK8sAgentResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "cluster_id", "name", "description", "expires_in_days", "token", "token_prefix", "status", "expires_at", "created_at"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestK8sAgentResource_Schema_ClusterIDIsRequired(t *testing.T) {
	r := NewK8sAgentResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["cluster_id"]
	if attr == nil {
		t.Fatal("cluster_id attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("cluster_id should be required")
	}
}

func TestK8sAgentResource_Schema_NameIsRequired(t *testing.T) {
	r := NewK8sAgentResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["name"]
	if attr == nil {
		t.Fatal("name attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestK8sAgentResource_Schema_TokenIsSensitive(t *testing.T) {
	r := NewK8sAgentResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["token"]
	if attr == nil {
		t.Fatal("token attribute not found")
	}
	if !attr.IsSensitive() {
		t.Error("token should be sensitive")
	}
	if !attr.IsComputed() {
		t.Error("token should be computed")
	}
}

func TestK8sAgentResource_Schema_IDIsComputed(t *testing.T) {
	r := NewK8sAgentResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["id"]
	if attr == nil {
		t.Fatal("id attribute not found")
	}
	if !attr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestK8sAgentResource_Configure_WithValidClient(t *testing.T) {
	r := &K8sAgentResource{}
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

func TestK8sAgentResource_Configure_NilProviderData(t *testing.T) {
	r := &K8sAgentResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Error("should not error on nil provider data")
	}
}

func TestK8sAgentResource_Configure_WrongType(t *testing.T) {
	r := &K8sAgentResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}
