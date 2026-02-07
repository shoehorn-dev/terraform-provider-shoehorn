package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &EntitiesDataSource{}

// EntitiesDataSource defines the data source implementation.
type EntitiesDataSource struct {
	client *client.Client
}

// EntitiesDataSourceModel describes the data source data model.
type EntitiesDataSourceModel struct {
	Entities []EntityModel `tfsdk:"entities"`
}

// EntityModel describes a single entity in the list.
type EntityModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Lifecycle   types.String `tfsdk:"entity_lifecycle"`
	Tier        types.String `tfsdk:"tier"`
}

// NewEntitiesDataSource creates a new entities data source.
func NewEntitiesDataSource() datasource.DataSource {
	return &EntitiesDataSource{}
}

func (d *EntitiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entities"
}

func (d *EntitiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Shoehorn catalog entities.",
		Attributes: map[string]schema.Attribute{
			"entities": schema.ListNestedAttribute{
				Description: "The list of entities.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The service ID of the entity.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the entity.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The entity type (service, library, website, etc.).",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The entity description.",
							Computed:    true,
						},
						"entity_lifecycle": schema.StringAttribute{
							Description: "The entity lifecycle stage.",
							Computed:    true,
						},
						"tier": schema.StringAttribute{
							Description: "The entity tier level.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *EntitiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EntitiesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	entities, err := d.client.ListEntities(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Entities", fmt.Sprintf("Could not list entities: %s", err))
		return
	}

	var state EntitiesDataSourceModel
	for _, e := range entities {
		state.Entities = append(state.Entities, EntityModel{
			ID:          types.StringValue(e.Service.ID),
			Name:        types.StringValue(e.Service.Name),
			Type:        types.StringValue(e.Service.Type),
			Description: types.StringValue(e.Description),
			Lifecycle:   types.StringValue(e.Lifecycle),
			Tier:        types.StringValue(e.Service.Tier),
		})
	}

	if state.Entities == nil {
		state.Entities = []EntityModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
