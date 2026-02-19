package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &UsersDataSource{}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	client *client.Client
}

// UsersDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	Users []UserModel `tfsdk:"users"`
}

// UserModel describes a single user in the list.
type UserModel struct {
	ID          types.String  `tfsdk:"id"`
	Username    types.String  `tfsdk:"username"`
	FirstName   types.String  `tfsdk:"first_name"`
	LastName    types.String  `tfsdk:"last_name"`
	Email       types.String  `tfsdk:"email"`
	Enabled     types.Bool    `tfsdk:"enabled"`
	GitProvider types.String  `tfsdk:"git_provider"`
	Bundles     []BundleModel `tfsdk:"bundles"`
}

// BundleModel describes a bundle summary.
type BundleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Color       types.String `tfsdk:"color"`
}

// NewUsersDataSource creates a new users data source.
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all users from the Shoehorn directory (synced from your IdP).",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description: "The list of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the user.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "The username.",
							Computed:    true,
						},
						"first_name": schema.StringAttribute{
							Description: "The user's first name.",
							Computed:    true,
						},
						"last_name": schema.StringAttribute{
							Description: "The user's last name.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The user's email address.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the user account is enabled.",
							Computed:    true,
						},
						"git_provider": schema.StringAttribute{
							Description: "The git provider associated with the user (e.g. github, gitlab).",
							Computed:    true,
						},
						"bundles": schema.ListNestedAttribute{
							Description: "Role bundles assigned to the user.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "The bundle ID.",
										Computed:    true,
									},
									"name": schema.StringAttribute{
										Description: "The bundle name.",
										Computed:    true,
									},
									"display_name": schema.StringAttribute{
										Description: "The bundle display name.",
										Computed:    true,
									},
									"color": schema.StringAttribute{
										Description: "The bundle color.",
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

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	users, err := d.client.ListDirectoryUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Users", fmt.Sprintf("Could not list users: %s", err))
		return
	}

	var state UsersDataSourceModel
	for _, u := range users {
		um := UserModel{
			ID:          types.StringValue(u.ID),
			Username:    types.StringValue(u.Username),
			FirstName:   types.StringValue(u.FirstName),
			LastName:    types.StringValue(u.LastName),
			Email:       types.StringValue(u.Email),
			Enabled:     types.BoolValue(u.Enabled),
			GitProvider: types.StringValue(u.GitProvider),
			Bundles:     []BundleModel{},
		}
		for _, b := range u.Bundles {
			um.Bundles = append(um.Bundles, BundleModel{
				ID:          types.StringValue(b.ID),
				Name:        types.StringValue(b.Name),
				DisplayName: types.StringValue(b.DisplayName),
				Color:       types.StringValue(b.Color),
			})
		}
		state.Users = append(state.Users, um)
	}

	if state.Users == nil {
		state.Users = []UserModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
