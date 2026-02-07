package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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
	_ resource.Resource                = &EntityResource{}
	_ resource.ResourceWithImportState = &EntityResource{}
)

// EntityResource defines the resource implementation.
type EntityResource struct {
	client *client.Client
}

// EntityResourceModel describes the resource data model.
type EntityResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Description    types.String `tfsdk:"description"`
	Lifecycle      types.String `tfsdk:"entity_lifecycle"`
	Tier           types.String `tfsdk:"tier"`
	Owner          types.String `tfsdk:"owner"`
	Tags           types.Set    `tfsdk:"tags"`
	Links          types.String `tfsdk:"links"`
	Relations      types.String `tfsdk:"relations"`
	Licenses       types.String `tfsdk:"licenses"`
	ChangelogPath  types.String `tfsdk:"changelog_path"`
	Interfaces     types.String `tfsdk:"interfaces"`
	RepositoryPath types.String `tfsdk:"repository_path"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// NewEntityResource creates a new entity resource.
func NewEntityResource() resource.Resource {
	return &EntityResource{}
}

func (r *EntityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (r *EntityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Shoehorn catalog entity via manifest.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The service ID (unique identifier) of the entity.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the entity (used as service ID).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of the entity (e.g., service, library, website).",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the entity.",
				Optional:    true,
			},
			"entity_lifecycle": schema.StringAttribute{
				Description: "The lifecycle stage (e.g., experimental, production, deprecated).",
				Optional:    true,
				Computed:    true,
			},
			"tier": schema.StringAttribute{
				Description: "The tier of the entity (e.g., tier1, tier2, tier3).",
				Optional:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The owner team slug for the entity.",
				Optional:    true,
			},
			"tags": schema.SetAttribute{
				Description: "Tags for the entity.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"links": schema.StringAttribute{
				Description: "JSON-encoded array of links for the entity. Each link has name, url, and optional icon.",
				Optional:    true,
			},
			"relations": schema.StringAttribute{
				Description: "JSON-encoded array of relations. Each relation has type (depends_on, calls, etc.), target (type:id), and optional via.",
				Optional:    true,
			},
			"licenses": schema.StringAttribute{
				Description: "JSON-encoded array of licenses. Each license has title (required), and optional vendor, purchased, expires, seats, cost, contract, notes.",
				Optional:    true,
			},
			"changelog_path": schema.StringAttribute{
				Description: "Path to the changelog file (e.g., CHANGELOG.md).",
				Optional:    true,
			},
			"interfaces": schema.StringAttribute{
				Description: "JSON-encoded interfaces definition. Supports http (with openapi, baseUrl, auth, graphql) and grpc (with package, proto).",
				Optional:    true,
			},
			"repository_path": schema.StringAttribute{
				Description: "The repository path for the entity (e.g., github:org/repo). Computed from the API.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *EntityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EntityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EntityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	manifest := buildManifestYAML(&plan)

	createResp, err := r.client.CreateEntity(ctx, client.CreateEntityRequest{
		Content: manifest,
		Source:  "terraform",
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Entity", fmt.Sprintf("Could not create entity: %s", err))
		return
	}

	// Set the service ID from the response
	plan.ID = types.StringValue(createResp.Entity.ServiceID)
	plan.CreatedAt = types.StringValue(createResp.Entity.CreatedAt)
	plan.UpdatedAt = types.StringValue(createResp.Entity.UpdatedAt)

	// Set lifecycle from response if not specified
	if plan.Lifecycle.IsNull() || plan.Lifecycle.IsUnknown() {
		plan.Lifecycle = types.StringValue(createResp.Entity.Lifecycle)
	}

	// Set computed repository_path if unknown
	if plan.RepositoryPath.IsUnknown() {
		plan.RepositoryPath = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EntityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EntityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save existing state values for order-insensitive comparison
	prevRelations := state.Relations
	prevLinks := state.Links
	prevLicenses := state.Licenses
	prevInterfaces := state.Interfaces

	entity, err := r.client.GetEntity(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Entity", fmt.Sprintf("Could not read entity %s: %s", state.ID.ValueString(), err))
		return
	}

	mapEntityToState(ctx, entity, &state)

	// Preserve original relations order if semantically equivalent
	if !prevRelations.IsNull() && !state.Relations.IsNull() {
		if relationsEquivalent(prevRelations.ValueString(), state.Relations.ValueString()) {
			state.Relations = prevRelations
		}
	}

	// Preserve original links order if semantically equivalent
	if !prevLinks.IsNull() && !state.Links.IsNull() {
		if linksEquivalent(prevLinks.ValueString(), state.Links.ValueString()) {
			state.Links = prevLinks
		}
	}

	// Preserve original interfaces if semantically equivalent
	if !prevInterfaces.IsNull() && !state.Interfaces.IsNull() {
		if interfacesEquivalent(prevInterfaces.ValueString(), state.Interfaces.ValueString()) {
			state.Interfaces = prevInterfaces
		}
	}

	// Preserve original licenses order if semantically equivalent
	if !prevLicenses.IsNull() && !state.Licenses.IsNull() {
		if licensesEquivalent(prevLicenses.ValueString(), state.Licenses.ValueString()) {
			state.Licenses = prevLicenses
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EntityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EntityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state EntityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	manifest := buildManifestYAML(&plan)

	updateResp, err := r.client.UpdateEntity(ctx, state.ID.ValueString(), client.CreateEntityRequest{
		Content: manifest,
		Source:  "terraform",
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Entity", fmt.Sprintf("Could not update entity %s: %s", state.ID.ValueString(), err))
		return
	}

	plan.ID = state.ID
	plan.CreatedAt = types.StringValue(updateResp.Entity.CreatedAt)
	plan.UpdatedAt = types.StringValue(updateResp.Entity.UpdatedAt)

	if plan.Lifecycle.IsNull() || plan.Lifecycle.IsUnknown() {
		plan.Lifecycle = types.StringValue(updateResp.Entity.Lifecycle)
	}

	// Preserve computed repository_path from state
	if plan.RepositoryPath.IsUnknown() {
		plan.RepositoryPath = state.RepositoryPath
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EntityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EntityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEntity(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Entity", fmt.Sprintf("Could not delete entity %s: %s", state.ID.ValueString(), err))
		return
	}
}

func (r *EntityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildManifestYAML generates the YAML manifest content from the resource model.
// The manifest follows the Shoehorn catalog manifest spec (schemaVersion: 1).
func buildManifestYAML(model *EntityResourceModel) string {
	var b strings.Builder

	b.WriteString("schemaVersion: 1\n\n")

	// Service block (required: id, name, type)
	b.WriteString("service:\n")
	b.WriteString(fmt.Sprintf("  id: %s\n", model.Name.ValueString()))
	b.WriteString(fmt.Sprintf("  name: %s\n", model.Name.ValueString()))
	b.WriteString(fmt.Sprintf("  type: %s\n", model.Type.ValueString()))

	if !model.Tier.IsNull() && !model.Tier.IsUnknown() {
		b.WriteString(fmt.Sprintf("  tier: %s\n", model.Tier.ValueString()))
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		b.WriteString(fmt.Sprintf("\ndescription: %s\n", model.Description.ValueString()))
	}

	if !model.Lifecycle.IsNull() && !model.Lifecycle.IsUnknown() {
		b.WriteString(fmt.Sprintf("\nlifecycle: %s\n", model.Lifecycle.ValueString()))
	}

	if !model.Owner.IsNull() && !model.Owner.IsUnknown() {
		b.WriteString("\nowner:\n")
		b.WriteString("  - type: team\n")
		b.WriteString(fmt.Sprintf("    id: %s\n", model.Owner.ValueString()))
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		for _, elem := range model.Tags.Elements() {
			if strVal, ok := elem.(types.String); ok {
				tags = append(tags, strVal.ValueString())
			}
		}
		if len(tags) > 0 {
			b.WriteString("\ntags:\n")
			for _, tag := range tags {
				b.WriteString(fmt.Sprintf("  - %s\n", tag))
			}
		}
	}

	if !model.Links.IsNull() && !model.Links.IsUnknown() {
		var links []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			Icon string `json:"icon,omitempty"`
		}
		if err := json.Unmarshal([]byte(model.Links.ValueString()), &links); err == nil && len(links) > 0 {
			b.WriteString("\nlinks:\n")
			for _, link := range links {
				b.WriteString(fmt.Sprintf("  - name: %s\n", link.Name))
				b.WriteString(fmt.Sprintf("    url: %s\n", link.URL))
				if link.Icon != "" {
					b.WriteString(fmt.Sprintf("    icon: %s\n", link.Icon))
				}
			}
		}
	}

	if !model.Relations.IsNull() && !model.Relations.IsUnknown() {
		var relations []struct {
			Type   string `json:"type"`
			Target string `json:"target"`
			Via    string `json:"via,omitempty"`
		}
		if err := json.Unmarshal([]byte(model.Relations.ValueString()), &relations); err == nil && len(relations) > 0 {
			b.WriteString("\nrelations:\n")
			for _, rel := range relations {
				b.WriteString(fmt.Sprintf("  - type: %s\n", rel.Type))
				b.WriteString(fmt.Sprintf("    target: %s\n", rel.Target))
				if rel.Via != "" {
					b.WriteString(fmt.Sprintf("    via: %s\n", rel.Via))
				}
			}
		}
	}

	// Build integrations section (changelog, licenses)
	hasChangelog := !model.ChangelogPath.IsNull() && !model.ChangelogPath.IsUnknown()
	hasLicenses := !model.Licenses.IsNull() && !model.Licenses.IsUnknown()

	if hasChangelog || hasLicenses {
		b.WriteString("\nintegrations:\n")

		if hasChangelog {
			b.WriteString("  changelog:\n")
			b.WriteString(fmt.Sprintf("    path: %s\n", model.ChangelogPath.ValueString()))
		}

		if hasLicenses {
			var licenses []struct {
				Title     string `json:"title"`
				Vendor    string `json:"vendor,omitempty"`
				Purchased string `json:"purchased,omitempty"`
				Expires   string `json:"expires,omitempty"`
				Seats     int    `json:"seats,omitempty"`
				Cost      string `json:"cost,omitempty"`
				Contract  string `json:"contract,omitempty"`
				Notes     string `json:"notes,omitempty"`
			}
			if err := json.Unmarshal([]byte(model.Licenses.ValueString()), &licenses); err == nil && len(licenses) > 0 {
				b.WriteString("  licenses:\n")
				for _, lic := range licenses {
					b.WriteString(fmt.Sprintf("    - title: %s\n", lic.Title))
					if lic.Vendor != "" {
						b.WriteString(fmt.Sprintf("      vendor: %s\n", lic.Vendor))
					}
					if lic.Purchased != "" {
						b.WriteString(fmt.Sprintf("      purchased: %s\n", lic.Purchased))
					}
					if lic.Expires != "" {
						b.WriteString(fmt.Sprintf("      expires: %s\n", lic.Expires))
					}
					if lic.Seats > 0 {
						b.WriteString(fmt.Sprintf("      seats: %d\n", lic.Seats))
					}
					if lic.Cost != "" {
						b.WriteString(fmt.Sprintf("      cost: %s\n", lic.Cost))
					}
					if lic.Contract != "" {
						b.WriteString(fmt.Sprintf("      contract: %s\n", lic.Contract))
					}
					if lic.Notes != "" {
						b.WriteString(fmt.Sprintf("      notes: %s\n", lic.Notes))
					}
				}
			}
		}
	}

	// Build interfaces section (http, grpc)
	if !model.Interfaces.IsNull() && !model.Interfaces.IsUnknown() {
		var ifaces map[string]interface{}
		if err := json.Unmarshal([]byte(model.Interfaces.ValueString()), &ifaces); err == nil && len(ifaces) > 0 {
			b.WriteString("\ninterfaces:\n")
			// Handle http interface
			if httpIface, ok := ifaces["http"].(map[string]interface{}); ok {
				b.WriteString("  http:\n")
				if v, ok := httpIface["baseUrl"].(string); ok && v != "" {
					b.WriteString(fmt.Sprintf("    baseUrl: %s\n", v))
				}
				if v, ok := httpIface["openapi"].(string); ok && v != "" {
					b.WriteString(fmt.Sprintf("    openapi: %s\n", v))
				}
				if auth, ok := httpIface["auth"].(map[string]interface{}); ok {
					b.WriteString("    auth:\n")
					if v, ok := auth["type"].(string); ok && v != "" {
						b.WriteString(fmt.Sprintf("      type: %s\n", v))
					}
				}
				if graphql, ok := httpIface["graphql"].(map[string]interface{}); ok {
					b.WriteString("    graphql:\n")
					if v, ok := graphql["endpoint"].(string); ok && v != "" {
						b.WriteString(fmt.Sprintf("      endpoint: %s\n", v))
					}
					if v, ok := graphql["schema"].(string); ok && v != "" {
						b.WriteString(fmt.Sprintf("      schema: %s\n", v))
					}
				}
			}
			// Handle grpc interface
			if grpcIface, ok := ifaces["grpc"].(map[string]interface{}); ok {
				b.WriteString("  grpc:\n")
				if v, ok := grpcIface["package"].(string); ok && v != "" {
					b.WriteString(fmt.Sprintf("    package: %s\n", v))
				}
				if v, ok := grpcIface["proto"].(string); ok && v != "" {
					b.WriteString(fmt.Sprintf("    proto: %s\n", v))
				}
			}
		}
	}

	return b.String()
}

// relationsEquivalent checks if two relations JSON strings contain the same set of relations
// regardless of ordering.
func relationsEquivalent(a, b string) bool {
	type rel struct {
		Type   string `json:"type"`
		Target string `json:"target"`
	}
	var relsA, relsB []rel
	if err := json.Unmarshal([]byte(a), &relsA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &relsB); err != nil {
		return false
	}
	if len(relsA) != len(relsB) {
		return false
	}
	setA := make(map[string]bool, len(relsA))
	for _, r := range relsA {
		setA[r.Type+"|"+r.Target] = true
	}
	for _, r := range relsB {
		if !setA[r.Type+"|"+r.Target] {
			return false
		}
	}
	return true
}

// linksEquivalent checks if two links JSON strings contain the same set of links
// regardless of ordering or whitespace differences.
func linksEquivalent(a, b string) bool {
	type link struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Icon string `json:"icon,omitempty"`
	}
	var linksA, linksB []link
	if err := json.Unmarshal([]byte(a), &linksA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &linksB); err != nil {
		return false
	}
	if len(linksA) != len(linksB) {
		return false
	}
	setA := make(map[string]bool, len(linksA))
	for _, l := range linksA {
		setA[l.Name+"|"+l.URL+"|"+l.Icon] = true
	}
	for _, l := range linksB {
		if !setA[l.Name+"|"+l.URL+"|"+l.Icon] {
			return false
		}
	}
	return true
}

// licensesEquivalent checks if two licenses JSON strings contain the same set of licenses
// regardless of ordering.
func licensesEquivalent(a, b string) bool {
	type lic struct {
		Title     string `json:"title"`
		Vendor    string `json:"vendor,omitempty"`
		Purchased string `json:"purchased,omitempty"`
		Expires   string `json:"expires,omitempty"`
		Seats     int    `json:"seats,omitempty"`
		Cost      string `json:"cost,omitempty"`
		Contract  string `json:"contract,omitempty"`
		Notes     string `json:"notes,omitempty"`
	}
	var licsA, licsB []lic
	if err := json.Unmarshal([]byte(a), &licsA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &licsB); err != nil {
		return false
	}
	if len(licsA) != len(licsB) {
		return false
	}
	setA := make(map[string]bool, len(licsA))
	for _, l := range licsA {
		key := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s|%s", l.Title, l.Vendor, l.Purchased, l.Expires, l.Seats, l.Cost, l.Contract, l.Notes)
		setA[key] = true
	}
	for _, l := range licsB {
		key := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s|%s", l.Title, l.Vendor, l.Purchased, l.Expires, l.Seats, l.Cost, l.Contract, l.Notes)
		if !setA[key] {
			return false
		}
	}
	return true
}

// interfacesEquivalent checks if two interfaces JSON strings are semantically equivalent.
func interfacesEquivalent(a, b string) bool {
	var ifacesA, ifacesB map[string]interface{}
	if err := json.Unmarshal([]byte(a), &ifacesA); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &ifacesB); err != nil {
		return false
	}
	// Re-marshal to canonical JSON for comparison
	canonA, errA := json.Marshal(ifacesA)
	canonB, errB := json.Marshal(ifacesB)
	if errA != nil || errB != nil {
		return false
	}
	return string(canonA) == string(canonB)
}

// mapEntityToState maps a client.Entity to the resource state model.
func mapEntityToState(ctx context.Context, entity *client.Entity, state *EntityResourceModel) {
	state.ID = types.StringValue(entity.Service.ID)
	state.Name = types.StringValue(entity.Service.Name)
	state.Type = types.StringValue(entity.Service.Type)

	if entity.Service.Tier != "" {
		state.Tier = types.StringValue(entity.Service.Tier)
	}

	if entity.Description != "" {
		state.Description = types.StringValue(entity.Description)
	}

	if entity.Lifecycle != "" {
		state.Lifecycle = types.StringValue(entity.Lifecycle)
	}

	if len(entity.Owner) > 0 {
		state.Owner = types.StringValue(entity.Owner[0].ID)
	}

	// Map tags from API response back to state
	if len(entity.Tags) > 0 {
		tagValues := make([]types.String, len(entity.Tags))
		for i, tag := range entity.Tags {
			tagValues[i] = types.StringValue(tag)
		}
		state.Tags, _ = types.SetValueFrom(ctx, types.StringType, tagValues)
	}

	// Map links from API response back to state as JSON string
	if len(entity.Links) > 0 {
		linksJSON, err := json.Marshal(entity.Links)
		if err == nil {
			state.Links = types.StringValue(string(linksJSON))
		}
	}

	// Map relations from API response back to state as JSON string
	// API returns {type, targetType, targetId} but terraform state uses {type, target: "type:id"}
	// Sort by type+target for deterministic ordering
	if len(entity.Relations) > 0 {
		type tfRelation struct {
			Type   string `json:"type"`
			Target string `json:"target"`
		}
		tfRelations := make([]tfRelation, len(entity.Relations))
		for i, rel := range entity.Relations {
			tfRelations[i] = tfRelation{
				Type:   rel.Type,
				Target: rel.TargetType + ":" + rel.TargetID,
			}
		}
		sort.Slice(tfRelations, func(i, j int) bool {
			if tfRelations[i].Type != tfRelations[j].Type {
				return tfRelations[i].Type < tfRelations[j].Type
			}
			return tfRelations[i].Target < tfRelations[j].Target
		})
		relationsJSON, err := json.Marshal(tfRelations)
		if err == nil {
			state.Relations = types.StringValue(string(relationsJSON))
		}
	}

	// Map repository path (computed, read-only)
	if entity.RepositoryPath != "" {
		state.RepositoryPath = types.StringValue(entity.RepositoryPath)
	}

	// Map interfaces from API response
	if entity.Interfaces != nil && len(entity.Interfaces) > 0 {
		ifacesJSON, err := json.Marshal(entity.Interfaces)
		if err == nil {
			state.Interfaces = types.StringValue(string(ifacesJSON))
		}
	}

	// Map integrations (changelog, licenses)
	if entity.Integrations != nil {
		if entity.Integrations.Changelog != nil && entity.Integrations.Changelog.Path != "" {
			state.ChangelogPath = types.StringValue(entity.Integrations.Changelog.Path)
		}

		if len(entity.Integrations.Licenses) > 0 {
			licensesJSON, err := json.Marshal(entity.Integrations.Licenses)
			if err == nil {
				state.Licenses = types.StringValue(string(licensesJSON))
			}
		}
	}

	if entity.CreatedAt != "" {
		state.CreatedAt = types.StringValue(entity.CreatedAt)
	}

	if entity.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(entity.UpdatedAt)
	}
}
