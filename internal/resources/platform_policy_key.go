package resources

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Mirrors internal/platformpolicy so a wrong key fails at plan time; keep in step.
var (
	// configurablePolicyKeys are the settings a tenant can change.
	configurablePolicyKeys = []string{"governance-auto-actions", "api-key-expiration"}

	// alwaysOnPolicyKeys are protections Shoehorn keeps on for every tenant.
	alwaysOnPolicyKeys = []string{
		"tenant-isolation", "rbac-enforcement", "audit-logging", "data-retention",
		"change-tracking", "repo-auto-sync", "github-rate-limit", "retry-failed-repos",
	}

	// removedPolicyKeys were listed by earlier releases with nothing behind them.
	removedPolicyKeys = []string{"required-entity-docs", "entity-metadata-validation", "stale-entity-cleanup"}
)

// removedPolicyRelease is the release that dropped removedPolicyKeys.
const removedPolicyRelease = "0.7.0"

// manageablePolicyKeyValidator refuses a policy key Terraform cannot manage.
type manageablePolicyKeyValidator struct{}

// manageablePolicyKey validates the key of a shoehorn_platform_policy.
func manageablePolicyKey() validator.String {
	return manageablePolicyKeyValidator{}
}

func (v manageablePolicyKeyValidator) Description(_ context.Context) string {
	return fmt.Sprintf("must be a policy Terraform can manage: %s", strings.Join(configurablePolicyKeys, ", "))
}

func (v manageablePolicyKeyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v manageablePolicyKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// A key from a variable is unknown until apply; Terraform rechecks it then.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	key := req.ConfigValue.ValueString()
	switch {
	case slices.Contains(configurablePolicyKeys, key):
		return
	case slices.Contains(alwaysOnPolicyKeys, key):
		resp.Diagnostics.AddAttributeError(req.Path, "Policy is always on",
			fmt.Sprintf("Shoehorn keeps %q on for every tenant, so Terraform can't manage it. "+
				"Remove this resource. To read the policy, use the shoehorn_platform_policies data source.", key))
	case slices.Contains(removedPolicyKeys, key):
		resp.Diagnostics.AddAttributeError(req.Path, "Policy was removed",
			fmt.Sprintf("Shoehorn dropped %q in %s and it no longer does anything. Remove this resource.",
				key, removedPolicyRelease))
	default:
		resp.Diagnostics.AddAttributeError(req.Path, "Unknown policy",
			fmt.Sprintf("%q isn't a Shoehorn platform policy. Terraform manages two: %s.",
				key, strings.Join(configurablePolicyKeys, " and ")))
	}
}
