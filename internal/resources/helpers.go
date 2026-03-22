package resources

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringValueOrNull returns a types.StringValue for non-empty strings,
// or types.StringNull for empty strings. This ensures that optional fields
// are properly cleared in Terraform state when the API returns empty values.
func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// yamlQuote returns a YAML-safe representation of a string value.
// If the string contains characters that are special in YAML (colons, hashes,
// quotes, newlines, leading/trailing spaces), it wraps the value in double quotes
// with proper escaping. Otherwise, it returns the string as-is.
func yamlQuote(s string) string {
	if s == "" {
		return `""`
	}
	needsQuoting := strings.ContainsAny(s, ":#{}&*!|>'\"\n\\@`,") ||
		strings.HasPrefix(s, " ") ||
		strings.HasSuffix(s, " ") ||
		strings.HasPrefix(s, "-") ||
		strings.HasPrefix(s, "[") ||
		strings.HasPrefix(s, "{") ||
		s == "true" || s == "false" || s == "null" || s == "yes" || s == "no" ||
		s == "True" || s == "False" || s == "Yes" || s == "No" ||
		s == "TRUE" || s == "FALSE" || s == "YES" || s == "NO" ||
		s == "on" || s == "off" || s == "On" || s == "Off" || s == "ON" || s == "OFF" ||
		s == "y" || s == "n" || s == "Y" || s == "N"
	if !needsQuoting {
		return s
	}
	// Escape backslashes and double quotes, then wrap in double quotes
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return `"` + s + `"`
}
