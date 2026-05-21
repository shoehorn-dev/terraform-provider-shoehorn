package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var (
	_ resource.Resource                = &NotificationSubscriptionResource{}
	_ resource.ResourceWithImportState = &NotificationSubscriptionResource{}
)

// NotificationSubscriptionResource defines the resource implementation.
type NotificationSubscriptionResource struct {
	client *client.Client
}

// NotificationSubscriptionResourceModel describes the resource data model.
type NotificationSubscriptionResourceModel struct {
	ID            types.String        `tfsdk:"id"`
	Scope         types.String        `tfsdk:"scope"`
	ScopeID       types.String        `tfsdk:"scope_id"`
	Name          types.String        `tfsdk:"name"`
	Enabled       types.Bool          `tfsdk:"enabled"`
	EventTypes    types.Set           `tfsdk:"event_types"`
	MinSeverity   types.String        `tfsdk:"min_severity"`
	EntityFilter  types.String        `tfsdk:"entity_filter"`
	ChannelType   types.String        `tfsdk:"channel_type"`
	Cadence       types.String        `tfsdk:"cadence"`
	CadenceConfig types.String        `tfsdk:"cadence_config"`
	Slack         *slackConfigModel   `tfsdk:"slack"`
	Webhook       *webhookConfigModel `tfsdk:"webhook"`
	Email         *emailConfigModel   `tfsdk:"email"`
	CreatedAt     types.String        `tfsdk:"created_at"`
	UpdatedAt     types.String        `tfsdk:"updated_at"`
}

type slackConfigModel struct {
	Mode        types.String `tfsdk:"mode"`
	Channel     types.String `tfsdk:"channel"`
	URLSecret   types.String `tfsdk:"url_secret"`
	TokenSecret types.String `tfsdk:"token_secret"`
}

type webhookConfigModel struct {
	URL           types.String `tfsdk:"url"`
	SigningSecret types.String `tfsdk:"signing_secret"`
	ContentType   types.String `tfsdk:"content_type"`
}

type emailConfigModel struct {
	Recipients types.List `tfsdk:"recipients"`
}

// NewNotificationSubscriptionResource creates a new notification subscription resource.
func NewNotificationSubscriptionResource() resource.Resource {
	return &NotificationSubscriptionResource{}
}

func (r *NotificationSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_subscription"
}

func (r *NotificationSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn notification subscription. A subscription routes platform " +
			"events (CVEs, unhealthy workloads, approvals, digests) to a channel for a team or user. " +
			"Channel secrets are referenced by `secret://` or `env://` ref only. Literal secrets are " +
			"rejected and must be managed through the portal.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the subscription.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scope": schema.StringAttribute{
				Description: "The subscription scope: `team` or `user`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("team", "user"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_id": schema.StringAttribute{
				Description: "The UUID of the team or user the subscription belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the subscription (1-255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the subscription is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"event_types": schema.SetAttribute{
				Description: "The set of event types the subscription matches. Must be non-empty.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"min_severity": schema.StringAttribute{
				Description: "The minimum severity to match: `info`, `warning`, or `critical`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("info", "warning", "critical"),
				},
			},
			"entity_filter": schema.StringAttribute{
				Description: "Optional JSON-encoded entity filter object.",
				Optional:    true,
			},
			"channel_type": schema.StringAttribute{
				Description: "The channel type: `inapp`, `slack`, `email`, or `webhook`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("inapp", "slack", "email", "webhook"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cadence": schema.StringAttribute{
				Description: "The delivery cadence: `realtime`, `hourly`, `daily`, or `weekly`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("realtime", "hourly", "daily", "weekly"),
				},
			},
			"cadence_config": schema.StringAttribute{
				Description: "Optional JSON-encoded cadence configuration object.",
				Optional:    true,
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
		Blocks: map[string]schema.Block{
			"slack": schema.SingleNestedBlock{
				Description: "Slack channel configuration. Set when `channel_type` is `slack`.",
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Slack delivery mode: `webhook` or `bot`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("webhook", "bot"),
						},
					},
					"channel": schema.StringAttribute{
						Description: "Slack channel name. Required for `bot` mode.",
						Optional:    true,
					},
					"url_secret": schema.StringAttribute{
						Description: "Reference to the incoming webhook URL secret (`secret://NAME` or `env://VAR`). Required for `webhook` mode.",
						Optional:    true,
						Validators: []validator.String{
							secretRefValidator{},
						},
					},
					"token_secret": schema.StringAttribute{
						Description: "Reference to the bot token secret (`secret://NAME` or `env://VAR`). Required for `bot` mode.",
						Optional:    true,
						Validators: []validator.String{
							secretRefValidator{},
						},
					},
				},
			},
			"webhook": schema.SingleNestedBlock{
				Description: "Webhook channel configuration. Set when `channel_type` is `webhook`.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "The webhook URL to POST events to.",
						Optional:    true,
					},
					"signing_secret": schema.StringAttribute{
						Description: "Reference to the request-signing secret (`secret://NAME` or `env://VAR`).",
						Optional:    true,
						Validators: []validator.String{
							secretRefValidator{},
						},
					},
					"content_type": schema.StringAttribute{
						Description: "The request content type. Defaults to `application/json`.",
						Optional:    true,
					},
				},
			},
			"email": schema.SingleNestedBlock{
				Description: "Email channel configuration. Set when `channel_type` is `email`.",
				Attributes: map[string]schema.Attribute{
					"recipients": schema.ListAttribute{
						Description: "The list of recipient email addresses.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *NotificationSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "creating notification subscription")

	var plan NotificationSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildSubscriptionRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.CreateNotificationSubscription(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Notification Subscription", fmt.Sprintf("Could not create subscription: %s", err))
		return
	}

	resp.Diagnostics.Append(mapSubscriptionToState(ctx, sub, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "reading notification subscription")

	var state NotificationSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// There is no GET /{id} endpoint. List by scope/scope_id and find the row.
	subs, err := r.client.ListNotificationSubscriptions(ctx, state.Scope.ValueString(), state.ScopeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Notification Subscription",
			fmt.Sprintf("Could not list subscriptions for %s %s: %s", state.Scope.ValueString(), state.ScopeID.ValueString(), err))
		return
	}

	var found *client.NotificationSubscription
	for i := range subs {
		if subs[i].ID == state.ID.ValueString() {
			found = &subs[i]
			break
		}
	}
	if found == nil {
		tflog.Warn(ctx, "notification subscription not found, removing from state", map[string]any{"id": state.ID.ValueString()})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(mapSubscriptionToState(ctx, found, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "updating notification subscription")

	var plan NotificationSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NotificationSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildSubscriptionRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.UpdateNotificationSubscription(ctx, state.ID.ValueString(), apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Notification Subscription",
			fmt.Sprintf("Could not update subscription %s: %s", state.ID.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(mapSubscriptionToState(ctx, sub, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "deleting notification subscription")

	var state NotificationSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNotificationSubscription(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Notification Subscription",
			fmt.Sprintf("Could not delete subscription %s: %s", state.ID.ValueString(), err))
		return
	}
}

// ImportState parses an import ID of the form `scope/scope_id/id`. The 3-part
// form is required because Read has no GET /{id} endpoint. It lists by
// scope/scope_id, so those values must be known before the first Read.
func (r *NotificationSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the form `scope/scope_id/id` (for example "+
				"`team/3f1c.../9a8b...`), got %q. The scope and scope_id are required because "+
				"the subscription is read by listing within its scope.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}

// buildSubscriptionRequest maps a plan model to an API request, building the
// channel_config JSON from the typed nested block matching channel_type.
func buildSubscriptionRequest(ctx context.Context, plan *NotificationSubscriptionResourceModel) (client.NotificationSubscriptionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := client.NotificationSubscriptionRequest{
		Scope:       plan.Scope.ValueString(),
		ScopeID:     plan.ScopeID.ValueString(),
		Name:        plan.Name.ValueString(),
		Enabled:     plan.Enabled.ValueBool(),
		MinSeverity: plan.MinSeverity.ValueString(),
		ChannelType: plan.ChannelType.ValueString(),
		Cadence:     plan.Cadence.ValueString(),
	}

	if !plan.EventTypes.IsNull() && !plan.EventTypes.IsUnknown() {
		var eventTypes []string
		diags.Append(plan.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
		if diags.HasError() {
			return req, diags
		}
		req.EventTypes = eventTypes
	}

	if !plan.EntityFilter.IsNull() && !plan.EntityFilter.IsUnknown() {
		raw, ok := validateRawJSON(plan.EntityFilter.ValueString())
		if !ok {
			diags.AddError("Invalid Entity Filter", "entity_filter must be a valid JSON object.")
			return req, diags
		}
		req.EntityFilter = raw
	}

	if !plan.CadenceConfig.IsNull() && !plan.CadenceConfig.IsUnknown() {
		raw, ok := validateRawJSON(plan.CadenceConfig.ValueString())
		if !ok {
			diags.AddError("Invalid Cadence Config", "cadence_config must be a valid JSON object.")
			return req, diags
		}
		req.CadenceConfig = raw
	}

	channelConfig, cfgDiags := buildChannelConfig(ctx, plan)
	diags.Append(cfgDiags...)
	if diags.HasError() {
		return req, diags
	}
	req.ChannelConfig = channelConfig

	return req, diags
}

// buildChannelConfig builds the channel_config JSON from the nested block that
// matches channel_type. Each *_secret ref is wrapped as {"mode":"ref","value":"<ref>"}.
func buildChannelConfig(_ context.Context, plan *NotificationSubscriptionResourceModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	channelType := plan.ChannelType.ValueString()

	switch channelType {
	case "inapp":
		return nil, diags
	case "slack":
		if plan.Slack == nil {
			diags.AddError("Missing Slack Block", "channel_type is \"slack\" but no `slack` block was set.")
			return nil, diags
		}
		cfg := map[string]interface{}{}
		if !plan.Slack.Mode.IsNull() {
			cfg["mode"] = plan.Slack.Mode.ValueString()
		}
		if !plan.Slack.Channel.IsNull() {
			cfg["channel"] = plan.Slack.Channel.ValueString()
		}
		if !plan.Slack.URLSecret.IsNull() {
			cfg["url_secret"] = secretRef(plan.Slack.URLSecret.ValueString())
		}
		if !plan.Slack.TokenSecret.IsNull() {
			cfg["token_secret"] = secretRef(plan.Slack.TokenSecret.ValueString())
		}
		return marshalChannelConfig(cfg, diags)
	case "webhook":
		if plan.Webhook == nil {
			diags.AddError("Missing Webhook Block", "channel_type is \"webhook\" but no `webhook` block was set.")
			return nil, diags
		}
		cfg := map[string]interface{}{}
		if !plan.Webhook.URL.IsNull() {
			cfg["url"] = plan.Webhook.URL.ValueString()
		}
		if !plan.Webhook.SigningSecret.IsNull() {
			cfg["signing_secret"] = secretRef(plan.Webhook.SigningSecret.ValueString())
		}
		if !plan.Webhook.ContentType.IsNull() {
			cfg["content_type"] = plan.Webhook.ContentType.ValueString()
		}
		return marshalChannelConfig(cfg, diags)
	case "email":
		if plan.Email == nil {
			diags.AddError("Missing Email Block", "channel_type is \"email\" but no `email` block was set.")
			return nil, diags
		}
		cfg := map[string]interface{}{}
		if !plan.Email.Recipients.IsNull() && !plan.Email.Recipients.IsUnknown() {
			var recipients []string
			diags.Append(plan.Email.Recipients.ElementsAs(context.Background(), &recipients, false)...)
			cfg["recipients"] = recipients
		}
		return marshalChannelConfig(cfg, diags)
	default:
		diags.AddError("Unknown Channel Type", fmt.Sprintf("Unsupported channel_type %q.", channelType))
		return nil, diags
	}
}

// marshalChannelConfig encodes the channel config map, appending any error to diags.
func marshalChannelConfig(cfg map[string]interface{}, diags diag.Diagnostics) ([]byte, diag.Diagnostics) {
	b, err := json.Marshal(cfg)
	if err != nil {
		diags.AddError("Channel Config Error", fmt.Sprintf("Could not encode channel_config: %s", err))
		return nil, diags
	}
	return b, diags
}

// secretRef wraps a ref string into the {"mode":"ref","value":"<ref>"} envelope
// the API expects for channel secrets.
func secretRef(ref string) map[string]interface{} {
	return map[string]interface{}{"mode": "ref", "value": ref}
}

// mapSubscriptionToState maps an API subscription to the resource state model,
// unwrapping the channel_config back into the typed nested block.
func mapSubscriptionToState(ctx context.Context, sub *client.NotificationSubscription, state *NotificationSubscriptionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(sub.ID)
	state.Scope = types.StringValue(sub.Scope)
	state.ScopeID = types.StringValue(sub.ScopeID)
	state.Name = types.StringValue(sub.Name)
	state.Enabled = types.BoolValue(sub.Enabled)
	state.MinSeverity = types.StringValue(sub.MinSeverity)
	state.ChannelType = types.StringValue(sub.ChannelType)
	state.Cadence = types.StringValue(sub.Cadence)
	state.CreatedAt = stringValueOrNull(sub.CreatedAt)
	state.UpdatedAt = stringValueOrNull(sub.UpdatedAt)

	eventTypes := sub.EventTypes
	if eventTypes == nil {
		eventTypes = []string{}
	}
	setVal, setDiags := types.SetValueFrom(ctx, types.StringType, eventTypes)
	diags.Append(setDiags...)
	if diags.HasError() {
		return diags
	}
	state.EventTypes = setVal

	state.EntityFilter = rawJSONToState(sub.EntityFilter, state.EntityFilter)
	state.CadenceConfig = rawJSONToState(sub.CadenceConfig, state.CadenceConfig)

	diags.Append(mapChannelConfigToState(ctx, sub, state)...)
	return diags
}

// mapChannelConfigToState unwraps the channel_config JSON into the typed nested
// block matching channel_type. Secret refs come back as {"mode":"ref","value"}.
func mapChannelConfigToState(ctx context.Context, sub *client.NotificationSubscription, state *NotificationSubscriptionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Slack = nil
	state.Webhook = nil
	state.Email = nil

	if len(sub.ChannelConfig) == 0 {
		return diags
	}

	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(sub.ChannelConfig, &cfg); err != nil {
		diags.AddError("Invalid Channel Config", fmt.Sprintf("Could not parse channel_config: %s", err))
		return diags
	}

	switch sub.ChannelType {
	case "slack":
		state.Slack = &slackConfigModel{
			Mode:        rawStringField(cfg, "mode"),
			Channel:     rawStringField(cfg, "channel"),
			URLSecret:   rawSecretField(cfg, "url_secret"),
			TokenSecret: rawSecretField(cfg, "token_secret"),
		}
	case "webhook":
		state.Webhook = &webhookConfigModel{
			URL:           rawStringField(cfg, "url"),
			SigningSecret: rawSecretField(cfg, "signing_secret"),
			ContentType:   rawStringField(cfg, "content_type"),
		}
	case "email":
		recipients := []string{}
		if raw, ok := cfg["recipients"]; ok {
			_ = json.Unmarshal(raw, &recipients)
		}
		listVal, listDiags := types.ListValueFrom(ctx, types.StringType, recipients)
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		state.Email = &emailConfigModel{Recipients: listVal}
	}

	return diags
}

// rawStringField extracts a plain string field from a channel_config map.
func rawStringField(cfg map[string]json.RawMessage, key string) types.String {
	raw, ok := cfg[key]
	if !ok {
		return types.StringNull()
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil || s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// rawSecretField extracts a secret field, unwrapping the {"mode":"ref","value"}
// envelope back to the ref string. Ref values come back verbatim.
func rawSecretField(cfg map[string]json.RawMessage, key string) types.String {
	raw, ok := cfg[key]
	if !ok {
		return types.StringNull()
	}
	var envelope struct {
		Mode  string `json:"mode"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Value == "" {
		return types.StringNull()
	}
	return types.StringValue(envelope.Value)
}

// rawJSONToState maps a raw JSON message to a string state attribute, preserving
// the prior state value when the API omits the field.
func rawJSONToState(raw json.RawMessage, current types.String) types.String {
	if len(raw) == 0 {
		if current.IsUnknown() {
			return types.StringNull()
		}
		return current
	}
	return types.StringValue(string(raw))
}

// validateRawJSON confirms a string is a valid JSON object and returns it as raw bytes.
func validateRawJSON(s string) (json.RawMessage, bool) {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return nil, false
	}
	return json.RawMessage(s), true
}
