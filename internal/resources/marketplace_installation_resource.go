package resources

import (
	"context"
	"encoding/json"
	"fmt"

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
	_ resource.Resource                = &MarketplaceInstallationResource{}
	_ resource.ResourceWithImportState = &MarketplaceInstallationResource{}
)

// MarketplaceInstallationResource defines the resource implementation.
type MarketplaceInstallationResource struct {
	client *client.Client
}

// MarketplaceInstallationResourceModel describes the resource data model.
type MarketplaceInstallationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	ConfigJSON  types.String `tfsdk:"config_json"`
	Kind        types.String `tfsdk:"kind"`
	Version     types.String `tfsdk:"version"`
	SyncStatus  types.String `tfsdk:"sync_status"`
	InstalledBy types.String `tfsdk:"installed_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewMarketplaceInstallationResource creates a new marketplace installation resource.
func NewMarketplaceInstallationResource() resource.Resource {
	return &MarketplaceInstallationResource{}
}

func (r *MarketplaceInstallationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marketplace_installation"
}

func (r *MarketplaceInstallationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn marketplace item installation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the installation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The marketplace item slug. Changing this requires uninstall/reinstall.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the marketplace item is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"config_json": schema.StringAttribute{
				Description: "The addon configuration as a JSON string.",
				Optional:    true,
				Sensitive:   true,
			},
			"kind": schema.StringAttribute{
				Description: "The marketplace item kind.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The installed version.",
				Computed:    true,
			},
			"sync_status": schema.StringAttribute{
				Description: "The synchronization status.",
				Computed:    true,
			},
			"installed_by": schema.StringAttribute{
				Description: "The user who installed the item.",
				Computed:    true,
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

func (r *MarketplaceInstallationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MarketplaceInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating marketplace installation")

	var plan MarketplaceInstallationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := plan.Slug.ValueString()

	// Install the marketplace item
	installation, err := r.client.InstallMarketplaceItem(ctx, slug)
	if err != nil {
		resp.Diagnostics.AddError("Error Installing Marketplace Item", fmt.Sprintf("Could not install marketplace item %q: %s", slug, err))
		return
	}

	// Save partial state so Terraform tracks the resource even if subsequent steps fail
	mapMarketplaceInstallationToState(installation, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If enabled is false, disable the item after install
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && !plan.Enabled.ValueBool() {
		tflog.Debug(ctx, "disabling marketplace item after install", map[string]any{"slug": slug})
		if err := r.client.DisableMarketplaceItem(ctx, slug); err != nil {
			resp.Diagnostics.AddError("Error Disabling Marketplace Item", fmt.Sprintf("Could not disable marketplace item %q: %s", slug, err))
			return
		}
		installation.Enabled = false
	}

	// If config_json is set, update the config
	if !plan.ConfigJSON.IsNull() && !plan.ConfigJSON.IsUnknown() {
		tflog.Debug(ctx, "updating marketplace item config", map[string]any{"slug": slug})
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.ConfigJSON.ValueString()), &configMap); err != nil {
			resp.Diagnostics.AddError("Invalid Config JSON", fmt.Sprintf("Could not parse config_json: %s", err))
			return
		}

		updated, err := r.client.UpdateMarketplaceItemConfig(ctx, slug, configMap)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Marketplace Item Config", fmt.Sprintf("Could not update config for marketplace item %q: %s", slug, err))
			return
		}
		installation = updated
	}

	mapMarketplaceInstallationToState(installation, &plan)
	// Preserve user's config_json to avoid format diffs
	if !plan.ConfigJSON.IsNull() && !plan.ConfigJSON.IsUnknown() {
		// keep plan value
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MarketplaceInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading marketplace installation")

	var state MarketplaceInstallationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := state.Slug.ValueString()
	installation, err := r.client.GetMarketplaceInstallation(ctx, slug)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "marketplace installation not found, removing from state", map[string]any{"slug": slug})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Marketplace Installation", fmt.Sprintf("Could not read marketplace installation %q: %s", slug, err))
		return
	}

	mapMarketplaceInstallationToState(installation, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MarketplaceInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating marketplace installation")

	var plan MarketplaceInstallationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MarketplaceInstallationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := plan.Slug.ValueString()

	// Toggle enabled/disabled if changed
	planEnabled := plan.Enabled.ValueBool()
	stateEnabled := state.Enabled.ValueBool()
	if planEnabled != stateEnabled {
		tflog.Debug(ctx, "toggling marketplace item enabled state", map[string]any{"slug": slug, "enabled": planEnabled})
		if planEnabled {
			if err := r.client.EnableMarketplaceItem(ctx, slug); err != nil {
				resp.Diagnostics.AddError("Error Enabling Marketplace Item", fmt.Sprintf("Could not enable marketplace item %q: %s", slug, err))
				return
			}
		} else {
			if err := r.client.DisableMarketplaceItem(ctx, slug); err != nil {
				resp.Diagnostics.AddError("Error Disabling Marketplace Item", fmt.Sprintf("Could not disable marketplace item %q: %s", slug, err))
				return
			}
		}
	}

	// Update config if changed
	if !plan.ConfigJSON.Equal(state.ConfigJSON) && !plan.ConfigJSON.IsNull() && !plan.ConfigJSON.IsUnknown() {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(plan.ConfigJSON.ValueString()), &configMap); err != nil {
			resp.Diagnostics.AddError("Invalid Config JSON", fmt.Sprintf("Could not parse config_json: %s", err))
			return
		}

		if _, err := r.client.UpdateMarketplaceItemConfig(ctx, slug, configMap); err != nil {
			resp.Diagnostics.AddError("Error Updating Marketplace Item Config", fmt.Sprintf("Could not update config for marketplace item %q: %s", slug, err))
			return
		}
	}

	// Re-read the installation to get the latest state
	installation, err := r.client.GetMarketplaceInstallation(ctx, slug)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Marketplace Installation", fmt.Sprintf("Could not read marketplace installation %q after update: %s", slug, err))
		return
	}

	mapMarketplaceInstallationToState(installation, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MarketplaceInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting marketplace installation")

	var state MarketplaceInstallationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := state.Slug.ValueString()
	if err := r.client.UninstallMarketplaceItem(ctx, slug); err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "marketplace installation already removed, removing from state", map[string]any{"slug": slug})
			return
		}
		resp.Diagnostics.AddError("Error Uninstalling Marketplace Item", fmt.Sprintf("Could not uninstall marketplace item %q: %s", slug, err))
		return
	}
}

func (r *MarketplaceInstallationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

func mapMarketplaceInstallationToState(installation *client.MarketplaceInstallation, state *MarketplaceInstallationResourceModel) {
	state.ID = types.StringValue(installation.ID)
	state.Slug = types.StringValue(installation.Slug)
	state.Enabled = types.BoolValue(installation.Enabled)
	state.Kind = stringValueOrNull(installation.Kind)
	state.Version = stringValueOrNull(installation.Version)
	state.SyncStatus = stringValueOrNull(installation.SyncStatus)
	state.InstalledBy = stringValueOrNull(installation.InstalledBy)
	state.CreatedAt = stringValueOrNull(installation.CreatedAt)
	state.UpdatedAt = stringValueOrNull(installation.UpdatedAt)

	// Don't overwrite config_json from API -- preserve user's original JSON to avoid format diffs.
	// config_json is only written during Create/Update from the plan value.
}
