package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &ForgeMoldResource{}
	_ resource.ResourceWithImportState = &ForgeMoldResource{}
)

// ForgeMoldResource defines the resource implementation.
type ForgeMoldResource struct {
	client *client.Client
}

// ForgeMoldResourceModel describes the resource data model.
type ForgeMoldResourceModel struct {
	ID           types.String          `tfsdk:"id"`
	Slug         types.String          `tfsdk:"slug"`
	Name         types.String          `tfsdk:"name"`
	Description  types.String          `tfsdk:"description"`
	Version      types.String          `tfsdk:"version"`
	Visibility   types.String          `tfsdk:"visibility"`
	Tags         types.List            `tfsdk:"tags"`
	Icon         types.String          `tfsdk:"icon"`
	Category     types.String          `tfsdk:"category"`
	SchemaJSON   types.String          `tfsdk:"schema_json"`
	DefaultsJSON types.String          `tfsdk:"defaults_json"`
	Actions      []ForgeMoldActionModel `tfsdk:"actions"`
	Published    types.Bool            `tfsdk:"published"`
	CreatedAt    types.String          `tfsdk:"created_at"`
	UpdatedAt    types.String          `tfsdk:"updated_at"`
}

// ForgeMoldActionModel describes a single action in the resource data model.
type ForgeMoldActionModel struct {
	Action      types.String `tfsdk:"action"`
	Label       types.String `tfsdk:"label"`
	Description types.String `tfsdk:"description"`
	Primary     types.Bool   `tfsdk:"primary"`
}

// NewForgeMoldResource creates a new forge mold resource.
func NewForgeMoldResource() resource.Resource {
	return &ForgeMoldResource{}
}

func (r *ForgeMoldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forge_mold"
}

func (r *ForgeMoldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn Forge mold template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the forge mold.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The immutable slug identifier of the forge mold.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the forge mold.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the forge mold.",
				Optional:    true,
			},
			"version": schema.StringAttribute{
				Description: "The version of the forge mold (e.g. 1.0.0). Changing forces replacement.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"visibility": schema.StringAttribute{
				Description: "The visibility of the forge mold. Valid values: public, tenant, private.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("public", "tenant", "private"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				Description: "Tags for categorizing the forge mold.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"icon": schema.StringAttribute{
				Description: "The icon identifier for the forge mold.",
				Optional:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category of the forge mold.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_json": schema.StringAttribute{
				Description: "The mold input schema as a JSON string.",
				Optional:    true,
			},
			"defaults_json": schema.StringAttribute{
				Description: "The default values as a JSON string.",
				Optional:    true,
			},
			"actions": schema.ListNestedAttribute{
				Description: "The actions available on the forge mold.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Description: "The action identifier.",
							Required:    true,
						},
						"label": schema.StringAttribute{
							Description: "The display label for the action.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of the action.",
							Optional:    true,
						},
						"primary": schema.BoolAttribute{
							Description: "Whether this is the primary action.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
			"published": schema.BoolAttribute{
				Description: "Whether the forge mold is published.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
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

func (r *ForgeMoldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ForgeMoldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating forge mold")

	var plan ForgeMoldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateForgeMoldRequest{
		Slug:       plan.Slug.ValueString(),
		Name:       plan.Name.ValueString(),
		Version:    plan.Version.ValueString(),
		Visibility: plan.Visibility.ValueString(),
		Category:   plan.Category.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.Description = plan.Description.ValueString()
	}

	if !plan.Icon.IsNull() && !plan.Icon.IsUnknown() {
		createReq.Icon = plan.Icon.ValueString()
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Tags = tags
	}

	if !plan.SchemaJSON.IsNull() && !plan.SchemaJSON.IsUnknown() {
		var schemaMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.SchemaJSON.ValueString()), &schemaMap); err != nil {
			resp.Diagnostics.AddError("Invalid Schema JSON", fmt.Sprintf("Could not parse schema_json: %s", err))
			return
		}
		createReq.Schema = schemaMap
	}

	if !plan.DefaultsJSON.IsNull() && !plan.DefaultsJSON.IsUnknown() {
		var defaultsMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.DefaultsJSON.ValueString()), &defaultsMap); err != nil {
			resp.Diagnostics.AddError("Invalid Defaults JSON", fmt.Sprintf("Could not parse defaults_json: %s", err))
			return
		}
		createReq.Defaults = defaultsMap
	}

	createReq.Actions = mapActionsFromModel(plan.Actions)

	mold, err := r.client.CreateForgeMold(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Forge Mold", fmt.Sprintf("Could not create forge mold: %s", err))
		return
	}

	// Save partial state so Terraform tracks the resource even if publish fails
	mapForgeMoldToState(mold, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Publish the mold after creation
	tflog.Debug(ctx, "publishing forge mold", map[string]any{"slug": mold.Slug, "version": mold.Version})
	mold, err = r.client.PublishForgeMold(ctx, mold.Slug, mold.Version)
	if err != nil {
		resp.Diagnostics.AddError("Error Publishing Forge Mold", fmt.Sprintf("Forge mold was created but could not be published: %s", err))
		return
	}

	mapForgeMoldToState(mold, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ForgeMoldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading forge mold")

	var state ForgeMoldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mold, err := r.client.GetForgeMold(ctx, state.Slug.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "forge mold not found, removing from state", map[string]any{"slug": state.Slug.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Forge Mold", fmt.Sprintf("Could not read forge mold %s: %s", state.Slug.ValueString(), err))
		return
	}

	mapForgeMoldToState(mold, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ForgeMoldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating forge mold")

	var plan ForgeMoldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateForgeMoldRequest{
		Version: plan.Version.ValueString(),
		Name:    plan.Name.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateReq.Description = plan.Description.ValueString()
	}

	if !plan.Icon.IsNull() && !plan.Icon.IsUnknown() {
		updateReq.Icon = plan.Icon.ValueString()
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Tags = tags
	}

	if !plan.SchemaJSON.IsNull() && !plan.SchemaJSON.IsUnknown() {
		var schemaMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.SchemaJSON.ValueString()), &schemaMap); err != nil {
			resp.Diagnostics.AddError("Invalid Schema JSON", fmt.Sprintf("Could not parse schema_json: %s", err))
			return
		}
		updateReq.Schema = schemaMap
	}

	if !plan.DefaultsJSON.IsNull() && !plan.DefaultsJSON.IsUnknown() {
		var defaultsMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.DefaultsJSON.ValueString()), &defaultsMap); err != nil {
			resp.Diagnostics.AddError("Invalid Defaults JSON", fmt.Sprintf("Could not parse defaults_json: %s", err))
			return
		}
		updateReq.Defaults = defaultsMap
	}

	mold, err := r.client.UpdateForgeMold(ctx, plan.Slug.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Forge Mold", fmt.Sprintf("Could not update forge mold %s: %s", plan.Slug.ValueString(), err))
		return
	}

	mapForgeMoldToState(mold, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ForgeMoldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting forge mold")

	var state ForgeMoldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteForgeMold(ctx, state.Slug.ValueString(), state.Version.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return // already deleted
		}
		resp.Diagnostics.AddError("Error Deleting Forge Mold", fmt.Sprintf("Could not delete forge mold %s: %s", state.Slug.ValueString(), err))
		return
	}
}

func (r *ForgeMoldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

// mapForgeMoldToState maps a client ForgeMold to the Terraform resource model.
// Fields that the API may transform (schema_json, defaults_json, visibility, actions)
// are preserved from the existing state to avoid perpetual diffs.
func mapForgeMoldToState(mold *client.ForgeMold, state *ForgeMoldResourceModel) {
	state.ID = types.StringValue(mold.ID)
	state.Slug = types.StringValue(mold.Slug)
	state.Name = types.StringValue(mold.Name)
	state.Version = types.StringValue(mold.Version)
	state.Category = types.StringValue(mold.Category)
	state.Published = types.BoolValue(mold.Published)

	// Preserve visibility from plan/state -- API may normalize (e.g., tenant -> public)
	if state.Visibility.IsNull() || state.Visibility.IsUnknown() {
		state.Visibility = types.StringValue(mold.Visibility)
	}

	state.Description = stringValueOrNull(mold.Description)
	state.Icon = stringValueOrNull(mold.Icon)
	state.CreatedAt = stringValueOrNull(mold.CreatedAt)
	state.UpdatedAt = stringValueOrNull(mold.UpdatedAt)

	// Map tags
	if len(mold.Tags) > 0 {
		tagValues := make([]types.String, len(mold.Tags))
		for i, t := range mold.Tags {
			tagValues[i] = types.StringValue(t)
		}
		listVal, _ := types.ListValueFrom(context.Background(), types.StringType, tagValues)
		state.Tags = listVal
	} else if state.Tags.IsNull() {
		state.Tags = types.ListNull(types.StringType)
	}

	// Preserve schema_json and defaults_json from plan/state -- API may transform types/ordering
	// Only set from API if not already present in state (e.g., during import)
	if state.SchemaJSON.IsNull() || state.SchemaJSON.IsUnknown() {
		if len(mold.Schema) > 0 {
			schemaJSON, err := json.Marshal(mold.Schema)
			if err == nil {
				state.SchemaJSON = types.StringValue(string(schemaJSON))
			}
		} else {
			state.SchemaJSON = types.StringNull()
		}
	}

	if state.DefaultsJSON.IsNull() || state.DefaultsJSON.IsUnknown() {
		if len(mold.Defaults) > 0 {
			defaultsJSON, err := json.Marshal(mold.Defaults)
			if err == nil {
				state.DefaultsJSON = types.StringValue(string(defaultsJSON))
			}
		} else {
			state.DefaultsJSON = types.StringNull()
		}
	}

	// Preserve actions from plan/state -- API may transform labels/descriptions
	// Only set from API if not already present in state (e.g., during import)
	if state.Actions == nil {
		if len(mold.Actions) > 0 {
			actions := make([]ForgeMoldActionModel, len(mold.Actions))
			for i, a := range mold.Actions {
				actions[i] = ForgeMoldActionModel{
					Action:      types.StringValue(a.Action),
					Label:       types.StringValue(a.Label),
					Description: stringValueOrNull(a.Description),
					Primary:     types.BoolValue(a.Primary),
				}
			}
			state.Actions = actions
		} else {
			state.Actions = []ForgeMoldActionModel{}
		}
	}
}

// mapActionsFromModel converts the Terraform model actions to client request actions.
func mapActionsFromModel(actions []ForgeMoldActionModel) []client.ForgeMoldAction {
	result := make([]client.ForgeMoldAction, len(actions))
	for i, a := range actions {
		result[i] = client.ForgeMoldAction{
			Action: a.Action.ValueString(),
			Label:  a.Label.ValueString(),
		}
		if !a.Description.IsNull() && !a.Description.IsUnknown() {
			result[i].Description = a.Description.ValueString()
		}
		if !a.Primary.IsNull() && !a.Primary.IsUnknown() {
			result[i].Primary = a.Primary.ValueBool()
		}
	}
	return result
}
