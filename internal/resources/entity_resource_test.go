package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestEntityResource_Metadata(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_entity" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_entity")
	}
}

func TestEntityResource_Schema_HasRequiredAttributes(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "name", "type", "description", "entity_lifecycle", "tier", "owner", "tags", "links", "relations", "licenses", "changelog_path", "interfaces", "repository_path", "created_at", "updated_at"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestEntityResource_Schema_NameIsRequired(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	nameAttr := resp.Schema.Attributes["name"]
	if nameAttr == nil {
		t.Fatal("name attribute not found")
	}
	if !nameAttr.IsRequired() {
		t.Error("name should be required")
	}
}

func TestEntityResource_Schema_TypeIsRequired(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	typeAttr := resp.Schema.Attributes["type"]
	if typeAttr == nil {
		t.Fatal("type attribute not found")
	}
	if !typeAttr.IsRequired() {
		t.Error("type should be required")
	}
}

func TestEntityResource_Schema_IDIsComputed(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	idAttr := resp.Schema.Attributes["id"]
	if idAttr == nil {
		t.Fatal("id attribute not found")
	}
	if !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
}

func TestEntityResource_Schema_DescriptionIsOptional(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	descAttr := resp.Schema.Attributes["description"]
	if descAttr == nil {
		t.Fatal("description attribute not found")
	}
	if !descAttr.IsOptional() {
		t.Error("description should be optional")
	}
}

func TestEntityResource_Configure_WithValidClient(t *testing.T) {
	r := &EntityResource{}
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

func TestEntityResource_Configure_NilProviderData(t *testing.T) {
	r := &EntityResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: nil,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Error("should not error on nil provider data")
	}
}

func TestEntityResource_Configure_WrongType(t *testing.T) {
	r := &EntityResource{}

	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestBuildManifestYAML_MinimalFields(t *testing.T) {
	model := &EntityResourceModel{
		Name: types.StringValue("my-service"),
		Type: types.StringValue("service"),
	}

	yaml := buildManifestYAML(model)

	if !strings.Contains(yaml, "schemaVersion: 1") {
		t.Error("missing schemaVersion")
	}
	if !strings.Contains(yaml, "name: my-service") {
		t.Error("missing name")
	}
	if !strings.Contains(yaml, "type: service") {
		t.Error("missing type")
	}
}

func TestBuildManifestYAML_AllFields(t *testing.T) {
	tags, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"go", "api"})

	model := &EntityResourceModel{
		Name:        types.StringValue("my-service"),
		Type:        types.StringValue("service"),
		Description: types.StringValue("A test service"),
		Lifecycle:   types.StringValue("production"),
		Tier:        types.StringValue("tier1"),
		Owner:       types.StringValue("platform"),
		Tags:        tags,
	}

	yaml := buildManifestYAML(model)

	expectedParts := []string{
		"schemaVersion: 1",
		"name: my-service",
		"type: service",
		"description: A test service",
		"lifecycle: production",
		"tier: tier1",
		"- type: team",
		"id: platform",
		"- go",
		"- api",
	}

	for _, part := range expectedParts {
		if !strings.Contains(yaml, part) {
			t.Errorf("manifest YAML missing %q\nGot:\n%s", part, yaml)
		}
	}
}

func TestMapEntityToState(t *testing.T) {
	entity := &client.Entity{
		Service: client.EntityService{
			ID:   "my-service",
			Name: "My Service",
			Type: "service",
			Tier: "tier1",
		},
		Description: "A test service",
		Lifecycle:   "production",
		Owner: []client.OwnerInfo{
			{Type: "team", ID: "platform"},
		},
		Tags: []string{"go", "api"},
		Links: []client.LinkInfo{
			{Name: "Repo", URL: "https://github.com/example/repo", Icon: "github"},
		},
		CreatedAt: "2025-01-15T10:30:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &EntityResourceModel{}
	mapEntityToState(context.Background(), entity, state)

	if state.ID.ValueString() != "my-service" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "my-service")
	}
	if state.Name.ValueString() != "My Service" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "My Service")
	}
	if state.Type.ValueString() != "service" {
		t.Errorf("Type = %q, want %q", state.Type.ValueString(), "service")
	}
	if state.Tier.ValueString() != "tier1" {
		t.Errorf("Tier = %q, want %q", state.Tier.ValueString(), "tier1")
	}
	if state.Description.ValueString() != "A test service" {
		t.Errorf("Description = %q, want %q", state.Description.ValueString(), "A test service")
	}
	if state.Lifecycle.ValueString() != "production" {
		t.Errorf("Lifecycle = %q, want %q", state.Lifecycle.ValueString(), "production")
	}
	if state.Owner.ValueString() != "platform" {
		t.Errorf("Owner = %q, want %q", state.Owner.ValueString(), "platform")
	}
	if state.CreatedAt.ValueString() != "2025-01-15T10:30:00Z" {
		t.Errorf("CreatedAt = %q, want %q", state.CreatedAt.ValueString(), "2025-01-15T10:30:00Z")
	}
	if state.UpdatedAt.ValueString() != "2025-01-15T11:00:00Z" {
		t.Errorf("UpdatedAt = %q, want %q", state.UpdatedAt.ValueString(), "2025-01-15T11:00:00Z")
	}
}

func TestMapEntityToState_EmptyOptionalFields(t *testing.T) {
	entity := &client.Entity{
		Service: client.EntityService{
			ID:   "my-service",
			Name: "My Service",
			Type: "service",
		},
	}

	state := &EntityResourceModel{}
	mapEntityToState(context.Background(), entity, state)

	if state.ID.ValueString() != "my-service" {
		t.Errorf("ID = %q, want %q", state.ID.ValueString(), "my-service")
	}
	if state.Name.ValueString() != "My Service" {
		t.Errorf("Name = %q, want %q", state.Name.ValueString(), "My Service")
	}
	// Optional fields should not crash
	if state.Tier.IsNull() {
		// Expected - tier not set
	}
	if state.Owner.IsNull() {
		// Expected - owner not set
	}
}

func TestEntityResource_Schema_HasNewAttributes(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	newAttrs := []string{"licenses", "changelog_path", "repository_path"}
	for _, name := range newAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}
}

func TestEntityResource_Schema_RepositoryPathIsComputed(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	rpAttr := resp.Schema.Attributes["repository_path"]
	if rpAttr == nil {
		t.Fatal("repository_path attribute not found")
	}
	if !rpAttr.IsComputed() {
		t.Error("repository_path should be computed")
	}
}

func TestEntityResource_Schema_LicensesIsOptional(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	licAttr := resp.Schema.Attributes["licenses"]
	if licAttr == nil {
		t.Fatal("licenses attribute not found")
	}
	if !licAttr.IsOptional() {
		t.Error("licenses should be optional")
	}
}

func TestEntityResource_Schema_ChangelogPathIsOptional(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	clAttr := resp.Schema.Attributes["changelog_path"]
	if clAttr == nil {
		t.Fatal("changelog_path attribute not found")
	}
	if !clAttr.IsOptional() {
		t.Error("changelog_path should be optional")
	}
}

func TestBuildManifestYAML_WithIntegrations(t *testing.T) {
	model := &EntityResourceModel{
		Name:          types.StringValue("my-service"),
		Type:          types.StringValue("service"),
		ChangelogPath: types.StringValue("CHANGELOG.md"),
		Licenses:      types.StringValue(`[{"title":"MSSQL Enterprise","vendor":"Microsoft","seats":10}]`),
	}

	yaml := buildManifestYAML(model)

	expectedParts := []string{
		"integrations:",
		"  changelog:",
		"    path: CHANGELOG.md",
		"  licenses:",
		"    - title: MSSQL Enterprise",
		"      vendor: Microsoft",
		"      seats: 10",
	}

	for _, part := range expectedParts {
		if !strings.Contains(yaml, part) {
			t.Errorf("manifest YAML missing %q\nGot:\n%s", part, yaml)
		}
	}
}

func TestBuildManifestYAML_WithChangelogOnly(t *testing.T) {
	model := &EntityResourceModel{
		Name:          types.StringValue("my-service"),
		Type:          types.StringValue("service"),
		ChangelogPath: types.StringValue("docs/CHANGELOG.md"),
	}

	yaml := buildManifestYAML(model)

	if !strings.Contains(yaml, "integrations:") {
		t.Error("missing integrations section")
	}
	if !strings.Contains(yaml, "path: docs/CHANGELOG.md") {
		t.Errorf("missing changelog path\nGot:\n%s", yaml)
	}
	if strings.Contains(yaml, "licenses:") {
		t.Error("should not contain licenses section when not set")
	}
}

func TestBuildManifestYAML_WithLicensesAllFields(t *testing.T) {
	model := &EntityResourceModel{
		Name: types.StringValue("my-service"),
		Type: types.StringValue("service"),
		Licenses: types.StringValue(`[{"title":"Enterprise DB","vendor":"Oracle","purchased":"2025-01-01","expires":"2025-12-31","seats":50,"cost":"$10000/year","contract":"CON-123","notes":"Annual renewal"}]`),
	}

	yaml := buildManifestYAML(model)

	expectedParts := []string{
		"    - title: Enterprise DB",
		"      vendor: Oracle",
		"      purchased: 2025-01-01",
		"      expires: 2025-12-31",
		"      seats: 50",
		"      cost: $10000/year",
		"      contract: CON-123",
		"      notes: Annual renewal",
	}

	for _, part := range expectedParts {
		if !strings.Contains(yaml, part) {
			t.Errorf("manifest YAML missing %q\nGot:\n%s", part, yaml)
		}
	}
}

func TestBuildManifestYAML_NoIntegrationsWhenNotSet(t *testing.T) {
	model := &EntityResourceModel{
		Name: types.StringValue("my-service"),
		Type: types.StringValue("service"),
	}

	yaml := buildManifestYAML(model)

	if strings.Contains(yaml, "integrations:") {
		t.Errorf("should not contain integrations section when neither changelog nor licenses set\nGot:\n%s", yaml)
	}
}

func TestMapEntityToState_WithIntegrationsAndRepositoryPath(t *testing.T) {
	entity := &client.Entity{
		Service: client.EntityService{
			ID:   "my-service",
			Name: "My Service",
			Type: "service",
		},
		RepositoryPath: "github:shoehorn-dev/my-service",
		Integrations: &client.Integrations{
			Changelog: &client.ChangelogIntegration{
				Path: "CHANGELOG.md",
			},
			Licenses: []client.LicenseInfo{
				{Title: "MSSQL Enterprise", Vendor: "Microsoft", Seats: 10},
			},
		},
		CreatedAt: "2025-01-15T10:30:00Z",
		UpdatedAt: "2025-01-15T11:00:00Z",
	}

	state := &EntityResourceModel{}
	mapEntityToState(context.Background(), entity, state)

	if state.RepositoryPath.ValueString() != "github:shoehorn-dev/my-service" {
		t.Errorf("RepositoryPath = %q, want %q", state.RepositoryPath.ValueString(), "github:shoehorn-dev/my-service")
	}
	if state.ChangelogPath.ValueString() != "CHANGELOG.md" {
		t.Errorf("ChangelogPath = %q, want %q", state.ChangelogPath.ValueString(), "CHANGELOG.md")
	}
	if state.Licenses.IsNull() {
		t.Fatal("Licenses should not be null")
	}

	var licenses []client.LicenseInfo
	if err := json.Unmarshal([]byte(state.Licenses.ValueString()), &licenses); err != nil {
		t.Fatalf("Failed to unmarshal licenses: %v", err)
	}
	if len(licenses) != 1 {
		t.Fatalf("Expected 1 license, got %d", len(licenses))
	}
	if licenses[0].Title != "MSSQL Enterprise" {
		t.Errorf("License title = %q, want %q", licenses[0].Title, "MSSQL Enterprise")
	}
}

func TestMapEntityToState_NilIntegrations(t *testing.T) {
	entity := &client.Entity{
		Service: client.EntityService{
			ID:   "my-service",
			Name: "My Service",
			Type: "service",
		},
	}

	state := &EntityResourceModel{}
	mapEntityToState(context.Background(), entity, state)

	if !state.RepositoryPath.IsNull() {
		t.Error("RepositoryPath should be null when not set")
	}
	if !state.ChangelogPath.IsNull() {
		t.Error("ChangelogPath should be null when not set")
	}
	if !state.Licenses.IsNull() {
		t.Error("Licenses should be null when not set")
	}
}

func TestLicensesEquivalent(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "same order",
			a:    `[{"title":"A","vendor":"V1"},{"title":"B","vendor":"V2"}]`,
			b:    `[{"title":"A","vendor":"V1"},{"title":"B","vendor":"V2"}]`,
			want: true,
		},
		{
			name: "different order",
			a:    `[{"title":"B","vendor":"V2"},{"title":"A","vendor":"V1"}]`,
			b:    `[{"title":"A","vendor":"V1"},{"title":"B","vendor":"V2"}]`,
			want: true,
		},
		{
			name: "different content",
			a:    `[{"title":"A","vendor":"V1"}]`,
			b:    `[{"title":"A","vendor":"V2"}]`,
			want: false,
		},
		{
			name: "different lengths",
			a:    `[{"title":"A"}]`,
			b:    `[{"title":"A"},{"title":"B"}]`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := licensesEquivalent(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("licensesEquivalent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntityResource_Schema_InterfacesIsOptional(t *testing.T) {
	r := NewEntityResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr := resp.Schema.Attributes["interfaces"]
	if attr == nil {
		t.Fatal("interfaces attribute not found")
	}
	if attr.IsRequired() {
		t.Error("interfaces should be optional, not required")
	}
	if attr.IsComputed() {
		t.Error("interfaces should not be computed")
	}
}

func TestBuildManifestYAML_WithInterfaces(t *testing.T) {
	model := &EntityResourceModel{
		Name: types.StringValue("api-gateway"),
		Type: types.StringValue("api"),
		Interfaces: types.StringValue(`{"http":{"openapi":"https://petstore3.swagger.io/api/v3/openapi.json"}}`),
	}

	yaml := buildManifestYAML(model)

	if !strings.Contains(yaml, "interfaces:") {
		t.Error("manifest should contain interfaces section")
	}
	if !strings.Contains(yaml, "http:") {
		t.Error("manifest should contain http interface")
	}
	if !strings.Contains(yaml, "openapi: https://petstore3.swagger.io/api/v3/openapi.json") {
		t.Error("manifest should contain openapi URL")
	}
}

func TestBuildManifestYAML_WithInterfacesFullHTTP(t *testing.T) {
	model := &EntityResourceModel{
		Name: types.StringValue("my-api"),
		Type: types.StringValue("api"),
		Interfaces: types.StringValue(`{"http":{"baseUrl":"https://api.example.com","openapi":"openapi.yaml","auth":{"type":"oauth2"},"graphql":{"endpoint":"/graphql","schema":"schema.graphql"}}}`),
	}

	yaml := buildManifestYAML(model)

	if !strings.Contains(yaml, "baseUrl: https://api.example.com") {
		t.Error("manifest should contain baseUrl")
	}
	if !strings.Contains(yaml, "openapi: openapi.yaml") {
		t.Error("manifest should contain openapi")
	}
	if !strings.Contains(yaml, "auth:") {
		t.Error("manifest should contain auth section")
	}
	if !strings.Contains(yaml, "type: oauth2") {
		t.Error("manifest should contain auth type")
	}
	if !strings.Contains(yaml, "graphql:") {
		t.Error("manifest should contain graphql section")
	}
	if !strings.Contains(yaml, "endpoint: /graphql") {
		t.Error("manifest should contain graphql endpoint")
	}
	if !strings.Contains(yaml, "schema: schema.graphql") {
		t.Error("manifest should contain graphql schema")
	}
}

func TestBuildManifestYAML_WithInterfacesGRPC(t *testing.T) {
	model := &EntityResourceModel{
		Name:       types.StringValue("grpc-service"),
		Type:       types.StringValue("service"),
		Interfaces: types.StringValue(`{"grpc":{"package":"com.example.api","proto":"api.proto"}}`),
	}

	yaml := buildManifestYAML(model)

	if !strings.Contains(yaml, "grpc:") {
		t.Error("manifest should contain grpc section")
	}
	if !strings.Contains(yaml, "package: com.example.api") {
		t.Error("manifest should contain grpc package")
	}
	if !strings.Contains(yaml, "proto: api.proto") {
		t.Error("manifest should contain grpc proto")
	}
}

func TestBuildManifestYAML_NoInterfacesWhenNotSet(t *testing.T) {
	model := &EntityResourceModel{
		Name: types.StringValue("my-service"),
		Type: types.StringValue("service"),
	}

	yaml := buildManifestYAML(model)

	if strings.Contains(yaml, "interfaces:") {
		t.Error("manifest should NOT contain interfaces section when not set")
	}
}

func TestMapEntityToState_WithInterfaces(t *testing.T) {
	entity := &client.Entity{
		Service: client.EntityService{
			ID:   "api-gateway",
			Name: "api-gateway",
			Type: "api",
			Tier: "critical",
		},
		Interfaces: map[string]interface{}{
			"http": map[string]interface{}{
				"openapi": "https://petstore3.swagger.io/api/v3/openapi.json",
			},
		},
	}

	state := &EntityResourceModel{}
	mapEntityToState(context.Background(), entity, state)

	if state.Interfaces.IsNull() {
		t.Fatal("interfaces should not be null")
	}

	var ifaces map[string]interface{}
	if err := json.Unmarshal([]byte(state.Interfaces.ValueString()), &ifaces); err != nil {
		t.Fatalf("failed to parse interfaces JSON: %v", err)
	}

	httpIface, ok := ifaces["http"].(map[string]interface{})
	if !ok {
		t.Fatal("expected http interface")
	}
	if httpIface["openapi"] != "https://petstore3.swagger.io/api/v3/openapi.json" {
		t.Errorf("unexpected openapi value: %v", httpIface["openapi"])
	}
}

func TestMapEntityToState_EmptyInterfaces(t *testing.T) {
	entity := &client.Entity{
		Service: client.EntityService{
			ID:   "my-service",
			Name: "my-service",
			Type: "service",
		},
		Interfaces: map[string]interface{}{},
	}

	state := &EntityResourceModel{}
	mapEntityToState(context.Background(), entity, state)

	if !state.Interfaces.IsNull() {
		t.Error("interfaces should be null when empty map")
	}
}

func TestInterfacesEquivalent(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "identical",
			a:    `{"http":{"openapi":"spec.json"}}`,
			b:    `{"http":{"openapi":"spec.json"}}`,
			want: true,
		},
		{
			name: "different values",
			a:    `{"http":{"openapi":"spec.json"}}`,
			b:    `{"http":{"openapi":"other.json"}}`,
			want: false,
		},
		{
			name: "different keys",
			a:    `{"http":{"openapi":"spec.json"}}`,
			b:    `{"grpc":{"package":"com.example"}}`,
			want: false,
		},
		{
			name: "key order different but equivalent",
			a:    `{"http":{"baseUrl":"https://example.com","openapi":"spec.json"}}`,
			b:    `{"http":{"openapi":"spec.json","baseUrl":"https://example.com"}}`,
			want: true,
		},
		{
			name: "invalid json a",
			a:    `not json`,
			b:    `{"http":{}}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interfacesEquivalent(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("interfacesEquivalent(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestEntityClient_CRUD_Integration tests the full CRUD lifecycle using a mock server.
func TestEntityClient_CRUD_Integration(t *testing.T) {
	entities := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/manifests/entities":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			entities["test-entity"] = map[string]interface{}{
				"service": map[string]interface{}{
					"id":   "test-entity",
					"name": "test-entity",
					"type": "service",
					"tier": "tier1",
				},
				"description": "A test entity",
				"lifecycle":   "experimental",
				"owner": []map[string]interface{}{
					{"type": "team", "id": "platform"},
				},
				"tags":      []string{"go"},
				"createdAt": "2025-01-15T10:30:00Z",
				"updatedAt": "2025-01-15T10:30:00Z",
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"entity": map[string]interface{}{
					"id":          1,
					"serviceId":   "test-entity",
					"name":        "test-entity",
					"type":        "service",
					"lifecycle":   "experimental",
					"description": "A test entity",
					"source":      "terraform",
					"createdAt":   "2025-01-15T10:30:00Z",
					"updatedAt":   "2025-01-15T10:30:00Z",
				},
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/entities/test-entity":
			if entity, ok := entities["test-entity"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"entity": entity})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "Entity not found"})
			}

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/manifests/entities/test-entity":
			entity := entities["test-entity"]
			svc := entity["service"].(map[string]interface{})
			svc["name"] = "test-entity"
			entity["service"] = svc
			entity["description"] = "Updated description"
			entity["lifecycle"] = "production"
			entity["updatedAt"] = "2025-01-15T12:00:00Z"
			entities["test-entity"] = entity

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"entity": map[string]interface{}{
					"id":          1,
					"serviceId":   "test-entity",
					"name":        "test-entity",
					"type":        "service",
					"lifecycle":   "production",
					"description": "Updated description",
					"source":      "terraform",
					"createdAt":   "2025-01-15T10:30:00Z",
					"updatedAt":   "2025-01-15T12:00:00Z",
				},
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/manifests/entities/test-entity":
			delete(entities, "test-entity")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"Not found: %s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	createResp, err := c.CreateEntity(context.Background(), client.CreateEntityRequest{
		Content: "apiVersion: shoehorn.io/v1alpha1\nkind: Entity\nmetadata:\n  name: test-entity\nspec:\n  type: service",
		Source:  "terraform",
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if !createResp.Success {
		t.Error("CREATE: Success = false, want true")
	}
	if createResp.Entity.ServiceID != "test-entity" {
		t.Errorf("CREATE: ServiceID = %q, want %q", createResp.Entity.ServiceID, "test-entity")
	}

	// READ
	entity, err := c.GetEntity(context.Background(), "test-entity")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if entity.Service.ID != "test-entity" {
		t.Errorf("READ: Service.ID = %q, want %q", entity.Service.ID, "test-entity")
	}
	if entity.Description != "A test entity" {
		t.Errorf("READ: Description = %q, want %q", entity.Description, "A test entity")
	}
	if entity.Lifecycle != "experimental" {
		t.Errorf("READ: Lifecycle = %q, want %q", entity.Lifecycle, "experimental")
	}

	// UPDATE
	updateResp, err := c.UpdateEntity(context.Background(), "test-entity", client.CreateEntityRequest{
		Content: "apiVersion: shoehorn.io/v1alpha1\nkind: Entity\nmetadata:\n  name: test-entity\nspec:\n  type: service\n  lifecycle: production",
		Source:  "terraform",
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updateResp.Entity.Lifecycle != "production" {
		t.Errorf("UPDATE: Lifecycle = %q, want %q", updateResp.Entity.Lifecycle, "production")
	}
	if updateResp.Entity.Description != "Updated description" {
		t.Errorf("UPDATE: Description = %q, want %q", updateResp.Entity.Description, "Updated description")
	}

	// DELETE
	err = c.DeleteEntity(context.Background(), "test-entity")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// Verify deleted
	_, err = c.GetEntity(context.Background(), "test-entity")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
