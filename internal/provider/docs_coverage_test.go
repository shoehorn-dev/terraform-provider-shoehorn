package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func registeredResourceTypes(t *testing.T) []string {
	t.Helper()
	p := &ShoehornProvider{}
	var names []string
	for _, ctor := range p.Resources(context.Background()) {
		var resp resource.MetadataResponse
		ctor().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, &resp)
		names = append(names, resp.TypeName)
	}
	if len(names) < 10 {
		t.Fatalf("found %d resources; the provider registers more than that", len(names))
	}
	return names
}

func registeredDataSourceTypes(t *testing.T) []string {
	t.Helper()
	p := &ShoehornProvider{}
	var names []string
	for _, ctor := range p.DataSources(context.Background()) {
		var resp datasource.MetadataResponse
		ctor().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "shoehorn"}, &resp)
		names = append(names, resp.TypeName)
	}
	if len(names) < 10 {
		t.Fatalf("found %d data sources; the provider registers more than that", len(names))
	}
	return names
}

func repoRoot() string { return filepath.Join("..", "..") }

func TestEveryDataSourceHasADocsPageAndAnExample(t *testing.T) {
	for _, name := range registeredDataSourceTypes(t) {
		short := strings.TrimPrefix(name, "shoehorn_")
		page, err := os.ReadFile(filepath.Join(repoRoot(), "docs", "data-sources", short+".md"))
		if err != nil {
			t.Errorf("%s has no docs/data-sources/%s.md (run `go generate` after adding an example)", name, short)
			continue
		}
		if _, err := os.Stat(filepath.Join(repoRoot(), "examples", "data-sources", name, "data-source.tf")); err != nil {
			t.Errorf("%s has no examples/data-sources/%s/data-source.tf", name, name)
		}
		if !strings.Contains(string(page), "## Example Usage") {
			t.Errorf("docs/data-sources/%s.md has no Example Usage section", short)
		}
	}
}

func TestEveryResourceHasADocsPageAndAnExample(t *testing.T) {
	for _, name := range registeredResourceTypes(t) {
		short := strings.TrimPrefix(name, "shoehorn_")
		if _, err := os.Stat(filepath.Join(repoRoot(), "docs", "resources", short+".md")); err != nil {
			t.Errorf("%s has no docs/resources/%s.md (run `go generate` after adding an example)", name, short)
		}
		if _, err := os.Stat(filepath.Join(repoRoot(), "examples", "resources", name, "resource.tf")); err != nil {
			t.Errorf("%s has no examples/resources/%s/resource.tf", name, name)
		}
	}
}

func TestApprovalPolicyDocsNameTheScopeItNeeds(t *testing.T) {
	page, err := os.ReadFile(filepath.Join(repoRoot(), "docs", "resources", "forge_approval_policy.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "forge:admin") {
		t.Error("forge_approval_policy.md must say that managing approval policies needs the forge:admin scope")
	}

	keyPage, err := os.ReadFile(filepath.Join(repoRoot(), "docs", "resources", "api_key.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(keyPage), "forge:admin") {
		t.Error("api_key.md must tell people which scope approval policies need, since it is where they choose scopes")
	}
}
