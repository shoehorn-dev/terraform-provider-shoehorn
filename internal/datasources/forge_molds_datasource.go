package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &ForgeMoldsDataSource{}

// ForgeMoldsDataSource defines the data source implementation.
type ForgeMoldsDataSource struct {
	client *client.Client
}

// ForgeMoldsDataSourceModel describes the data source data model.
type ForgeMoldsDataSourceModel struct {
	Molds []ForgeMoldModel `tfsdk:"molds"`
}

// ForgeMoldModel describes a single forge mold in the data source.
type ForgeMoldModel struct {
	Slug        types.String `tfsdk:"slug"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Version     types.String `tfsdk:"version"`
	Visibility  types.String `tfsdk:"visibility"`
	Category    types.String `tfsdk:"category"`
	Tags        types.List   `tfsdk:"tags"`
}

// NewForgeMoldsDataSource creates a new forge molds data source.
func NewForgeMoldsDataSource() datasource.DataSource {
	return &ForgeMoldsDataSource{}
}

func (d *ForgeMoldsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forge_molds"
}

func (d *ForgeMoldsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves all available Forge mold templates.",
		Attributes: map[string]schema.Attribute{
			"molds": schema.ListNestedAttribute{
				Description: "The list of forge molds.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"slug": schema.StringAttribute{
							Description: "The slug identifier of the forge mold.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the forge mold.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of the forge mold.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "The version of the forge mold.",
							Computed:    true,
						},
						"visibility": schema.StringAttribute{
							Description: "The visibility of the forge mold (public, tenant, private).",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							Description: "The category of the forge mold.",
							Computed:    true,
						},
						"tags": schema.ListAttribute{
							Description: "Tags for categorizing the forge mold.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *ForgeMoldsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ForgeMoldsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "reading forge molds data source")

	molds, err := d.client.ListForgeMolds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Forge Molds", fmt.Sprintf("Could not list forge molds: %s", err))
		return
	}

	state := ForgeMoldsDataSourceModel{}

	for _, m := range molds {
		model := ForgeMoldModel{
			Slug:       types.StringValue(m.Slug),
			Name:       types.StringValue(m.Name),
			Version:    types.StringValue(m.Version),
			Visibility: types.StringValue(m.Visibility),
			Category:   types.StringValue(m.Category),
		}

		if m.Description != "" {
			model.Description = types.StringValue(m.Description)
		} else {
			model.Description = types.StringNull()
		}

		if len(m.Tags) > 0 {
			tagValues := make([]types.String, len(m.Tags))
			for i, t := range m.Tags {
				tagValues[i] = types.StringValue(t)
			}
			listVal, diags := types.ListValueFrom(ctx, types.StringType, tagValues)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			model.Tags = listVal
		} else {
			model.Tags = types.ListNull(types.StringType)
		}

		state.Molds = append(state.Molds, model)
	}

	if state.Molds == nil {
		state.Molds = []ForgeMoldModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
