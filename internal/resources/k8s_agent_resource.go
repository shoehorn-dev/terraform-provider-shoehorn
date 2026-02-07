package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ resource.Resource = &K8sAgentResource{}

// K8sAgentResource defines the resource implementation.
type K8sAgentResource struct {
	client *client.Client
}

// K8sAgentResourceModel describes the resource data model.
type K8sAgentResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ClusterID   types.String `tfsdk:"cluster_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ExpiresIn   types.Int64  `tfsdk:"expires_in_days"`
	Token       types.String `tfsdk:"token"`
	TokenPrefix types.String `tfsdk:"token_prefix"`
	Status      types.String `tfsdk:"status"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// NewK8sAgentResource creates a new K8s agent resource.
func NewK8sAgentResource() resource.Resource {
	return &K8sAgentResource{}
}

func (r *K8sAgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_agent"
}

func (r *K8sAgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers a Shoehorn K8s agent. The agent token is only available on creation and stored in state. Deleting this resource revokes and removes the agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The cluster ID (used as the unique identifier).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The unique cluster identifier.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the cluster.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_in_days": schema.Int64Attribute{
				Description: "Number of days until the agent token expires. Null means never expires.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The agent token. Only available on creation.",
				Computed:    true,
				Sensitive:   true,
			},
			"token_prefix": schema.StringAttribute{
				Description: "The token prefix for identification.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The agent status (active, inactive, revoked, expired).",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "The expiration timestamp of the agent token.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *K8sAgentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *K8sAgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan K8sAgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registerReq := client.RegisterK8sAgentRequest{
		ClusterID:   plan.ClusterID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.ExpiresIn.IsNull() && !plan.ExpiresIn.IsUnknown() {
		days := int(plan.ExpiresIn.ValueInt64())
		registerReq.ExpiresIn = &days
	}

	regResp, err := r.client.RegisterK8sAgent(ctx, registerReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Registering K8s Agent", fmt.Sprintf("Could not register K8s agent: %s", err))
		return
	}

	plan.ID = types.StringValue(regResp.ClusterID)
	plan.Token = types.StringValue(regResp.Token)
	plan.TokenPrefix = types.StringValue(regResp.TokenPrefix)
	plan.Status = types.StringValue("active")

	if regResp.ExpiresAt != "" {
		plan.ExpiresAt = types.StringValue(regResp.ExpiresAt)
	}
	if regResp.CreatedAt != "" {
		plan.CreatedAt = types.StringValue(regResp.CreatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *K8sAgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state K8sAgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := r.client.GetK8sAgent(ctx, state.ClusterID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// If agent is revoked, remove from state
	if agent.Status == "revoked" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(agent.ClusterID)
	state.Name = types.StringValue(agent.Name)
	state.Status = types.StringValue(agent.Status)

	if agent.TokenPrefix != "" {
		state.TokenPrefix = types.StringValue(agent.TokenPrefix)
	}
	if agent.ExpiresAt != "" {
		state.ExpiresAt = types.StringValue(agent.ExpiresAt)
	}
	if agent.CreatedAt != "" {
		state.CreatedAt = types.StringValue(agent.CreatedAt)
	}
	// token is preserved from state - not returned by API after creation

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *K8sAgentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// K8s agents are immutable - all mutable fields have RequiresReplace
	resp.Diagnostics.AddError("Update Not Supported", "K8s agent registrations cannot be updated. Changes require creating a new agent.")
}

func (r *K8sAgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state K8sAgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterID.ValueString()

	// Revoke first, then delete
	if err := r.client.RevokeK8sAgent(ctx, clusterID); err != nil {
		// Revocation may fail if already revoked; try delete anyway
	}

	if err := r.client.DeleteK8sAgent(ctx, clusterID); err != nil {
		resp.Diagnostics.AddError("Error Deleting K8s Agent", fmt.Sprintf("Could not delete K8s agent %s: %s", clusterID, err))
		return
	}
}
