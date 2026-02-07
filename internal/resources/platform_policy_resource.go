package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &PlatformPolicyResource{}
	_ resource.ResourceWithImportState = &PlatformPolicyResource{}
)

// PlatformPolicyResource defines the resource implementation.
// Platform policies are configuration-only: they are pre-seeded by Shoehorn and
// cannot be created or destroyed. Terraform manages their enabled/enforcement state.
type PlatformPolicyResource struct {
	client *client.Client
}

// PlatformPolicyResourceModel describes the resource data model.
type PlatformPolicyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Category    types.String `tfsdk:"category"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Enforcement types.String `tfsdk:"enforcement"`
	System      types.Bool   `tfsdk:"system"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewPlatformPolicyResource creates a new platform policy resource.
func NewPlatformPolicyResource() resource.Resource {
	return &PlatformPolicyResource{}
}

func (r *PlatformPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_platform_policy"
}

func (r *PlatformPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn platform policy configuration. Policies are pre-seeded and cannot be created or destroyed. Use this resource to configure enabled state and enforcement level.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The unique key of the policy (used to identify pre-seeded policies).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the policy.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the policy.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "The policy category (security, governance, compliance, performance).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the policy is enabled. System policies cannot be disabled.",
				Required:    true,
			},
			"enforcement": schema.StringAttribute{
				Description: "The enforcement level (warn, block, audit).",
				Required:    true,
			},
			"system": schema.BoolAttribute{
				Description: "Whether this is a system policy (cannot be disabled).",
				Computed:    true,
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

func (r *PlatformPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PlatformPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PlatformPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Policies are pre-seeded. "Create" means find the policy by key and configure it.
	policy, err := r.client.GetPolicy(ctx, plan.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Policy Not Found", fmt.Sprintf("Policy %q not found. Platform policies are pre-seeded and cannot be created: %s", plan.Key.ValueString(), err))
		return
	}

	// Update the policy configuration
	enabled := plan.Enabled.ValueBool()
	updated, err := r.client.UpdatePolicy(ctx, policy.ID, client.UpdatePolicyRequest{
		Enabled:     &enabled,
		Enforcement: plan.Enforcement.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Configuring Policy", fmt.Sprintf("Could not configure policy %s: %s", plan.Key.ValueString(), err))
		return
	}

	mapPolicyToState(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PlatformPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PlatformPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetPolicy(ctx, state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Policy", fmt.Sprintf("Could not read policy %s: %s", state.Key.ValueString(), err))
		return
	}

	mapPolicyToState(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PlatformPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PlatformPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := plan.Enabled.ValueBool()
	updated, err := r.client.UpdatePolicy(ctx, plan.ID.ValueString(), client.UpdatePolicyRequest{
		Enabled:     &enabled,
		Enforcement: plan.Enforcement.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Policy", fmt.Sprintf("Could not update policy %s: %s", plan.Key.ValueString(), err))
		return
	}

	mapPolicyToState(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PlatformPolicyResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Policies cannot be deleted - they are pre-seeded.
	// Removing from Terraform state only. The policy remains in Shoehorn.
}

func (r *PlatformPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func mapPolicyToState(policy *client.PlatformPolicy, state *PlatformPolicyResourceModel) {
	state.ID = types.StringValue(policy.ID)
	state.Key = types.StringValue(policy.Key)
	state.Name = types.StringValue(policy.Name)
	state.Enabled = types.BoolValue(policy.Enabled)
	state.System = types.BoolValue(policy.System)

	if policy.Description != "" {
		state.Description = types.StringValue(policy.Description)
	}
	if policy.Category != "" {
		state.Category = types.StringValue(policy.Category)
	}
	if policy.Enforcement != "" {
		state.Enforcement = types.StringValue(policy.Enforcement)
	}
	if policy.CreatedAt != "" {
		state.CreatedAt = types.StringValue(policy.CreatedAt)
	}
	if policy.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(policy.UpdatedAt)
	}
}
