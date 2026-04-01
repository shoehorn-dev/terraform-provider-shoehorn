package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestGovernanceActionResource_Metadata(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_governance_action" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_governance_action")
	}
}

func TestGovernanceActionResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{
		"id", "entity_id", "entity_name", "title", "description",
		"priority", "status", "source_type", "source_id", "assigned_to",
		"sla_days", "due_date", "resolution_note", "created_at", "updated_at",
	}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestGovernanceActionResource_Schema_EntityIDIsRequired(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["entity_id"]
	if attr == nil {
		t.Fatal("entity_id attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("entity_id should be required")
	}
}

func TestGovernanceActionResource_Schema_TitleIsRequired(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["title"]
	if attr == nil {
		t.Fatal("title attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("title should be required")
	}
}

func TestGovernanceActionResource_Schema_PriorityIsRequired(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["priority"]
	if attr == nil {
		t.Fatal("priority attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("priority should be required")
	}
}

func TestGovernanceActionResource_Schema_SourceTypeIsRequired(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["source_type"]
	if attr == nil {
		t.Fatal("source_type attribute not found")
	}
	if !attr.IsRequired() {
		t.Error("source_type should be required")
	}
}

func TestGovernanceActionResource_Schema_IDIsComputed(t *testing.T) {
	r := NewGovernanceActionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["id"]
	if attr == nil {
		t.Fatal("id attribute not found")
	}
	if !attr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestGovernanceActionResource_Configure_WithValidClient(t *testing.T) {
	r := &GovernanceActionResource{}
	c := client.NewClient("https://test.example.com", "key", 30*time.Second)

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: c,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}
	if r.client != c {
		t.Error("client not set correctly")
	}
}

func TestGovernanceActionResource_Configure_WrongType(t *testing.T) {
	r := &GovernanceActionResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestMapGovernanceActionToState(t *testing.T) {
	slaDays := 14
	action := &client.GovernanceAction{
		ID:             "ga-123",
		EntityID:       "entity-1",
		EntityName:     "My Service",
		Title:          "Fix vulnerability",
		Description:    "CVE-2025-001 needs patching",
		Priority:       "high",
		Status:         "open",
		SourceType:     "security",
		SourceID:       "scan-42",
		AssignedTo:     "team-alpha",
		SLADays:        &slaDays,
		DueDate:        "2025-02-01",
		ResolutionNote: "",
		CreatedAt:      "2025-01-15T10:00:00Z",
		UpdatedAt:      "2025-01-15T11:00:00Z",
	}

	state := &GovernanceActionResourceModel{}
	mapGovernanceActionToState(action, state)

	if state.ID.ValueString() != "ga-123" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "ga-123")
	}
	if state.EntityID.ValueString() != "entity-1" {
		t.Errorf("EntityID = %q, want %q", state.EntityID.ValueString(), "entity-1")
	}
	if state.EntityName.ValueString() != "My Service" {
		t.Errorf("EntityName = %q, want %q", state.EntityName.ValueString(), "My Service")
	}
	if state.Title.ValueString() != "Fix vulnerability" {
		t.Errorf("Title = %q, want %q", state.Title.ValueString(), "Fix vulnerability")
	}
	if state.Description.ValueString() != "CVE-2025-001 needs patching" {
		t.Errorf("Description = %q, want %q", state.Description.ValueString(), "CVE-2025-001 needs patching")
	}
	if state.Priority.ValueString() != "high" {
		t.Errorf("Priority = %q, want %q", state.Priority.ValueString(), "high")
	}
	if state.Status.ValueString() != "open" {
		t.Errorf("Status = %q, want %q", state.Status.ValueString(), "open")
	}
	if state.SourceType.ValueString() != "security" {
		t.Errorf("SourceType = %q, want %q", state.SourceType.ValueString(), "security")
	}
	if state.SourceID.ValueString() != "scan-42" {
		t.Errorf("SourceID = %q, want %q", state.SourceID.ValueString(), "scan-42")
	}
	if state.AssignedTo.ValueString() != "team-alpha" {
		t.Errorf("AssignedTo = %q, want %q", state.AssignedTo.ValueString(), "team-alpha")
	}
	if state.SLADays.ValueInt64() != 14 {
		t.Errorf("SLADays = %d, want %d", state.SLADays.ValueInt64(), 14)
	}
	if state.DueDate.ValueString() != "2025-02-01" {
		t.Errorf("DueDate = %q, want %q", state.DueDate.ValueString(), "2025-02-01")
	}
}

func TestMapGovernanceActionToState_EmptyOptionalFields(t *testing.T) {
	action := &client.GovernanceAction{
		ID:         "ga-456",
		EntityID:   "entity-2",
		Title:      "Review access",
		Priority:   "low",
		SourceType: "policy",
	}

	state := &GovernanceActionResourceModel{}
	mapGovernanceActionToState(action, state)

	if state.ID.ValueString() != "ga-456" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "ga-456")
	}
	if !state.EntityName.IsNull() {
		t.Errorf("EntityName should be null when empty, got %q", state.EntityName.ValueString())
	}
	if !state.Description.IsNull() {
		t.Errorf("Description should be null when empty, got %q", state.Description.ValueString())
	}
	if !state.Status.IsNull() {
		t.Errorf("Status should be null when empty, got %q", state.Status.ValueString())
	}
	if !state.AssignedTo.IsNull() {
		t.Errorf("AssignedTo should be null when empty, got %q", state.AssignedTo.ValueString())
	}
	if !state.SLADays.IsNull() {
		t.Errorf("SLADays should be null when nil, got %d", state.SLADays.ValueInt64())
	}
}

func TestMapGovernanceActionToState_PreservesUserSetValues(t *testing.T) {
	// User-set Optional fields should be preserved when API returns empty
	state := &GovernanceActionResourceModel{
		AssignedTo:     types.StringValue("old-team"),
		ResolutionNote: types.StringValue("old note"),
		SLADays:        types.Int64Value(7),
	}

	action := &client.GovernanceAction{
		ID:         "ga-789",
		EntityID:   "entity-3",
		Title:      "Audit logs",
		Priority:   "medium",
		SourceType: "scorecard",
	}

	mapGovernanceActionToState(action, state)

	if state.AssignedTo.ValueString() != "old-team" {
		t.Errorf("AssignedTo should be preserved, got %q", state.AssignedTo.ValueString())
	}
	if state.ResolutionNote.ValueString() != "old note" {
		t.Errorf("ResolutionNote should be preserved, got %q", state.ResolutionNote.ValueString())
	}
	if !state.SLADays.IsNull() {
		t.Errorf("SLADays should be null when API returns nil, got %d", state.SLADays.ValueInt64())
	}
}
