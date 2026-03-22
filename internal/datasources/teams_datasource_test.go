package datasources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestTeamsDataSource_Metadata(t *testing.T) {
	d := NewTeamsDataSource()
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_teams" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_teams")
	}
}

func TestTeamsDataSource_Schema_HasTeamsAttribute(t *testing.T) {
	d := NewTeamsDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if _, ok := resp.Schema.Attributes["teams"]; !ok {
		t.Error("schema missing 'teams' attribute")
	}
}

func TestTeamsDataSource_Configure_WithValidClient(t *testing.T) {
	d := &TeamsDataSource{}
	c := client.NewClient("https://test.example.com", "key", 30*time.Second)

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: c,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}
	if d.client != c {
		t.Error("client not set correctly")
	}
}

func TestTeamsDataSource_Configure_WrongType(t *testing.T) {
	d := &TeamsDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: "not a client",
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestTeamsDataSource_Configure_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"teams": []map[string]any{
				{"id": "team-1", "name": "Platform", "slug": "platform"},
			},
		})
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "test-key", 30*time.Second)

	// Verify client can successfully talk to the mock server
	teams, err := c.ListTeams(context.Background())
	if err != nil {
		t.Fatalf("ListTeams() error = %v", err)
	}
	if len(teams) != 1 {
		t.Errorf("team count = %d, want 1", len(teams))
	}

	// Verify datasource configures with this client
	d := &TeamsDataSource{}
	configResp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{
		ProviderData: c,
	}, configResp)

	if configResp.Diagnostics.HasError() {
		t.Fatalf("unexpected configure error: %v", configResp.Diagnostics)
	}
	if d.client != c {
		t.Fatal("client not set after Configure")
	}
}
