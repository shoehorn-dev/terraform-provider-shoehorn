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

var _ datasource.DataSource = &GitOpsResourcesDataSource{}

// GitOpsResourcesDataSource defines the data source implementation.
type GitOpsResourcesDataSource struct {
	client *client.Client
}

// GitOpsResourcesDataSourceModel describes the data source data model.
type GitOpsResourcesDataSourceModel struct {
	ClusterID    types.String           `tfsdk:"cluster_id"`
	Tool         types.String           `tfsdk:"tool"`
	SyncStatus   types.String           `tfsdk:"sync_status"`
	HealthStatus types.String           `tfsdk:"health_status"`
	Total        types.Int64            `tfsdk:"total"`
	Resources    []GitOpsResourceModel  `tfsdk:"resources"`
}

// GitOpsResourceModel describes a single GitOps resource.
type GitOpsResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ClusterID      types.String `tfsdk:"cluster_id"`
	Tool           types.String `tfsdk:"tool"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	Kind           types.String `tfsdk:"kind"`
	SyncStatus     types.String `tfsdk:"sync_status"`
	HealthStatus   types.String `tfsdk:"health_status"`
	SourceURL      types.String `tfsdk:"source_url"`
	Revision       types.String `tfsdk:"revision"`
	TargetRevision types.String `tfsdk:"target_revision"`
	AutoSync       types.Bool   `tfsdk:"auto_sync"`
	Suspended      types.Bool   `tfsdk:"suspended"`
	EntityID       types.String `tfsdk:"entity_id"`
	EntityName     types.String `tfsdk:"entity_name"`
	OwnerTeam      types.String `tfsdk:"owner_team"`
	LastSyncedAt   types.String `tfsdk:"last_synced_at"`
}

// NewGitOpsResourcesDataSource creates a new GitOps resources data source.
func NewGitOpsResourcesDataSource() datasource.DataSource {
	return &GitOpsResourcesDataSource{}
}

func (d *GitOpsResourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitops_resources"
}

func (d *GitOpsResourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves GitOps resources (ArgoCD Applications, FluxCD Kustomizations, HelmReleases, etc.) pushed by K8s agents.",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "Filter resources by cluster ID.",
				Optional:    true,
			},
			"tool": schema.StringAttribute{
				Description: "Filter resources by GitOps tool (argocd, fluxcd).",
				Optional:    true,
			},
			"sync_status": schema.StringAttribute{
				Description: "Filter resources by sync status.",
				Optional:    true,
			},
			"health_status": schema.StringAttribute{
				Description: "Filter resources by health status.",
				Optional:    true,
			},
			"total": schema.Int64Attribute{
				Description: "Total number of resources matching the filters.",
				Computed:    true,
			},
			"resources": schema.ListNestedAttribute{
				Description: "The list of GitOps resources.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the resource.",
							Computed:    true,
						},
						"cluster_id": schema.StringAttribute{
							Description: "The cluster ID where this resource lives.",
							Computed:    true,
						},
						"tool": schema.StringAttribute{
							Description: "The GitOps tool (argocd, fluxcd).",
							Computed:    true,
						},
						"namespace": schema.StringAttribute{
							Description: "The Kubernetes namespace.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The resource name.",
							Computed:    true,
						},
						"kind": schema.StringAttribute{
							Description: "The resource kind (Application, Kustomization, HelmRelease).",
							Computed:    true,
						},
						"sync_status": schema.StringAttribute{
							Description: "The sync status of the resource.",
							Computed:    true,
						},
						"health_status": schema.StringAttribute{
							Description: "The health status of the resource.",
							Computed:    true,
						},
						"source_url": schema.StringAttribute{
							Description: "The Git source URL.",
							Computed:    true,
						},
						"revision": schema.StringAttribute{
							Description: "The current revision (commit SHA or tag).",
							Computed:    true,
						},
						"target_revision": schema.StringAttribute{
							Description: "The target revision (branch, tag, or commit).",
							Computed:    true,
						},
						"auto_sync": schema.BoolAttribute{
							Description: "Whether auto-sync is enabled.",
							Computed:    true,
						},
						"suspended": schema.BoolAttribute{
							Description: "Whether the resource is suspended.",
							Computed:    true,
						},
						"entity_id": schema.StringAttribute{
							Description: "The Shoehorn entity ID this resource is linked to.",
							Computed:    true,
						},
						"entity_name": schema.StringAttribute{
							Description: "The Shoehorn entity name this resource is linked to.",
							Computed:    true,
						},
						"owner_team": schema.StringAttribute{
							Description: "The owning team.",
							Computed:    true,
						},
						"last_synced_at": schema.StringAttribute{
							Description: "The last sync timestamp.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *GitOpsResourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GitOpsResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "reading gitops resources data source")

	var config GitOpsResourcesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := client.ListGitOpsResourcesParams{}
	if !config.ClusterID.IsNull() && !config.ClusterID.IsUnknown() {
		params.ClusterID = config.ClusterID.ValueString()
	}
	if !config.Tool.IsNull() && !config.Tool.IsUnknown() {
		params.Tool = config.Tool.ValueString()
	}
	if !config.SyncStatus.IsNull() && !config.SyncStatus.IsUnknown() {
		params.SyncStatus = config.SyncStatus.ValueString()
	}
	if !config.HealthStatus.IsNull() && !config.HealthStatus.IsUnknown() {
		params.HealthStatus = config.HealthStatus.ValueString()
	}

	resources, total, err := d.client.ListGitOpsResources(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading GitOps Resources", fmt.Sprintf("Could not list gitops resources: %s", err))
		return
	}

	state := GitOpsResourcesDataSourceModel{
		ClusterID:    config.ClusterID,
		Tool:         config.Tool,
		SyncStatus:   config.SyncStatus,
		HealthStatus: config.HealthStatus,
		Total:        types.Int64Value(int64(total)),
	}

	for _, r := range resources {
		model := GitOpsResourceModel{
			ID:             types.StringValue(r.ID),
			ClusterID:      types.StringValue(r.ClusterID),
			Tool:           types.StringValue(r.Tool),
			Namespace:      types.StringValue(r.Namespace),
			Name:           types.StringValue(r.Name),
			Kind:           types.StringValue(r.Kind),
			SyncStatus:     types.StringValue(r.SyncStatus),
			HealthStatus:   types.StringValue(r.HealthStatus),
			SourceURL:      types.StringValue(r.SourceURL),
			Revision:       types.StringValue(r.Revision),
			TargetRevision: types.StringValue(r.TargetRevision),
			AutoSync:       types.BoolValue(r.AutoSync),
			Suspended:      types.BoolValue(r.Suspended),
			EntityID:       types.StringValue(r.EntityID),
			EntityName:     types.StringValue(r.EntityName),
			OwnerTeam:      types.StringValue(r.OwnerTeam),
			LastSyncedAt:   types.StringValue(r.LastSyncedAt),
		}
		state.Resources = append(state.Resources, model)
	}

	if state.Resources == nil {
		state.Resources = []GitOpsResourceModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
