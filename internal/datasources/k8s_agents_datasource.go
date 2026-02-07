package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &K8sAgentsDataSource{}

// K8sAgentsDataSource defines the data source implementation.
type K8sAgentsDataSource struct {
	client *client.Client
}

// K8sAgentsDataSourceModel describes the data source data model.
type K8sAgentsDataSourceModel struct {
	Agents []K8sAgentModel `tfsdk:"agents"`
}

// K8sAgentModel describes a single K8s agent in the list.
type K8sAgentModel struct {
	ClusterID   types.String `tfsdk:"cluster_id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	TokenPrefix types.String `tfsdk:"token_prefix"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// NewK8sAgentsDataSource creates a new K8s agents data source.
func NewK8sAgentsDataSource() datasource.DataSource {
	return &K8sAgentsDataSource{}
}

func (d *K8sAgentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_agents"
}

func (d *K8sAgentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all registered Shoehorn K8s agents.",
		Attributes: map[string]schema.Attribute{
			"agents": schema.ListNestedAttribute{
				Description: "The list of K8s agents.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster_id": schema.StringAttribute{
							Description: "The unique cluster identifier.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the cluster.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The agent status (active, inactive, revoked, expired).",
							Computed:    true,
						},
						"token_prefix": schema.StringAttribute{
							Description: "The token prefix for identification.",
							Computed:    true,
						},
						"expires_at": schema.StringAttribute{
							Description: "The expiration timestamp of the agent token.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *K8sAgentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *K8sAgentsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	agents, err := d.client.ListK8sAgents(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading K8s Agents", fmt.Sprintf("Could not list K8s agents: %s", err))
		return
	}

	var state K8sAgentsDataSourceModel
	for _, a := range agents {
		state.Agents = append(state.Agents, K8sAgentModel{
			ClusterID:   types.StringValue(a.ClusterID),
			Name:        types.StringValue(a.Name),
			Status:      types.StringValue(a.Status),
			TokenPrefix: types.StringValue(a.TokenPrefix),
			ExpiresAt:   types.StringValue(a.ExpiresAt),
			CreatedAt:   types.StringValue(a.CreatedAt),
		})
	}

	if state.Agents == nil {
		state.Agents = []K8sAgentModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
