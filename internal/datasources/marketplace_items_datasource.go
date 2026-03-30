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

var _ datasource.DataSource = &MarketplaceItemsDataSource{}

// MarketplaceItemsDataSource defines the data source implementation.
type MarketplaceItemsDataSource struct {
	client *client.Client
}

// MarketplaceItemsDataSourceModel describes the data source data model.
type MarketplaceItemsDataSourceModel struct {
	Kind     types.String            `tfsdk:"kind"`
	Category types.String            `tfsdk:"category"`
	Items    []MarketplaceItemModel  `tfsdk:"items"`
}

// MarketplaceItemModel describes a single marketplace item.
type MarketplaceItemModel struct {
	Slug        types.String `tfsdk:"slug"`
	Kind        types.String `tfsdk:"kind"`
	Name        types.String `tfsdk:"name"`
	Version     types.String `tfsdk:"version"`
	Description types.String `tfsdk:"description"`
	AuthorName  types.String `tfsdk:"author_name"`
	Category    types.String `tfsdk:"category"`
	Tier        types.String `tfsdk:"tier"`
	Verified    types.Bool   `tfsdk:"verified"`
	Featured    types.Bool   `tfsdk:"featured"`
}

// NewMarketplaceItemsDataSource creates a new marketplace items data source.
func NewMarketplaceItemsDataSource() datasource.DataSource {
	return &MarketplaceItemsDataSource{}
}

func (d *MarketplaceItemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marketplace_items"
}

func (d *MarketplaceItemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves available items from the Shoehorn marketplace catalog.",
		Attributes: map[string]schema.Attribute{
			"kind": schema.StringAttribute{
				Description: "Filter items by kind.",
				Optional:    true,
			},
			"category": schema.StringAttribute{
				Description: "Filter items by category.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of marketplace items.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"slug": schema.StringAttribute{
							Description: "The unique slug of the marketplace item.",
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Description: "The kind of marketplace item.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the marketplace item.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "The current version of the marketplace item.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the marketplace item.",
							Computed:    true,
						},
						"author_name": schema.StringAttribute{
							Description: "The author of the marketplace item.",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							Description: "The category of the marketplace item.",
							Computed:    true,
						},
						"tier": schema.StringAttribute{
							Description: "The tier of the marketplace item (free, premium, enterprise).",
							Computed:    true,
						},
						"verified": schema.BoolAttribute{
							Description: "Whether the marketplace item is verified.",
							Computed:    true,
						},
						"featured": schema.BoolAttribute{
							Description: "Whether the marketplace item is featured.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *MarketplaceItemsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MarketplaceItemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "reading marketplace items data source")

	var config MarketplaceItemsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var kind, category string
	if !config.Kind.IsNull() && !config.Kind.IsUnknown() {
		kind = config.Kind.ValueString()
	}
	if !config.Category.IsNull() && !config.Category.IsUnknown() {
		category = config.Category.ValueString()
	}

	items, err := d.client.ListMarketplaceItems(ctx, kind, category)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Marketplace Items", fmt.Sprintf("Could not list marketplace items: %s", err))
		return
	}

	state := MarketplaceItemsDataSourceModel{
		Kind:     config.Kind,
		Category: config.Category,
	}

	for _, item := range items {
		state.Items = append(state.Items, MarketplaceItemModel{
			Slug:        types.StringValue(item.Slug),
			Kind:        types.StringValue(item.Kind),
			Name:        types.StringValue(item.Name),
			Version:     types.StringValue(item.Version),
			Description: types.StringValue(item.Description),
			AuthorName:  types.StringValue(item.AuthorName),
			Category:    types.StringValue(item.Category),
			Tier:        types.StringValue(item.Tier),
			Verified:    types.BoolValue(item.Verified),
			Featured:    types.BoolValue(item.Featured),
		})
	}

	if state.Items == nil {
		state.Items = []MarketplaceItemModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
