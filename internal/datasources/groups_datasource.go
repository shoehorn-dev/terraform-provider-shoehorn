package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &GroupsDataSource{}

// GroupsDataSource defines the data source implementation.
type GroupsDataSource struct {
	client *client.Client
}

// GroupsDataSourceModel describes the data source data model.
type GroupsDataSourceModel struct {
	Groups []GroupModel `tfsdk:"groups"`
}

// GroupModel describes a single group in the list.
type GroupModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	Path        types.String      `tfsdk:"path"`
	MemberCount types.Int64       `tfsdk:"member_count"`
	Roles       []GroupRoleModel  `tfsdk:"roles"`
}

// GroupRoleModel describes a role mapping for a group.
type GroupRoleModel struct {
	RoleName          types.String `tfsdk:"role_name"`
	BundleDisplayName types.String `tfsdk:"bundle_display_name"`
	Provider          types.String `tfsdk:"provider"`
}

// NewGroupsDataSource creates a new groups data source.
func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all groups from the Shoehorn directory (synced from your IdP).",
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Description: "The list of groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The group name.",
							Computed:    true,
						},
						"path": schema.StringAttribute{
							Description: "The group path.",
							Computed:    true,
						},
						"member_count": schema.Int64Attribute{
							Description: "The number of group members.",
							Computed:    true,
						},
						"roles": schema.ListNestedAttribute{
							Description: "Role mappings assigned to the group.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role_name": schema.StringAttribute{
										Description: "The Cerbos role name.",
										Computed:    true,
									},
									"bundle_display_name": schema.StringAttribute{
										Description: "The display name of the role bundle this role belongs to.",
										Computed:    true,
									},
									"provider": schema.StringAttribute{
										Description: "The auth provider this mapping applies to.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	groups, err := d.client.ListGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Groups", fmt.Sprintf("Could not list groups: %s", err))
		return
	}

	var state GroupsDataSourceModel
	for _, g := range groups {
		gm := GroupModel{
			ID:          types.StringValue(g.ID),
			Name:        types.StringValue(g.Name),
			Path:        types.StringValue(g.Path),
			MemberCount: types.Int64Value(int64(g.MemberCount)),
			Roles:       []GroupRoleModel{},
		}
		for _, r := range g.Roles {
			gm.Roles = append(gm.Roles, GroupRoleModel{
				RoleName:          types.StringValue(r.RoleName),
				BundleDisplayName: types.StringValue(r.BundleDisplayName),
				Provider:          types.StringValue(r.Provider),
			})
		}
		state.Groups = append(state.Groups, gm)
	}

	if state.Groups == nil {
		state.Groups = []GroupModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
