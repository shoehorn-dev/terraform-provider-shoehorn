package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

var _ datasource.DataSource = &GovernanceActionsDataSource{}

// GovernanceActionsDataSource defines the data source implementation.
type GovernanceActionsDataSource struct {
	client *client.Client
}

// GovernanceActionsDataSourceModel describes the data source data model.
type GovernanceActionsDataSourceModel struct {
	Status     types.String               `tfsdk:"status"`
	Priority   types.String               `tfsdk:"priority"`
	EntityID   types.String               `tfsdk:"entity_id"`
	SourceType types.String               `tfsdk:"source_type"`
	Overdue    types.Bool                 `tfsdk:"overdue"`
	Total      types.Int64                `tfsdk:"total"`
	Actions    []GovernanceActionModel    `tfsdk:"actions"`
}

// GovernanceActionModel describes a single governance action in the list.
type GovernanceActionModel struct {
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

// NewGovernanceActionsDataSource creates a new governance actions data source.
func NewGovernanceActionsDataSource() datasource.DataSource {
	return &GovernanceActionsDataSource{}
}

func (d *GovernanceActionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_governance_actions"
}

func (d *GovernanceActionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists governance actions with optional filters.",
		Attributes: map[string]schema.Attribute{
			"status": schema.StringAttribute{
				Description: "Filter actions by status (open, in_progress, resolved, dismissed, wont_fix).",
				Optional:    true,
			},
			"priority": schema.StringAttribute{
				Description: "Filter actions by priority (critical, high, medium, low).",
				Optional:    true,
			},
			"entity_id": schema.StringAttribute{
				Description: "Filter actions by entity ID.",
				Optional:    true,
			},
			"source_type": schema.StringAttribute{
				Description: "Filter actions by source type (scorecard, security, policy).",
				Optional:    true,
			},
			"overdue": schema.BoolAttribute{
				Description: "Filter to only overdue actions.",
				Optional:    true,
			},
			"total": schema.Int64Attribute{
				Description: "Total number of actions matching the filters.",
				Computed:    true,
			},
			"actions": schema.ListNestedAttribute{
				Description: "The list of governance actions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the governance action.",
							Computed:    true,
						},
						"entity_id": schema.StringAttribute{
							Description: "The entity ID this action is associated with.",
							Computed:    true,
						},
						"entity_name": schema.StringAttribute{
							Description: "The display name of the associated entity.",
							Computed:    true,
						},
						"title": schema.StringAttribute{
							Description: "The title of the governance action.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of the governance action.",
							Computed:    true,
						},
						"priority": schema.StringAttribute{
							Description: "The priority level (critical, high, medium, low).",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The action status.",
							Computed:    true,
						},
						"source_type": schema.StringAttribute{
							Description: "The source type that created this action.",
							Computed:    true,
						},
						"source_id": schema.StringAttribute{
							Description: "The ID of the source that created this action.",
							Computed:    true,
						},
						"assigned_to": schema.StringAttribute{
							Description: "The user or team this action is assigned to.",
							Computed:    true,
						},
						"sla_days": schema.Int64Attribute{
							Description: "The SLA in days for resolving this action.",
							Computed:    true,
						},
						"due_date": schema.StringAttribute{
							Description: "The computed due date based on SLA.",
							Computed:    true,
						},
						"resolution_note": schema.StringAttribute{
							Description: "A note explaining how the action was resolved.",
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
				},
			},
		},
	}
}

func (d *GovernanceActionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *GovernanceActionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "reading governance actions data source")

	var config GovernanceActionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build filters from optional attributes
	var filters *client.GovernanceActionFilters
	hasFilters := !config.Status.IsNull() || !config.Priority.IsNull() ||
		!config.EntityID.IsNull() || !config.SourceType.IsNull() || !config.Overdue.IsNull()

	if hasFilters {
		filters = &client.GovernanceActionFilters{}
		if !config.Status.IsNull() {
			filters.Status = config.Status.ValueString()
		}
		if !config.Priority.IsNull() {
			filters.Priority = config.Priority.ValueString()
		}
		if !config.EntityID.IsNull() {
			filters.EntityID = config.EntityID.ValueString()
		}
		if !config.SourceType.IsNull() {
			filters.SourceType = config.SourceType.ValueString()
		}
		if !config.Overdue.IsNull() {
			v := config.Overdue.ValueBool()
			filters.Overdue = &v
		}
	}

	actions, total, err := d.client.ListGovernanceActions(ctx, filters)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Governance Actions", fmt.Sprintf("Could not list governance actions: %s", err))
		return
	}

	state := GovernanceActionsDataSourceModel{
		Status:     config.Status,
		Priority:   config.Priority,
		EntityID:   config.EntityID,
		SourceType: config.SourceType,
		Overdue:    config.Overdue,
		Total:      types.Int64Value(int64(total)),
	}

	for _, a := range actions {
		model := GovernanceActionModel{
			ID:             types.StringValue(a.ID),
			EntityID:       types.StringValue(a.EntityID),
			EntityName:     stringValueOrNull(a.EntityName),
			Title:          types.StringValue(a.Title),
			Description:    stringValueOrNull(a.Description),
			Priority:       types.StringValue(a.Priority),
			Status:         stringValueOrNull(a.Status),
			SourceType:     types.StringValue(a.SourceType),
			SourceID:       stringValueOrNull(a.SourceID),
			AssignedTo:     stringValueOrNull(a.AssignedTo),
			DueDate:        stringValueOrNull(a.DueDate),
			ResolutionNote: stringValueOrNull(a.ResolutionNote),
			CreatedAt:      stringValueOrNull(a.CreatedAt),
			UpdatedAt:      stringValueOrNull(a.UpdatedAt),
		}

		if a.SLADays != nil {
			model.SLADays = types.Int64Value(int64(*a.SLADays))
		} else {
			model.SLADays = types.Int64Null()
		}

		state.Actions = append(state.Actions, model)
	}

	if state.Actions == nil {
		state.Actions = []GovernanceActionModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// stringValueOrNull returns a types.StringValue for non-empty strings,
// or types.StringNull for empty strings.
func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
