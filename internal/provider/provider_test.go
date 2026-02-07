package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNew_ReturnsProvider(t *testing.T) {
	p := New("test")()
	if p == nil {
		t.Fatal("New() returned nil provider")
	}
}

func TestProvider_Metadata(t *testing.T) {
	p := &ShoehornProvider{version: "1.0.0"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "shoehorn" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn")
	}
	if resp.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", resp.Version, "1.0.0")
	}
}

func TestProvider_Schema_HasRequiredAttributes(t *testing.T) {
	p := &ShoehornProvider{version: "test"}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	requiredAttrs := []string{"host", "api_key", "timeout"}
	for _, name := range requiredAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestProvider_Schema_APIKeyIsSensitive(t *testing.T) {
	p := &ShoehornProvider{version: "test"}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	apiKeyAttr := resp.Schema.Attributes["api_key"]
	if apiKeyAttr == nil {
		t.Fatal("api_key attribute not found in schema")
	}
	if !apiKeyAttr.IsSensitive() {
		t.Error("api_key should be marked as sensitive")
	}
}

func newTestConfigValue(host, apiKey *string, timeout *int64) tftypes.Value {
	hostVal := tftypes.NewValue(tftypes.String, nil)
	if host != nil {
		hostVal = tftypes.NewValue(tftypes.String, *host)
	}

	apiKeyVal := tftypes.NewValue(tftypes.String, nil)
	if apiKey != nil {
		apiKeyVal = tftypes.NewValue(tftypes.String, *apiKey)
	}

	timeoutVal := tftypes.NewValue(tftypes.Number, nil)
	if timeout != nil {
		timeoutVal = tftypes.NewValue(tftypes.Number, timeout)
	}

	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"host":    tftypes.String,
			"api_key": tftypes.String,
			"timeout": tftypes.Number,
		},
	}, map[string]tftypes.Value{
		"host":    hostVal,
		"api_key": apiKeyVal,
		"timeout": timeoutVal,
	})
}

func TestProvider_Configure_MissingHost_ReturnsError(t *testing.T) {
	t.Setenv("SHOEHORN_HOST", "")
	t.Setenv("SHOEHORN_API_KEY", "test-key")

	p := &ShoehornProvider{version: "test"}

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	rawConfig := newTestConfigValue(nil, nil, nil)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    rawConfig,
		},
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for missing host, got none")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Missing Host" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Missing Host' error diagnostic")
	}
}

func TestProvider_Configure_MissingAPIKey_ReturnsError(t *testing.T) {
	t.Setenv("SHOEHORN_HOST", "https://test.example.com")
	t.Setenv("SHOEHORN_API_KEY", "")

	p := &ShoehornProvider{version: "test"}

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	rawConfig := newTestConfigValue(nil, nil, nil)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    rawConfig,
		},
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for missing api_key, got none")
	}

	found := false
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Missing API Key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Missing API Key' error diagnostic")
	}
}

func TestProvider_Configure_WithEnvVars_Succeeds(t *testing.T) {
	t.Setenv("SHOEHORN_HOST", "https://test.example.com")
	t.Setenv("SHOEHORN_API_KEY", "shp_svc_testkey")

	p := &ShoehornProvider{version: "test"}

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	rawConfig := newTestConfigValue(nil, nil, nil)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    rawConfig,
		},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}

	if resp.ResourceData == nil {
		t.Error("ResourceData should be set after successful configure")
	}
	if resp.DataSourceData == nil {
		t.Error("DataSourceData should be set after successful configure")
	}
}

func TestProvider_Configure_ExplicitConfig_OverridesEnv(t *testing.T) {
	t.Setenv("SHOEHORN_HOST", "https://env.example.com")
	t.Setenv("SHOEHORN_API_KEY", "env_key")

	p := &ShoehornProvider{version: "test"}

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	host := "https://explicit.example.com"
	apiKey := "explicit_key"
	rawConfig := newTestConfigValue(&host, &apiKey, nil)

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    rawConfig,
		},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}

	if resp.ResourceData == nil {
		t.Error("ResourceData should be set after successful configure")
	}
}

// testAccProtoV6ProviderFactories creates provider factories for acceptance testing.
func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"shoehorn": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("SHOEHORN_HOST"); v == "" {
		t.Fatal("SHOEHORN_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("SHOEHORN_API_KEY"); v == "" {
		t.Fatal("SHOEHORN_API_KEY must be set for acceptance tests")
	}
}
