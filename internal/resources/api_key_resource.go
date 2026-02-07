package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ resource.Resource = &APIKeyResource{}

// APIKeyResource defines the resource implementation.
type APIKeyResource struct {
	client *client.Client
}

// APIKeyResourceModel describes the resource data model.
type APIKeyResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Scopes        types.List   `tfsdk:"scopes"`
	ExpiresInDays types.Int64  `tfsdk:"expires_in_days"`
	KeyPrefix     types.String `tfsdk:"key_prefix"`
	RawKey        types.String `tfsdk:"raw_key"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

// NewAPIKeyResource creates a new API key resource.
func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn API key. The raw key is only available on creation and stored in state. Deleting this resource revokes the key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the API key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the API key.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.ListAttribute{
				Description: "The scopes granted to this API key.",
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"expires_in_days": schema.Int64Attribute{
				Description: "Number of days until the key expires. Null means never expires.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"key_prefix": schema.StringAttribute{
				Description: "The prefix of the API key (for identification).",
				Computed:    true,
			},
			"raw_key": schema.StringAttribute{
				Description: "The full API key value. Only available on creation.",
				Computed:    true,
				Sensitive:   true,
			},
			"expires_at": schema.StringAttribute{
				Description: "The expiration timestamp of the key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateAPIKeyRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Scopes:      scopes,
	}

	if !plan.ExpiresInDays.IsNull() && !plan.ExpiresInDays.IsUnknown() {
		days := int(plan.ExpiresInDays.ValueInt64())
		createReq.ExpiresInDays = &days
	}

	createResp, err := r.client.CreateAPIKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating API Key", fmt.Sprintf("Could not create API key: %s", err))
		return
	}

	plan.ID = types.StringValue(createResp.Key.ID)
	plan.KeyPrefix = types.StringValue(createResp.Key.KeyPrefix)
	plan.RawKey = types.StringValue(createResp.RawKey)

	if createResp.Key.ExpiresAt != "" {
		plan.ExpiresAt = types.StringValue(createResp.Key.ExpiresAt)
	}
	if createResp.Key.CreatedAt != "" {
		plan.CreatedAt = types.StringValue(createResp.Key.CreatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.GetAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading API Key", fmt.Sprintf("Could not read API key %s: %s", state.ID.ValueString(), err))
		return
	}

	// If key is revoked, remove from state
	if apiKey.RevokedAt != "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(apiKey.Name)
	state.KeyPrefix = types.StringValue(apiKey.KeyPrefix)
	if apiKey.ExpiresAt != "" {
		state.ExpiresAt = types.StringValue(apiKey.ExpiresAt)
	}
	if apiKey.CreatedAt != "" {
		state.CreatedAt = types.StringValue(apiKey.CreatedAt)
	}
	// raw_key is preserved from state - not returned by API after creation

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *APIKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API keys are immutable - all mutable fields have RequiresReplace
	resp.Diagnostics.AddError("Update Not Supported", "API keys cannot be updated. Changes require creating a new key.")
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RevokeAPIKey(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Revoking API Key", fmt.Sprintf("Could not revoke API key %s: %s", state.ID.ValueString(), err))
		return
	}
}
