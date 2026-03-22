package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &TeamResource{}
	_ resource.ResourceWithImportState = &TeamResource{}
)

// TeamResource defines the resource implementation.
type TeamResource struct {
	client *client.Client
}

// TeamResourceModel describes the resource data model.
type TeamResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	Metadata    types.String `tfsdk:"metadata"`
	Members     types.String `tfsdk:"members"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	MemberCount types.Int64  `tfsdk:"member_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewTeamResource creates a new team resource.
func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

func (r *TeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *TeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the team.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the team.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the team.",
				Optional:    true,
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The unique slug for the team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the team.",
				Optional:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "JSON-encoded metadata for the team.",
				Optional:    true,
			},
			"members": schema.StringAttribute{
				Description: "JSON-encoded array of team members. Each member has user_id (required) and optional role (e.g., manager, admin, member).",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the team is active.",
				Computed:    true,
			},
			"member_count": schema.Int64Attribute{
				Description: "The number of members in the team.",
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

func (r *TeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating team")

	var plan TeamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateTeamRequest{
		Name:        plan.Name.ValueString(),
		Slug:        plan.Slug.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Metadata.ValueString()), &metadata); err != nil {
			resp.Diagnostics.AddError("Invalid Metadata", fmt.Sprintf("Failed to parse metadata JSON: %s", err))
			return
		}
		createReq.Metadata = metadata
	}

	team, err := r.client.CreateTeam(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Team", fmt.Sprintf("Could not create team: %s", err))
		return
	}

	// Save partial state immediately so the team is tracked even if member addition fails
	mapTeamToState(team, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add members if specified
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var members []client.AddMemberRequest
		if err := json.Unmarshal([]byte(plan.Members.ValueString()), &members); err != nil {
			resp.Diagnostics.AddError("Invalid Members JSON", fmt.Sprintf("Team was created but members could not be parsed: %s. Fix the members JSON configuration and run terraform apply again.", err))
			return
		}
		if len(members) > 0 {
			updateReq := client.UpdateTeamRequest{
				Name:       team.Name,
				AddMembers: members,
			}
			team, err = r.client.UpdateTeam(ctx, team.ID, updateReq)
			if err != nil {
				resp.Diagnostics.AddError("Error Adding Members", fmt.Sprintf("Team was created but members could not be added: %s. Run terraform apply again to retry.", err))
				return
			}
			// Update state with members; preserve planned members if API omits them
			plannedMembers := plan.Members
			mapTeamToState(team, &plan)
			if plan.Members.IsNull() && !plannedMembers.IsNull() {
				plan.Members = plannedMembers
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		}
	}
}

func (r *TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading team")

	var state TeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prevMembers := state.Members

	team, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "team not found, removing from state", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Team", fmt.Sprintf("Could not read team %s: %s", state.ID.ValueString(), err))
		return
	}

	mapTeamToState(team, &state)

	// Preserve original members order if semantically equivalent
	if !prevMembers.IsNull() && !state.Members.IsNull() {
		if membersEquivalent(prevMembers.ValueString(), state.Members.ValueString()) {
			state.Members = prevMembers
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating team")

	var plan TeamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateTeamRequest{
		Name:        plan.Name.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Metadata.ValueString()), &metadata); err != nil {
			resp.Diagnostics.AddError("Invalid Metadata", fmt.Sprintf("Failed to parse metadata JSON: %s", err))
			return
		}
		updateReq.Metadata = metadata
	}

	// Compute member diff
	addMembers, removeMembers := computeMemberDiff(state.Members, plan.Members)
	updateReq.AddMembers = addMembers
	updateReq.RemoveMembers = removeMembers

	team, err := r.client.UpdateTeam(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Team", fmt.Sprintf("Could not update team %s: %s", state.ID.ValueString(), err))
		return
	}

	plannedMembers := plan.Members
	mapTeamToState(team, &plan)
	if plan.Members.IsNull() && !plannedMembers.IsNull() {
		plan.Members = plannedMembers
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting team")

	var state TeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTeam(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Team", fmt.Sprintf("Could not delete team %s: %s", state.ID.ValueString(), err))
		return
	}
}

func (r *TeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapTeamToState(team *client.Team, state *TeamResourceModel) {
	state.ID = types.StringValue(team.ID)
	state.Name = types.StringValue(team.Name)
	state.Slug = types.StringValue(team.Slug)
	state.IsActive = types.BoolValue(team.IsActive)
	state.MemberCount = types.Int64Value(int64(team.MemberCount))

	state.DisplayName = stringValueOrNull(team.DisplayName)
	state.Description = stringValueOrNull(team.Description)
	state.CreatedAt = stringValueOrNull(team.CreatedAt)
	state.UpdatedAt = stringValueOrNull(team.UpdatedAt)

	if len(team.Metadata) > 0 {
		metadataJSON, err := json.Marshal(team.Metadata)
		if err == nil {
			state.Metadata = types.StringValue(string(metadataJSON))
		}
	} else {
		state.Metadata = types.StringNull()
	}

	// Map members from API response to terraform state
	if len(team.Members) > 0 {
		type tfMember struct {
			UserID string `json:"user_id"`
			Role   string `json:"role,omitempty"`
		}
		tfMembers := make([]tfMember, len(team.Members))
		for i, m := range team.Members {
			tfMembers[i] = tfMember{
				UserID: m.UserID,
				Role:   m.Role,
			}
		}
		membersJSON, err := json.Marshal(tfMembers)
		if err == nil {
			state.Members = types.StringValue(string(membersJSON))
		}
	} else {
		state.Members = types.StringNull()
	}
}

// tfMemberEntry represents a member entry in terraform config.
type tfMemberEntry struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

// computeMemberDiff computes the add/remove member operations needed to go from
// current state to desired plan.
func computeMemberDiff(stateMembersAttr, planMembersAttr types.String) ([]client.AddMemberRequest, []string) {
	var stateMembers, planMembers []tfMemberEntry

	if !stateMembersAttr.IsNull() && !stateMembersAttr.IsUnknown() {
		json.Unmarshal([]byte(stateMembersAttr.ValueString()), &stateMembers)
	}
	if !planMembersAttr.IsNull() && !planMembersAttr.IsUnknown() {
		json.Unmarshal([]byte(planMembersAttr.ValueString()), &planMembers)
	}

	// Build maps for comparison
	stateMap := make(map[string]string, len(stateMembers))
	for _, m := range stateMembers {
		stateMap[m.UserID] = m.Role
	}
	planMap := make(map[string]string, len(planMembers))
	for _, m := range planMembers {
		planMap[m.UserID] = m.Role
	}

	// Members to add: in plan but not in state, or role changed
	var addMembers []client.AddMemberRequest
	for _, m := range planMembers {
		existingRole, exists := stateMap[m.UserID]
		if !exists || existingRole != m.Role {
			addMembers = append(addMembers, client.AddMemberRequest{
				UserID: m.UserID,
				Role:   m.Role,
			})
		}
	}

	// Members to remove: in state but not in plan
	var removeMembers []string
	for _, m := range stateMembers {
		if _, exists := planMap[m.UserID]; !exists {
			removeMembers = append(removeMembers, m.UserID)
		}
	}

	return addMembers, removeMembers
}

// membersEquivalent checks if two members JSON strings contain the same set of members
// regardless of ordering.
func membersEquivalent(a, b string) bool {
	var membersA, membersB []tfMemberEntry
	if err := json.Unmarshal([]byte(a), &membersA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &membersB); err != nil {
		return false
	}
	if len(membersA) != len(membersB) {
		return false
	}
	setA := make(map[string]bool, len(membersA))
	for _, m := range membersA {
		setA[m.UserID+"|"+m.Role] = true
	}
	for _, m := range membersB {
		if !setA[m.UserID+"|"+m.Role] {
			return false
		}
	}
	return true
}
