package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &FeatureFlagResource{}
	_ resource.ResourceWithImportState = &FeatureFlagResource{}
)

// FeatureFlagResource defines the resource implementation.
type FeatureFlagResource struct {
	client *client.Client
}

// FeatureFlagResourceModel describes the resource data model.
type FeatureFlagResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	DefaultEnabled types.Bool   `tfsdk:"default_enabled"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// NewFeatureFlagResource creates a new feature flag resource.
func NewFeatureFlagResource() resource.Resource {
	return &FeatureFlagResource{}
}

func (r *FeatureFlagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feature_flag"
}

func (r *FeatureFlagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn feature flag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the feature flag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The unique key for the feature flag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the feature flag.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the feature flag.",
				Optional:    true,
			},
			"default_enabled": schema.BoolAttribute{
				Description: "Whether the feature flag is enabled by default.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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

func (r *FeatureFlagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FeatureFlagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FeatureFlagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flag, err := r.client.CreateFeatureFlag(ctx, client.CreateFeatureFlagRequest{
		Key:            plan.Key.ValueString(),
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		DefaultEnabled: plan.DefaultEnabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Feature Flag", fmt.Sprintf("Could not create feature flag: %s", err))
		return
	}

	mapFeatureFlagToState(flag, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FeatureFlagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FeatureFlagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flag, err := r.client.GetFeatureFlag(ctx, state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Feature Flag", fmt.Sprintf("Could not read feature flag %s: %s", state.Key.ValueString(), err))
		return
	}

	mapFeatureFlagToState(flag, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FeatureFlagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FeatureFlagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := plan.DefaultEnabled.ValueBool()
	flag, err := r.client.UpdateFeatureFlag(ctx, plan.Key.ValueString(), client.UpdateFeatureFlagRequest{
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		DefaultEnabled: &enabled,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Feature Flag", fmt.Sprintf("Could not update feature flag %s: %s", plan.Key.ValueString(), err))
		return
	}

	mapFeatureFlagToState(flag, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FeatureFlagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FeatureFlagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFeatureFlag(ctx, state.Key.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Feature Flag", fmt.Sprintf("Could not delete feature flag %s: %s", state.Key.ValueString(), err))
		return
	}
}

func (r *FeatureFlagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func mapFeatureFlagToState(flag *client.FeatureFlag, state *FeatureFlagResourceModel) {
	state.ID = types.StringValue(flag.ID)
	state.Key = types.StringValue(flag.Key)
	state.Name = types.StringValue(flag.Name)
	state.DefaultEnabled = types.BoolValue(flag.DefaultEnabled)

	if flag.Description != "" {
		state.Description = types.StringValue(flag.Description)
	}
	if flag.CreatedAt != "" {
		state.CreatedAt = types.StringValue(flag.CreatedAt)
	}
	if flag.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(flag.UpdatedAt)
	}
}
