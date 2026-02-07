package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &IntegrationsDataSource{}

// IntegrationsDataSource defines the data source implementation.
type IntegrationsDataSource struct {
	client *client.Client
}

// IntegrationsDataSourceModel describes the data source data model.
type IntegrationsDataSourceModel struct {
	Integrations []IntegrationModel `tfsdk:"integrations"`
	Total        types.Int64        `tfsdk:"total"`
	Healthy      types.Int64        `tfsdk:"healthy"`
}

// IntegrationModel describes a single system integration.
type IntegrationModel struct {
	Type     types.String `tfsdk:"type"`
	Provider types.String `tfsdk:"provider"`
	Status   types.String `tfsdk:"status"`
	Config   types.String `tfsdk:"config"`
	Metadata types.String `tfsdk:"metadata"`
}

// NewIntegrationsDataSource creates a new integrations data source.
func NewIntegrationsDataSource() datasource.DataSource {
	return &IntegrationsDataSource{}
}

func (d *IntegrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

func (d *IntegrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the status of all system integrations (authentication, GitHub, org data).",
		Attributes: map[string]schema.Attribute{
			"total": schema.Int64Attribute{
				Description: "Total number of integrations.",
				Computed:    true,
			},
			"healthy": schema.Int64Attribute{
				Description: "Number of healthy integrations.",
				Computed:    true,
			},
			"integrations": schema.ListNestedAttribute{
				Description: "The list of system integrations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The integration type (github, authentication, orgdata).",
							Computed:    true,
						},
						"provider": schema.StringAttribute{
							Description: "The integration provider (github, zitadel, etc.).",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The integration status (connected, disconnected, error).",
							Computed:    true,
						},
						"config": schema.StringAttribute{
							Description: "The integration configuration as JSON.",
							Computed:    true,
						},
						"metadata": schema.StringAttribute{
							Description: "The integration metadata as JSON.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *IntegrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IntegrationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	integrations, total, healthy, err := d.client.GetIntegrationsStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Integrations", fmt.Sprintf("Could not get integrations status: %s", err))
		return
	}

	state := IntegrationsDataSourceModel{
		Total:   types.Int64Value(int64(total)),
		Healthy: types.Int64Value(int64(healthy)),
	}

	for _, i := range integrations {
		model := IntegrationModel{
			Type:     types.StringValue(i.Type),
			Provider: types.StringValue(i.Provider),
			Status:   types.StringValue(i.Status),
		}

		if i.Config != nil {
			configJSON, err := json.Marshal(i.Config)
			if err == nil {
				model.Config = types.StringValue(string(configJSON))
			}
		}
		if model.Config.IsNull() {
			model.Config = types.StringValue("{}")
		}

		if i.Metadata != nil {
			metadataJSON, err := json.Marshal(i.Metadata)
			if err == nil {
				model.Metadata = types.StringValue(string(metadataJSON))
			}
		}
		if model.Metadata.IsNull() {
			model.Metadata = types.StringValue("{}")
		}

		state.Integrations = append(state.Integrations, model)
	}

	if state.Integrations == nil {
		state.Integrations = []IntegrationModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
