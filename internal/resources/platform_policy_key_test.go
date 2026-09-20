package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validateKey runs the key validator over one configured value.
func validateKey(t *testing.T, value types.String) validator.StringResponse {
	t.Helper()
	resp := validator.StringResponse{}
	manageablePolicyKey().ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("key"),
		ConfigValue: value,
	}, &resp)
	return resp
}

// details joins every diagnostic's summary and detail so a test can look for the
// phrase that has to reach the person running terraform.
func details(resp validator.StringResponse) string {
	var b strings.Builder
	for _, d := range resp.Diagnostics {
		b.WriteString(d.Summary())
		b.WriteString(" ")
		b.WriteString(d.Detail())
		b.WriteString("\n")
	}
	return b.String()
}

func TestManageablePolicyKey_AcceptsTheTwoSettings(t *testing.T) {
	for _, key := range []string{"governance-auto-actions", "api-key-expiration"} {
		resp := validateKey(t, types.StringValue(key))
		if resp.Diagnostics.HasError() {
			t.Errorf("%s: %s", key, details(resp))
		}
	}
}

func TestManageablePolicyKey_RefusesAlwaysOnPolicies(t *testing.T) {
	alwaysOn := []string{
		"tenant-isolation", "rbac-enforcement", "audit-logging", "data-retention",
		"change-tracking", "repo-auto-sync", "github-rate-limit", "retry-failed-repos",
	}
	for _, key := range alwaysOn {
		resp := validateKey(t, types.StringValue(key))
		if !resp.Diagnostics.HasError() {
			t.Errorf("%s: always-on policy accepted", key)
			continue
		}
		text := details(resp)
		if !strings.Contains(text, key) {
			t.Errorf("%s: message does not name the policy: %s", key, text)
		}
		if !strings.Contains(text, "always on") {
			t.Errorf("%s: message does not say the policy is always on: %s", key, text)
		}
		if !strings.Contains(text, "shoehorn_platform_policies") {
			t.Errorf("%s: message does not point at the data source: %s", key, text)
		}
	}
}

func TestManageablePolicyKey_RefusesRemovedPolicies(t *testing.T) {
	for _, key := range []string{"required-entity-docs", "entity-metadata-validation", "stale-entity-cleanup"} {
		resp := validateKey(t, types.StringValue(key))
		if !resp.Diagnostics.HasError() {
			t.Errorf("%s: removed policy accepted", key)
			continue
		}
		text := details(resp)
		if !strings.Contains(text, key) {
			t.Errorf("%s: message does not name the policy: %s", key, text)
		}
		if !strings.Contains(text, "0.7.0") {
			t.Errorf("%s: message does not name the release that dropped it: %s", key, text)
		}
	}
}

func TestManageablePolicyKey_RefusesUnknownPolicyAndListsWhatWorks(t *testing.T) {
	resp := validateKey(t, types.StringValue("no-such-policy"))
	if !resp.Diagnostics.HasError() {
		t.Fatal("unknown policy accepted")
	}
	text := details(resp)
	if !strings.Contains(text, "no-such-policy") {
		t.Errorf("message does not name the policy: %s", text)
	}
	for _, key := range []string{"governance-auto-actions", "api-key-expiration"} {
		if !strings.Contains(text, key) {
			t.Errorf("message does not list %s as manageable: %s", key, text)
		}
	}
}

// A key that comes from a variable is unknown at validate time. Refusing it there
// would fail plans that are fine.
func TestManageablePolicyKey_PassesNullAndUnknownThrough(t *testing.T) {
	for name, value := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
	} {
		if resp := validateKey(t, value); resp.Diagnostics.HasError() {
			t.Errorf("%s value rejected: %s", name, details(resp))
		}
	}
}

func TestManageablePolicyKey_HasADescription(t *testing.T) {
	v := manageablePolicyKey()
	if v.Description(context.Background()) == "" {
		t.Error("validator has no description")
	}
	if v.MarkdownDescription(context.Background()) == "" {
		t.Error("validator has no markdown description")
	}
}
