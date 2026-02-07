package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &TeamsDataSource{}

// TeamsDataSource defines the data source implementation.
type TeamsDataSource struct {
	client *client.Client
}

// TeamsDataSourceModel describes the data source data model.
type TeamsDataSourceModel struct {
	Teams []TeamModel `tfsdk:"teams"`
}

// TeamModel describes a single team in the list.
type TeamModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	MemberCount types.Int64  `tfsdk:"member_count"`
}

// NewTeamsDataSource creates a new teams data source.
func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

func (d *TeamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *TeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Shoehorn teams.",
		Attributes: map[string]schema.Attribute{
			"teams": schema.ListNestedAttribute{
				Description: "The list of teams.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the team.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The team name.",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "The team slug.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the team.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The team description.",
							Computed:    true,
						},
						"member_count": schema.Int64Attribute{
							Description: "The number of team members.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	teams, err := d.client.ListTeams(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Teams", fmt.Sprintf("Could not list teams: %s", err))
		return
	}

	var state TeamsDataSourceModel
	for _, t := range teams {
		state.Teams = append(state.Teams, TeamModel{
			ID:          types.StringValue(t.ID),
			Name:        types.StringValue(t.Name),
			Slug:        types.StringValue(t.Slug),
			DisplayName: types.StringValue(t.DisplayName),
			Description: types.StringValue(t.Description),
			MemberCount: types.Int64Value(int64(t.MemberCount)),
		})
	}

	if state.Teams == nil {
		state.Teams = []TeamModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
