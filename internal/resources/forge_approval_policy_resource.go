package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &ForgeApprovalPolicyResource{}
	_ resource.ResourceWithImportState = &ForgeApprovalPolicyResource{}
)

// ForgeApprovalPolicyResource defines the resource implementation.
type ForgeApprovalPolicyResource struct {
	client *client.Client
}

// ForgeApprovalPolicyResourceModel describes the resource data model.
type ForgeApprovalPolicyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	ApprovalChain       types.List   `tfsdk:"steps"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// ApprovalStepModel describes a single step in an approval policy.
type ApprovalStepModel struct {
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Approvers     types.List   `tfsdk:"approvers"`
	RequiredCount types.Int64  `tfsdk:"required_count"`
}

// approvalStepModelAttrTypes returns the attribute types for the step nested object.
func approvalStepModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"description":    types.StringType,
		"approvers":      types.ListType{ElemType: types.StringType},
		"required_count": types.Int64Type,
	}
}

// NewForgeApprovalPolicyResource creates a new forge approval policy resource.
func NewForgeApprovalPolicyResource() resource.Resource {
	return &ForgeApprovalPolicyResource{}
}

func (r *ForgeApprovalPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forge_approval_policy"
}

func (r *ForgeApprovalPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn Forge approval policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the approval policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the approval policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the approval policy.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the approval policy is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"steps": schema.ListNestedAttribute{
				Description: "The approval chain steps in the policy workflow.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the approval step.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of the approval step.",
							Optional:    true,
						},
						"approvers": schema.ListAttribute{
							Description: "The list of approver identifiers for this step.",
							Required:    true,
							ElementType: types.StringType,
						},
						"required_count": schema.Int64Attribute{
							Description: "Number of approvers required. 0 means all must approve.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *ForgeApprovalPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ForgeApprovalPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating forge approval policy")

	var plan ForgeApprovalPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	steps, diags := expandApprovalApprovalChain(ctx, plan.ApprovalChain)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateApprovalPolicyRequest{
		Name:    plan.Name.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		ApprovalChain:   steps,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.Description = plan.Description.ValueString()
	}

	policy, err := r.client.CreateApprovalPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Forge Approval Policy", fmt.Sprintf("Could not create approval policy: %s", err))
		return
	}

	diags = mapApprovalPolicyToState(ctx, policy, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ForgeApprovalPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading forge approval policy")

	var state ForgeApprovalPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetApprovalPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "forge approval policy not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Forge Approval Policy", fmt.Sprintf("Could not read approval policy %s: %s", state.ID.ValueString(), err))
		return
	}

	diags := mapApprovalPolicyToState(ctx, policy, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ForgeApprovalPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating forge approval policy")

	var plan ForgeApprovalPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	steps, diags := expandApprovalApprovalChain(ctx, plan.ApprovalChain)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := plan.Enabled.ValueBool()
	updateReq := client.UpdateApprovalPolicyRequest{
		Name:    plan.Name.ValueString(),
		Enabled: &enabled,
		ApprovalChain:   steps,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateReq.Description = plan.Description.ValueString()
	}

	policy, err := r.client.UpdateApprovalPolicy(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Forge Approval Policy", fmt.Sprintf("Could not update approval policy: %s", err))
		return
	}

	diags = mapApprovalPolicyToState(ctx, policy, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ForgeApprovalPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting forge approval policy")

	var state ForgeApprovalPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApprovalPolicy(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "forge approval policy already deleted, removing from state", map[string]any{"id": state.ID.ValueString()})
			return
		}
		resp.Diagnostics.AddError("Error Deleting Forge Approval Policy", fmt.Sprintf("Could not delete approval policy: %s", err))
		return
	}
}

func (r *ForgeApprovalPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// expandApprovalApprovalChain converts the Terraform list of steps into client ApprovalStep structs.
func expandApprovalApprovalChain(ctx context.Context, stepsList types.List) ([]client.ApprovalStep, diag.Diagnostics) {
	var diags diag.Diagnostics

	var stepModels []ApprovalStepModel
	diags.Append(stepsList.ElementsAs(ctx, &stepModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	steps := make([]client.ApprovalStep, len(stepModels))
	for i, m := range stepModels {
		var approvers []string
		diags.Append(m.Approvers.ElementsAs(ctx, &approvers, false)...)
		if diags.HasError() {
			return nil, diags
		}

		steps[i] = client.ApprovalStep{
			Name:      m.Name.ValueString(),
			Approvers: approvers,
		}
		if !m.Description.IsNull() && !m.Description.IsUnknown() {
			steps[i].Description = m.Description.ValueString()
		}
		if !m.RequiredCount.IsNull() && !m.RequiredCount.IsUnknown() {
			steps[i].RequiredCount = int(m.RequiredCount.ValueInt64())
		}
	}

	return steps, diags
}

// mapApprovalPolicyToState maps a client ForgeApprovalPolicy to the Terraform state model.
func mapApprovalPolicyToState(ctx context.Context, policy *client.ForgeApprovalPolicy, state *ForgeApprovalPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	state.Description = stringValueOrNull(policy.Description)
	state.Enabled = types.BoolValue(policy.Enabled)
	state.CreatedAt = stringValueOrNull(policy.CreatedAt)
	state.UpdatedAt = stringValueOrNull(policy.UpdatedAt)

	stepValues := make([]attr.Value, len(policy.ApprovalChain))
	for i, s := range policy.ApprovalChain {
		approverValues := make([]attr.Value, len(s.Approvers))
		for j, a := range s.Approvers {
			approverValues[j] = types.StringValue(a)
		}

		approversList, d := types.ListValue(types.StringType, approverValues)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		stepObj, d := types.ObjectValue(approvalStepModelAttrTypes(), map[string]attr.Value{
			"name":           types.StringValue(s.Name),
			"description":    stringValueOrNull(s.Description),
			"approvers":      approversList,
			"required_count": types.Int64Value(int64(s.RequiredCount)),
		})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		stepValues[i] = stepObj
	}

	stepsList, d := types.ListValue(types.ObjectType{AttrTypes: approvalStepModelAttrTypes()}, stepValues)
	diags.Append(d...)
	state.ApprovalChain = stepsList

	return diags
}
