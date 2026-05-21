package resources

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// secretRefValidator rejects any channel-secret value that is not a reference to
// a secret manager entry. Terraform-managed subscriptions must point at a
// `secret://NAME` or `env://VAR` ref so no plaintext secret ever lands in state.
type secretRefValidator struct{}

// Description returns a plain-text description of the validator's behavior.
func (v secretRefValidator) Description(_ context.Context) string {
	return "value must be a secret reference (secret://NAME or env://VAR)"
}

// MarkdownDescription returns a Markdown description of the validator's behavior.
func (v secretRefValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a secret reference (`secret://NAME` or `env://VAR`)"
}

// ValidateString checks that the configured value is a valid secret reference.
func (v secretRefValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !isSecretRef(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Secret Reference",
			"Terraform-managed notification subscriptions must reference a secret manager "+
				"entry, not a literal secret. Use a `secret://NAME` ref (resolved from the "+
				"deployment's notifications.secrets) or an `env://VAR` ref. Pasted literal "+
				"secrets are portal-only: anything in HCL lands in Terraform state in "+
				"plaintext.",
		)
	}
}

// isSecretRef reports whether s is a secret-manager reference.
func isSecretRef(s string) bool {
	return strings.HasPrefix(s, "secret://") || strings.HasPrefix(s, "env://")
}
