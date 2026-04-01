package resources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &TenantSettingsResource{}
	_ resource.ResourceWithImportState = &TenantSettingsResource{}
)

// TenantSettingsResource defines the resource implementation.
type TenantSettingsResource struct {
	client *client.Client
}

// TenantSettingsResourceModel describes the resource data model.
type TenantSettingsResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	PrimaryColor        types.String `tfsdk:"primary_color"`
	SecondaryColor      types.String `tfsdk:"secondary_color"`
	AccentColor         types.String `tfsdk:"accent_color"`
	LogoURL             types.String `tfsdk:"logo_url"`
	FaviconURL          types.String `tfsdk:"favicon_url"`
	DefaultTheme        types.String `tfsdk:"default_theme"`
	PlatformName        types.String `tfsdk:"platform_name"`
	PlatformDescription types.String `tfsdk:"platform_description"`
	CompanyName         types.String `tfsdk:"company_name"`
	HiddenPages         types.List   `tfsdk:"hidden_pages"`
	Announcement        types.Object `tfsdk:"announcement"`
	Forge               types.Object `tfsdk:"forge"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

// AnnouncementSettingsModel represents announcement configuration.
type AnnouncementSettingsModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	Message   types.String `tfsdk:"message"`
	Type      types.String `tfsdk:"type"`
	Pinned    types.Bool   `tfsdk:"pinned"`
	LinkURL   types.String `tfsdk:"link_url"`
	LinkText  types.String `tfsdk:"link_text"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// ForgeSettingsModel represents forge configuration.
type ForgeSettingsModel struct {
	AllowedOrgs types.List   `tfsdk:"allowed_orgs"`
	DefaultOrg  types.String `tfsdk:"default_org"`
}

// NewTenantSettingsResource creates a new tenant settings resource.
func NewTenantSettingsResource() resource.Resource {
	return &TenantSettingsResource{}
}

func (r *TenantSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_settings"
}

func (r *TenantSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Shoehorn tenant appearance settings. This is a singleton resource per tenant - create performs an upsert and delete only removes from Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the settings.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_color": schema.StringAttribute{
				Description: "Primary brand color (hex, e.g., #3b82f6). Used for active states and primary buttons.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`),
						"must be a valid hex color code (e.g., #3b82f6)",
					),
				},
			},
			"secondary_color": schema.StringAttribute{
				Description: "Secondary brand color (hex, e.g., #64748b). Used for hover states and secondary UI elements.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`),
						"must be a valid hex color code (e.g., #64748b)",
					),
				},
			},
			"accent_color": schema.StringAttribute{
				Description: "Accent color (hex, e.g., #8b5cf6). Used for highlights, badges, and emphasis.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`),
						"must be a valid hex color code (e.g., #8b5cf6)",
					),
				},
			},
			"logo_url": schema.StringAttribute{
				Description: "URL to the company logo (must be http:// or https://).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https?://`),
						"must be an HTTP or HTTPS URL",
					),
				},
			},
			"favicon_url": schema.StringAttribute{
				Description: "URL to the favicon (must be http:// or https://).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https?://`),
						"must be an HTTP or HTTPS URL",
					),
				},
			},
			"default_theme": schema.StringAttribute{
				Description: "Default theme for users. Valid values: light, dark, system.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("light", "dark", "system"),
				},
			},
			"platform_name": schema.StringAttribute{
				Description: "Name of the platform displayed in the UI.",
				Optional:    true,
			},
			"platform_description": schema.StringAttribute{
				Description: "Description of the platform.",
				Optional:    true,
			},
			"company_name": schema.StringAttribute{
				Description: "Company name.",
				Optional:    true,
			},
			"hidden_pages": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of page slugs to hide from non-admin users (e.g., forge, insights).",
				Optional:    true,
			},
			"announcement": schema.SingleNestedAttribute{
				Description: "Announcement bar configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether announcement bar is enabled.",
						Optional:    true,
						Computed:    true,
					},
					"message": schema.StringAttribute{
						Description: "Announcement message text.",
						Optional:    true,
						Computed:    true,
					},
					"type": schema.StringAttribute{
						Description: "Announcement type. Valid values: info, warning, error, success.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("info", "warning", "error", "success"),
						},
					},
					"pinned": schema.BoolAttribute{
						Description: "If true, users cannot dismiss the announcement.",
						Optional:    true,
						Computed:    true,
					},
					"link_url": schema.StringAttribute{
						Description: "Optional call-to-action link URL (must be http:// or https://).",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								regexp.MustCompile(`^https?://`),
								"must be an HTTP or HTTPS URL",
							),
						},
					},
					"link_text": schema.StringAttribute{
						Description: "Optional call-to-action link text.",
						Optional:    true,
						Computed:    true,
					},
					"updated_at": schema.StringAttribute{
						Description: "Announcement last update timestamp (used for dismiss tracking).",
						Computed:    true,
					},
				},
			},
			"forge": schema.SingleNestedAttribute{
				Description: "Forge configuration for scaffolding and templates.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"allowed_orgs": schema.ListAttribute{
						ElementType: types.StringType,
						Description: "List of GitHub organizations allowed for Forge templates.",
						Optional:    true,
						Computed:    true,
					},
					"default_org": schema.StringAttribute{
						Description: "Default GitHub organization for Forge templates.",
						Optional:    true,
						Computed:    true,
					},
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

func (r *TenantSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating tenant settings")

	var plan TenantSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appearance := buildAppearanceFromModel(ctx, &plan, &resp.Diagnostics)
	announcement := buildAnnouncementFromModel(ctx, &plan, &resp.Diagnostics)
	forge := buildForgeFromModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.UpdateSettings(ctx, client.UpdateSettingsRequest{
		Appearance:   appearance,
		Announcement: announcement,
		Forge:        forge,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Tenant Settings", fmt.Sprintf("Could not create/update settings: %s", err))
		return
	}

	mapSettingsToState(ctx, settings, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading tenant settings")

	var state TenantSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Tenant Settings", fmt.Sprintf("Could not read settings: %s", err))
		return
	}

	mapSettingsToState(ctx, settings, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TenantSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating tenant settings")

	var plan TenantSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appearance := buildAppearanceFromModel(ctx, &plan, &resp.Diagnostics)
	announcement := buildAnnouncementFromModel(ctx, &plan, &resp.Diagnostics)
	forge := buildForgeFromModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.UpdateSettings(ctx, client.UpdateSettingsRequest{
		Appearance:   appearance,
		Announcement: announcement,
		Forge:        forge,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Tenant Settings", fmt.Sprintf("Could not update settings: %s", err))
		return
	}

	mapSettingsToState(ctx, settings, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TenantSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting tenant settings (state only)")
	// Tenant settings are a singleton and can't be truly deleted.
	// Removing from state only - the settings will remain in the API.
}

func (r *TenantSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildAppearanceFromModel(ctx context.Context, model *TenantSettingsResourceModel, diags *diag.Diagnostics) client.AppearanceSettings {
	appearance := client.AppearanceSettings{}
	if !model.PrimaryColor.IsNull() && !model.PrimaryColor.IsUnknown() {
		appearance.PrimaryColor = model.PrimaryColor.ValueString()
	}
	if !model.SecondaryColor.IsNull() && !model.SecondaryColor.IsUnknown() {
		appearance.SecondaryColor = model.SecondaryColor.ValueString()
	}
	if !model.AccentColor.IsNull() && !model.AccentColor.IsUnknown() {
		appearance.AccentColor = model.AccentColor.ValueString()
	}
	if !model.LogoURL.IsNull() && !model.LogoURL.IsUnknown() {
		appearance.LogoURL = model.LogoURL.ValueString()
	}
	if !model.FaviconURL.IsNull() && !model.FaviconURL.IsUnknown() {
		appearance.FaviconURL = model.FaviconURL.ValueString()
	}
	if !model.DefaultTheme.IsNull() && !model.DefaultTheme.IsUnknown() {
		appearance.DefaultTheme = model.DefaultTheme.ValueString()
	}
	if !model.PlatformName.IsNull() && !model.PlatformName.IsUnknown() {
		appearance.PlatformName = model.PlatformName.ValueString()
	}
	if !model.PlatformDescription.IsNull() && !model.PlatformDescription.IsUnknown() {
		appearance.PlatformDescription = model.PlatformDescription.ValueString()
	}
	if !model.CompanyName.IsNull() && !model.CompanyName.IsUnknown() {
		appearance.CompanyName = model.CompanyName.ValueString()
	}

	if !model.HiddenPages.IsNull() && !model.HiddenPages.IsUnknown() {
		var pages []string
		d := model.HiddenPages.ElementsAs(ctx, &pages, false)
		diags.Append(d...)
		if !d.HasError() {
			appearance.HiddenPages = pages
		}
	}

	return appearance
}

func buildAnnouncementFromModel(ctx context.Context, model *TenantSettingsResourceModel, diags *diag.Diagnostics) *client.AnnouncementSettings {
	if model.Announcement.IsNull() || model.Announcement.IsUnknown() {
		return nil
	}

	var announcement AnnouncementSettingsModel
	d := model.Announcement.As(ctx, &announcement, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if d.HasError() {
		return nil
	}

	return &client.AnnouncementSettings{
		Enabled:  announcement.Enabled.ValueBool(),
		Message:  announcement.Message.ValueString(),
		Type:     announcement.Type.ValueString(),
		Pinned:   announcement.Pinned.ValueBool(),
		LinkURL:  announcement.LinkURL.ValueString(),
		LinkText: announcement.LinkText.ValueString(),
	}
}

func buildForgeFromModel(ctx context.Context, model *TenantSettingsResourceModel, diags *diag.Diagnostics) *client.ForgeSettings {
	if model.Forge.IsNull() || model.Forge.IsUnknown() {
		return nil
	}

	var forge ForgeSettingsModel
	d := model.Forge.As(ctx, &forge, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if d.HasError() {
		return nil
	}

	result := &client.ForgeSettings{
		DefaultOrg: forge.DefaultOrg.ValueString(),
	}

	if !forge.AllowedOrgs.IsNull() && !forge.AllowedOrgs.IsUnknown() {
		var orgs []string
		d := forge.AllowedOrgs.ElementsAs(ctx, &orgs, false)
		diags.Append(d...)
		if !d.HasError() {
			result.AllowedOrgs = orgs
		}
	}

	return result
}

func mapSettingsToState(ctx context.Context, settings *client.TenantSettings, state *TenantSettingsResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(settings.ID)

	// For Optional fields, preserve plan/state value when API returns empty to avoid "" -> null drift
	state.PrimaryColor = preserveOrNull(settings.Appearance.PrimaryColor, state.PrimaryColor)
	state.SecondaryColor = preserveOrNull(settings.Appearance.SecondaryColor, state.SecondaryColor)
	state.AccentColor = preserveOrNull(settings.Appearance.AccentColor, state.AccentColor)
	state.LogoURL = preserveOrNull(settings.Appearance.LogoURL, state.LogoURL)
	state.FaviconURL = preserveOrNull(settings.Appearance.FaviconURL, state.FaviconURL)
	state.DefaultTheme = preserveOrNull(settings.Appearance.DefaultTheme, state.DefaultTheme)
	state.PlatformName = preserveOrNull(settings.Appearance.PlatformName, state.PlatformName)
	state.PlatformDescription = preserveOrNull(settings.Appearance.PlatformDescription, state.PlatformDescription)
	state.CompanyName = preserveOrNull(settings.Appearance.CompanyName, state.CompanyName)

	// Map hidden_pages
	if len(settings.Appearance.HiddenPages) > 0 {
		pages := make([]attr.Value, len(settings.Appearance.HiddenPages))
		for i, p := range settings.Appearance.HiddenPages {
			pages[i] = types.StringValue(p)
		}
		hiddenPagesList, d := types.ListValue(types.StringType, pages)
		diags.Append(d...)
		if !d.HasError() {
			state.HiddenPages = hiddenPagesList
		}
	} else if state.HiddenPages.IsNull() || state.HiddenPages.IsUnknown() {
		state.HiddenPages = types.ListNull(types.StringType)
	}
	// else: API returned empty but user set hidden_pages = [] -- keep their empty list

	// Map announcement if present
	announcementTypes := map[string]attr.Type{
		"enabled":    types.BoolType,
		"message":    types.StringType,
		"type":       types.StringType,
		"pinned":     types.BoolType,
		"link_url":   types.StringType,
		"link_text":  types.StringType,
		"updated_at": types.StringType,
	}
	if settings.Announcement.Enabled || settings.Announcement.Message != "" {
		announcementAttrs := map[string]attr.Value{
			"enabled":    types.BoolValue(settings.Announcement.Enabled),
			"message":    types.StringValue(settings.Announcement.Message),
			"type":       types.StringValue(settings.Announcement.Type),
			"pinned":     types.BoolValue(settings.Announcement.Pinned),
			"link_url":   types.StringValue(settings.Announcement.LinkURL),
			"link_text":  types.StringValue(settings.Announcement.LinkText),
			"updated_at": types.StringValue(settings.Announcement.UpdatedAt),
		}

		announcementObj, d := types.ObjectValue(announcementTypes, announcementAttrs)
		diags.Append(d...)
		if !d.HasError() {
			state.Announcement = announcementObj
		}
	} else {
		state.Announcement = types.ObjectNull(announcementTypes)
	}

	// Map forge if present
	forgeTypes := map[string]attr.Type{
		"allowed_orgs": types.ListType{ElemType: types.StringType},
		"default_org":  types.StringType,
	}
	if len(settings.Forge.AllowedOrgs) > 0 || settings.Forge.DefaultOrg != "" {
		orgs := make([]attr.Value, len(settings.Forge.AllowedOrgs))
		for i, o := range settings.Forge.AllowedOrgs {
			orgs[i] = types.StringValue(o)
		}
		var allowedOrgsList basetypes.ListValue
		if len(orgs) > 0 {
			var d diag.Diagnostics
			allowedOrgsList, d = types.ListValue(types.StringType, orgs)
			diags.Append(d...)
		} else {
			allowedOrgsList = types.ListValueMust(types.StringType, []attr.Value{})
		}

		forgeAttrs := map[string]attr.Value{
			"allowed_orgs": allowedOrgsList,
			"default_org":  types.StringValue(settings.Forge.DefaultOrg),
		}

		forgeObj, d := types.ObjectValue(forgeTypes, forgeAttrs)
		diags.Append(d...)
		if !d.HasError() {
			state.Forge = forgeObj
		}
	} else {
		state.Forge = types.ObjectNull(forgeTypes)
	}

	state.CreatedAt = stringValueOrNull(settings.CreatedAt)
	state.UpdatedAt = stringValueOrNull(settings.UpdatedAt)
}
