package datasources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestPlatformPoliciesDataSource_Metadata(t *testing.T) {
	d := NewPlatformPoliciesDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_platform_policies" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_platform_policies")
	}
}

func TestPlatformPoliciesDataSource_Schema_HasPoliciesAttribute(t *testing.T) {
	d := NewPlatformPoliciesDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["policies"]; !ok {
		t.Error("schema missing 'policies' attribute")
	}
}

func TestPlatformPoliciesDataSource_Configure_WithValidClient(t *testing.T) {
	d := &PlatformPoliciesDataSource{}
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

func TestPlatformPoliciesDataSource_Configure_WrongType(t *testing.T) {
	d := &PlatformPoliciesDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

// Reading `configurable` separates the settings from the always-on protections.
func TestPlatformPoliciesDataSource_Schema_PoliciesCarryConfigurable(t *testing.T) {
	d := NewPlatformPoliciesDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	policies, ok := resp.Schema.Attributes["policies"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("policies is not a list of nested objects")
	}
	attr, ok := policies.NestedObject.Attributes["configurable"]
	if !ok {
		t.Fatal("a policy has no 'configurable' attribute")
	}
	if !attr.IsComputed() {
		t.Error("configurable should be computed")
	}
}

func TestPlatformPoliciesDataSource_Read_CarriesConfigurable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"policies":[
			{"id":"governance-auto-actions","key":"governance-auto-actions","name":"Governance actions","enabled":true,"system":false,"configurable":true},
			{"id":"tenant-isolation","key":"tenant-isolation","name":"Tenant isolation","enabled":true,"system":true,"configurable":false}
		]}`))
	}))
	defer server.Close()

	policies, err := client.NewClient(server.URL, "key", 30*time.Second).ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListPolicies() error = %v", err)
	}

	var state PlatformPoliciesDataSourceModel
	for _, p := range policies {
		state.Policies = append(state.Policies, platformPolicyModel(p))
	}
	if !state.Policies[0].Configurable.ValueBool() {
		t.Error("governance-auto-actions: Configurable = false, want true")
	}
	if state.Policies[1].Configurable.ValueBool() {
		t.Error("tenant-isolation: Configurable = true, want false")
	}
}
