package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/datasources"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/resources"
)

var _ provider.Provider = &ShoehornProvider{}

// ShoehornProvider defines the provider implementation.
type ShoehornProvider struct {
	version string
}

// ShoehornProviderModel describes the provider data model.
type ShoehornProviderModel struct {
	Host    types.String `tfsdk:"host"`
	APIKey  types.String `tfsdk:"api_key"`
	Timeout types.Int64  `tfsdk:"timeout"`
}

// New returns a function that creates the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShoehornProvider{
			version: version,
		}
	}
}

func (p *ShoehornProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shoehorn"
	resp.Version = p.version
}

func (p *ShoehornProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for configuring Shoehorn Internal Developer Portal resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Shoehorn API host URL. Can also be set with the SHOEHORN_HOST environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Shoehorn API key for authentication. Can also be set with the SHOEHORN_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"timeout": schema.Int64Attribute{
				Description: "HTTP request timeout in seconds. Defaults to 30.",
				Optional:    true,
			},
		},
	}
}

func (p *ShoehornProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ShoehornProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve host
	host := os.Getenv("SHOEHORN_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if host == "" {
		resp.Diagnostics.AddError(
			"Missing Host",
			"The provider requires a host to be set. Set the 'host' attribute or the SHOEHORN_HOST environment variable.",
		)
	}

	// Resolve API key
	apiKey := os.Getenv("SHOEHORN_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires an API key. Set the 'api_key' attribute or the SHOEHORN_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve timeout
	timeout := 30 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	// Create client
	c := client.NewClient(host, apiKey, timeout)

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *ShoehornProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewTeamResource,
		resources.NewEntityResource,
		resources.NewFeatureFlagResource,
		resources.NewTenantSettingsResource,
		resources.NewAPIKeyResource,
		resources.NewUserRoleResource,
		resources.NewIntegrationResource,
		resources.NewK8sAgentResource,
		resources.NewPlatformPolicyResource,
		resources.NewGroupRoleMappingResource,
	}
}

func (p *ShoehornProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewEntitiesDataSource,
		datasources.NewTeamsDataSource,
		datasources.NewFeatureFlagsDataSource,
		datasources.NewIntegrationsDataSource,
		datasources.NewAPIKeysDataSource,
		datasources.NewK8sAgentsDataSource,
		datasources.NewPlatformPoliciesDataSource,
		datasources.NewUsersDataSource,
		datasources.NewGroupsDataSource,
	}
}
