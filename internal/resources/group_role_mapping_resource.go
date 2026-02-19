package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &GroupRoleMappingResource{}
	_ resource.ResourceWithImportState = &GroupRoleMappingResource{}
)

// GroupRoleMappingResource defines the resource implementation.
type GroupRoleMappingResource struct {
	client *client.Client
}

// GroupRoleMappingResourceModel describes the resource data model.
type GroupRoleMappingResourceModel struct {
	ID          types.String `tfsdk:"id"`
	GroupName   types.String `tfsdk:"group_name"`
	RoleName    types.String `tfsdk:"role_name"`
	Provider    types.String `tfsdk:"provider"`
	Description types.String `tfsdk:"description"`
}

// NewGroupRoleMappingResource creates a new group role mapping resource.
func NewGroupRoleMappingResource() resource.Resource {
	return &GroupRoleMappingResource{}
}

func (r *GroupRoleMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_role_mapping"
}

func (r *GroupRoleMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn group role mapping. Maps an IdP group to a Cerbos role so all members of the group inherit that role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the format group_name:role_name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_name": schema.StringAttribute{
				Description: "The IdP group name to map the role to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_name": schema.StringAttribute{
				Description: "The Cerbos role name to assign to the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider": schema.StringAttribute{
				Description: "The auth provider identifier. Defaults to 'default'.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description for the role mapping.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GroupRoleMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupRoleMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupRoleMappingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := plan.GroupName.ValueString()
	roleName := plan.RoleName.ValueString()
	provider := "default"
	if !plan.Provider.IsNull() && !plan.Provider.IsUnknown() && plan.Provider.ValueString() != "" {
		provider = plan.Provider.ValueString()
	}
	description := ""
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		description = plan.Description.ValueString()
	}

	err := r.client.AssignGroupRole(ctx, groupName, client.GroupRoleRequest{
		RoleName:    roleName,
		Provider:    provider,
		Description: description,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Group Role Mapping",
			fmt.Sprintf("Could not assign role %q to group %q: %s", roleName, groupName, err),
		)
		return
	}

	plan.ID = types.StringValue(groupName + ":" + roleName)
	plan.Provider = types.StringValue(provider)
	if description != "" {
		plan.Description = types.StringValue(description)
	} else {
		plan.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupRoleMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupRoleMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := state.GroupName.ValueString()
	roleName := state.RoleName.ValueString()

	roles, err := r.client.GetGroupRoles(ctx, groupName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Group Role Mapping",
			fmt.Sprintf("Could not get roles for group %q: %s", groupName, err),
		)
		return
	}

	// Check if mapping still exists
	found := false
	for _, role := range roles {
		if role.RoleName == roleName {
			found = true
			if role.Provider != "" {
				state.Provider = types.StringValue(role.Provider)
			}
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(groupName + ":" + roleName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupRoleMappingResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// group_name and role_name use RequiresReplace, so Update is never called.
	resp.Diagnostics.AddError("Update Not Supported", "Group role mappings cannot be updated. Changes require replacement.")
}

func (r *GroupRoleMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupRoleMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := state.GroupName.ValueString()
	roleName := state.RoleName.ValueString()

	if err := r.client.RemoveGroupRole(ctx, groupName, roleName); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Group Role Mapping",
			fmt.Sprintf("Could not remove role %q from group %q: %s", roleName, groupName, err),
		)
		return
	}
}

func (r *GroupRoleMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: group_name:role_name
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'group_name:role_name', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_name"), parts[1])...)
}
