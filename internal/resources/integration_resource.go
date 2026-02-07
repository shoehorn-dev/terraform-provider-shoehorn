package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &IntegrationResource{}
	_ resource.ResourceWithImportState = &IntegrationResource{}
)

// IntegrationResource defines the resource implementation.
type IntegrationResource struct {
	client *client.Client
}

// IntegrationResourceModel describes the resource data model.
type IntegrationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	ConfigJSON types.String `tfsdk:"config_json"`
	TeamID    types.String `tfsdk:"team_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// NewIntegrationResource creates a new integration resource.
func NewIntegrationResource() resource.Resource {
	return &IntegrationResource{}
}

func (r *IntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn external integration (GitHub, Slack, AWS, Kubernetes).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the integration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the integration.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The integration type (github, slack, aws, kubernetes).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The integration status (pending, active, inactive, error).",
				Computed:    true,
			},
			"config_json": schema.StringAttribute{
				Description: "The integration configuration as a JSON string. Sensitive fields (tokens, secrets) will be masked on read.",
				Required:    true,
				Sensitive:   true,
			},
			"team_id": schema.StringAttribute{
				Description: "Optional team ID to scope the integration.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *IntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *IntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(plan.ConfigJSON.ValueString()), &configMap); err != nil {
		resp.Diagnostics.AddError("Invalid Config JSON", fmt.Sprintf("Could not parse config_json: %s", err))
		return
	}

	createReq := client.CreateIntegrationRequest{
		Name:   plan.Name.ValueString(),
		Type:   plan.Type.ValueString(),
		Config: configMap,
	}
	if !plan.TeamID.IsNull() && !plan.TeamID.IsUnknown() {
		createReq.TeamID = plan.TeamID.ValueString()
	}

	integration, err := r.client.CreateIntegration(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Integration", fmt.Sprintf("Could not create integration: %s", err))
		return
	}

	mapIntegrationToState(integration, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse integration ID %q: %s", state.ID.ValueString(), err))
		return
	}

	integration, err := r.client.GetIntegration(ctx, id)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapIntegrationToState(integration, &state)
	// Preserve config_json from state since API masks sensitive fields
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse integration ID: %s", err))
		return
	}

	var configMap map[string]interface{}
	if !plan.ConfigJSON.IsNull() && !plan.ConfigJSON.IsUnknown() {
		if err := json.Unmarshal([]byte(plan.ConfigJSON.ValueString()), &configMap); err != nil {
			resp.Diagnostics.AddError("Invalid Config JSON", fmt.Sprintf("Could not parse config_json: %s", err))
			return
		}
	}

	integration, err := r.client.UpdateIntegration(ctx, id, client.UpdateIntegrationRequest{
		Name:   plan.Name.ValueString(),
		Config: configMap,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Integration", fmt.Sprintf("Could not update integration: %s", err))
		return
	}

	mapIntegrationToState(integration, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Could not parse integration ID: %s", err))
		return
	}

	if err := r.client.DeleteIntegration(ctx, id); err != nil {
		resp.Diagnostics.AddError("Error Deleting Integration", fmt.Sprintf("Could not delete integration: %s", err))
		return
	}
}

func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapIntegrationToState(integration *client.Integration, state *IntegrationResourceModel) {
	state.ID = types.StringValue(strconv.Itoa(integration.ID))
	state.Name = types.StringValue(integration.Name)
	state.Type = types.StringValue(integration.Type)

	if integration.Status != "" {
		state.Status = types.StringValue(integration.Status)
	}
	if integration.TeamID != "" {
		state.TeamID = types.StringValue(integration.TeamID)
	}
	if integration.CreatedAt != "" {
		state.CreatedAt = types.StringValue(integration.CreatedAt)
	}
	if integration.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(integration.UpdatedAt)
	}
	// config_json is preserved from state since API masks sensitive fields
}
