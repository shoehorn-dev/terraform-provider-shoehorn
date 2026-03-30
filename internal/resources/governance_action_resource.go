package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &GovernanceActionResource{}
	_ resource.ResourceWithImportState = &GovernanceActionResource{}
)

// GovernanceActionResource defines the resource implementation.
type GovernanceActionResource struct {
	client *client.Client
}

// GovernanceActionResourceModel describes the resource data model.
type GovernanceActionResourceModel struct {
	ID             types.String `tfsdk:"id"`
	EntityID       types.String `tfsdk:"entity_id"`
	EntityName     types.String `tfsdk:"entity_name"`
	Title          types.String `tfsdk:"title"`
	Description    types.String `tfsdk:"description"`
	Priority       types.String `tfsdk:"priority"`
	Status         types.String `tfsdk:"status"`
	SourceType     types.String `tfsdk:"source_type"`
	SourceID       types.String `tfsdk:"source_id"`
	AssignedTo     types.String `tfsdk:"assigned_to"`
	SLADays        types.Int64  `tfsdk:"sla_days"`
	DueDate        types.String `tfsdk:"due_date"`
	ResolutionNote types.String `tfsdk:"resolution_note"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// NewGovernanceActionResource creates a new governance action resource.
func NewGovernanceActionResource() resource.Resource {
	return &GovernanceActionResource{}
}

func (r *GovernanceActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_governance_action"
}

func (r *GovernanceActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn governance action item.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the governance action.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity_id": schema.StringAttribute{
				Description: "The entity ID this action is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_name": schema.StringAttribute{
				Description: "The display name of the associated entity.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the governance action.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the governance action.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"priority": schema.StringAttribute{
				Description: "The priority level (critical, high, medium, low).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("critical", "high", "medium", "low"),
				},
			},
			"status": schema.StringAttribute{
				Description: "The action status (open, in_progress, resolved, dismissed, wont_fix). Defaults to open.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("open", "in_progress", "resolved", "dismissed", "wont_fix"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_type": schema.StringAttribute{
				Description: "The source type that created this action (scorecard, security, policy).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("scorecard", "security", "policy"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_id": schema.StringAttribute{
				Description: "The ID of the source that created this action.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assigned_to": schema.StringAttribute{
				Description: "The user or team this action is assigned to.",
				Optional:    true,
			},
			"sla_days": schema.Int64Attribute{
				Description: "The SLA in days for resolving this action.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"due_date": schema.StringAttribute{
				Description: "The computed due date based on SLA.",
				Computed:    true,
			},
			"resolution_note": schema.StringAttribute{
				Description: "A note explaining how the action was resolved.",
				Optional:    true,
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

func (r *GovernanceActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GovernanceActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating governance action")

	var plan GovernanceActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateGovernanceActionRequest{
		EntityID:   plan.EntityID.ValueString(),
		Title:      plan.Title.ValueString(),
		Priority:   plan.Priority.ValueString(),
		SourceType: plan.SourceType.ValueString(),
	}
	if !plan.EntityName.IsNull() && !plan.EntityName.IsUnknown() {
		createReq.EntityName = plan.EntityName.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.Description = plan.Description.ValueString()
	}
	if !plan.SourceID.IsNull() && !plan.SourceID.IsUnknown() {
		createReq.SourceID = plan.SourceID.ValueString()
	}
	if !plan.AssignedTo.IsNull() && !plan.AssignedTo.IsUnknown() {
		v := plan.AssignedTo.ValueString()
		createReq.AssignedTo = &v
	}
	if !plan.SLADays.IsNull() && !plan.SLADays.IsUnknown() {
		v := int(plan.SLADays.ValueInt64())
		createReq.SLADays = &v
	}

	action, err := r.client.CreateGovernanceAction(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Governance Action", fmt.Sprintf("Could not create governance action: %s", err))
		return
	}

	mapGovernanceActionToState(action, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GovernanceActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading governance action")

	var state GovernanceActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	action, err := r.client.GetGovernanceAction(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "governance action not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Governance Action", fmt.Sprintf("Could not read governance action %s: %s", state.ID.ValueString(), err))
		return
	}

	mapGovernanceActionToState(action, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GovernanceActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating governance action")

	var plan GovernanceActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateGovernanceActionRequest{}

	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		v := plan.Status.ValueString()
		updateReq.Status = &v
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		v := plan.Priority.ValueString()
		updateReq.Priority = &v
	}
	if !plan.AssignedTo.IsNull() && !plan.AssignedTo.IsUnknown() {
		v := plan.AssignedTo.ValueString()
		updateReq.AssignedTo = &v
	}
	if !plan.ResolutionNote.IsNull() && !plan.ResolutionNote.IsUnknown() {
		v := plan.ResolutionNote.ValueString()
		updateReq.ResolutionNote = &v
	}

	if err := r.client.UpdateGovernanceAction(ctx, plan.ID.ValueString(), updateReq); err != nil {
		resp.Diagnostics.AddError("Error Updating Governance Action", fmt.Sprintf("Could not update governance action: %s", err))
		return
	}

	// Re-read to get the updated state from the API
	action, err := r.client.GetGovernanceAction(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Governance Action After Update", fmt.Sprintf("Could not read governance action %s: %s", plan.ID.ValueString(), err))
		return
	}

	mapGovernanceActionToState(action, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GovernanceActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting governance action")

	var state GovernanceActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGovernanceAction(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return // already deleted
		}
		resp.Diagnostics.AddError("Error Deleting Governance Action", fmt.Sprintf("Could not delete governance action %s: %s", state.ID.ValueString(), err))
		return
	}
}

func (r *GovernanceActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapGovernanceActionToState(action *client.GovernanceAction, state *GovernanceActionResourceModel) {
	state.ID = types.StringValue(action.ID)
	state.EntityID = types.StringValue(action.EntityID)
	state.Title = types.StringValue(action.Title)
	state.Priority = types.StringValue(action.Priority)
	state.SourceType = types.StringValue(action.SourceType)

	state.EntityName = stringValueOrNull(action.EntityName)
	state.Status = stringValueOrNull(action.Status)
	state.SourceID = stringValueOrNull(action.SourceID)
	state.DueDate = stringValueOrNull(action.DueDate)
	state.CreatedAt = stringValueOrNull(action.CreatedAt)
	state.UpdatedAt = stringValueOrNull(action.UpdatedAt)

	// For user-settable Optional fields, preserve the plan value when API returns empty
	// to avoid "" -> null drift
	if action.Description != "" {
		state.Description = types.StringValue(action.Description)
	} else if state.Description.IsNull() || state.Description.IsUnknown() {
		state.Description = types.StringNull()
	}
	// else: keep existing state value (user set "")

	if action.AssignedTo != "" {
		state.AssignedTo = types.StringValue(action.AssignedTo)
	} else if state.AssignedTo.IsNull() || state.AssignedTo.IsUnknown() {
		state.AssignedTo = types.StringNull()
	}

	if action.ResolutionNote != "" {
		state.ResolutionNote = types.StringValue(action.ResolutionNote)
	} else if state.ResolutionNote.IsNull() || state.ResolutionNote.IsUnknown() {
		state.ResolutionNote = types.StringNull()
	}

	if action.SLADays != nil {
		state.SLADays = types.Int64Value(int64(*action.SLADays))
	} else {
		state.SLADays = types.Int64Null()
	}
}

