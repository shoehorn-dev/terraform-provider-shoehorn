package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &APIKeysDataSource{}

// APIKeysDataSource defines the data source implementation.
type APIKeysDataSource struct {
	client *client.Client
}

// APIKeysDataSourceModel describes the data source data model.
type APIKeysDataSourceModel struct {
	APIKeys []APIKeyModel `tfsdk:"api_keys"`
}

// APIKeyModel describes a single API key in the list.
type APIKeyModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Prefix    types.String `tfsdk:"prefix"`
	Status    types.String `tfsdk:"status"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// NewAPIKeysDataSource creates a new API keys data source.
func NewAPIKeysDataSource() datasource.DataSource {
	return &APIKeysDataSource{}
}

func (d *APIKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *APIKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Shoehorn API keys. Note: secret key values are not returned.",
		Attributes: map[string]schema.Attribute{
			"api_keys": schema.ListNestedAttribute{
				Description: "The list of API keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the API key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The API key name.",
							Computed:    true,
						},
						"prefix": schema.StringAttribute{
							Description: "The API key prefix for identification.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The API key status (active, revoked, expired).",
							Computed:    true,
						},
						"expires_at": schema.StringAttribute{
							Description: "The expiration timestamp.",
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

func (d *APIKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeysDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	keys, err := d.client.ListAPIKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading API Keys", fmt.Sprintf("Could not list API keys: %s", err))
		return
	}

	var state APIKeysDataSourceModel
	for _, k := range keys {
		status := "active"
		if k.RevokedAt != "" {
			status = "revoked"
		}
		state.APIKeys = append(state.APIKeys, APIKeyModel{
			ID:        types.StringValue(k.ID),
			Name:      types.StringValue(k.Name),
			Prefix:    types.StringValue(k.KeyPrefix),
			Status:    types.StringValue(status),
			ExpiresAt: types.StringValue(k.ExpiresAt),
			CreatedAt: types.StringValue(k.CreatedAt),
		})
	}

	if state.APIKeys == nil {
		state.APIKeys = []APIKeyModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
