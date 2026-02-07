package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &FeatureFlagsDataSource{}

// FeatureFlagsDataSource defines the data source implementation.
type FeatureFlagsDataSource struct {
	client *client.Client
}

// FeatureFlagsDataSourceModel describes the data source data model.
type FeatureFlagsDataSourceModel struct {
	FeatureFlags []FeatureFlagModel `tfsdk:"feature_flags"`
}

// FeatureFlagModel describes a single feature flag in the list.
type FeatureFlagModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

// NewFeatureFlagsDataSource creates a new feature flags data source.
func NewFeatureFlagsDataSource() datasource.DataSource {
	return &FeatureFlagsDataSource{}
}

func (d *FeatureFlagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feature_flags"
}

func (d *FeatureFlagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Shoehorn feature flags.",
		Attributes: map[string]schema.Attribute{
			"feature_flags": schema.ListNestedAttribute{
				Description: "The list of feature flags.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the feature flag.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "The unique key of the feature flag.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the feature flag.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The feature flag description.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the feature flag is enabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *FeatureFlagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FeatureFlagsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	flags, err := d.client.ListFeatureFlags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Feature Flags", fmt.Sprintf("Could not list feature flags: %s", err))
		return
	}

	var state FeatureFlagsDataSourceModel
	for _, f := range flags {
		state.FeatureFlags = append(state.FeatureFlags, FeatureFlagModel{
			ID:          types.StringValue(f.ID),
			Key:         types.StringValue(f.Key),
			Name:        types.StringValue(f.Name),
			Description: types.StringValue(f.Description),
			Enabled:     types.BoolValue(f.DefaultEnabled),
		})
	}

	if state.FeatureFlags == nil {
		state.FeatureFlags = []FeatureFlagModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
