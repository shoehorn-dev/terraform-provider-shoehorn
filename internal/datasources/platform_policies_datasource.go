package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &PlatformPoliciesDataSource{}

// PlatformPoliciesDataSource defines the data source implementation.
type PlatformPoliciesDataSource struct {
	client *client.Client
}

// PlatformPoliciesDataSourceModel describes the data source data model.
type PlatformPoliciesDataSourceModel struct {
	Policies []PlatformPolicyModel `tfsdk:"policies"`
}

// PlatformPolicyModel describes a single platform policy in the list.
type PlatformPolicyModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Category    types.String `tfsdk:"category"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Enforcement types.String `tfsdk:"enforcement"`
	System      types.Bool   `tfsdk:"system"`
}

// NewPlatformPoliciesDataSource creates a new platform policies data source.
func NewPlatformPoliciesDataSource() datasource.DataSource {
	return &PlatformPoliciesDataSource{}
}

func (d *PlatformPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_platform_policies"
}

func (d *PlatformPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Shoehorn platform policies.",
		Attributes: map[string]schema.Attribute{
			"policies": schema.ListNestedAttribute{
				Description: "The list of platform policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the policy.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "The unique key of the policy.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the policy.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The policy description.",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							Description: "The policy category.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the policy is enabled.",
							Computed:    true,
						},
						"enforcement": schema.StringAttribute{
							Description: "The enforcement level (warn, block, audit).",
							Computed:    true,
						},
						"system": schema.BoolAttribute{
							Description: "Whether this is a system policy.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *PlatformPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PlatformPoliciesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	policies, err := d.client.ListPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Platform Policies", fmt.Sprintf("Could not list platform policies: %s", err))
		return
	}

	var state PlatformPoliciesDataSourceModel
	for _, p := range policies {
		state.Policies = append(state.Policies, PlatformPolicyModel{
			ID:          types.StringValue(p.ID),
			Key:         types.StringValue(p.Key),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			Category:    types.StringValue(p.Category),
			Enabled:     types.BoolValue(p.Enabled),
			Enforcement: types.StringValue(p.Enforcement),
			System:      types.BoolValue(p.System),
		})
	}

	if state.Policies == nil {
		state.Policies = []PlatformPolicyModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
