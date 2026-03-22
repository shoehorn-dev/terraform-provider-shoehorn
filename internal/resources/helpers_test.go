package resources

import (
	"testing"
)

func TestYamlQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: `""`},
		{name: "plain string", input: "my-service", want: "my-service"},
		{name: "with colon", input: "key: value", want: `"key: value"`},
		{name: "with hash", input: "name # comment", want: `"name # comment"`},
		{name: "with double quote", input: `say "hello"`, want: `"say \"hello\""`},
		{name: "with newline", input: "line1\nline2", want: `"line1\nline2"`},
		{name: "with backslash", input: `path\to\file`, want: `"path\\to\\file"`},
		{name: "leading space", input: " leading", want: `" leading"`},
		{name: "trailing space", input: "trailing ", want: `"trailing "`},
		{name: "leading dash", input: "- item", want: `"- item"`},
		{name: "leading bracket", input: "[1,2,3]", want: `"[1,2,3]"`},
		{name: "leading brace", input: "{key: val}", want: `"{key: val}"`},
		{name: "boolean true", input: "true", want: `"true"`},
		{name: "boolean false", input: "false", want: `"false"`},
		{name: "boolean yes", input: "yes", want: `"yes"`},
		{name: "boolean no", input: "no", want: `"no"`},
		{name: "boolean Yes", input: "Yes", want: `"Yes"`},
		{name: "boolean on", input: "on", want: `"on"`},
		{name: "boolean OFF", input: "OFF", want: `"OFF"`},
		{name: "null", input: "null", want: `"null"`},
		{name: "URL with colon", input: "https://example.com", want: `"https://example.com"`},
		{name: "with ampersand", input: "a&b", want: `"a&b"`},
		{name: "with asterisk", input: "*alias", want: `"*alias"`},
		{name: "with exclamation", input: "!tag", want: `"!tag"`},
		{name: "with comma", input: "a,b", want: `"a,b"`},
		{name: "with at sign", input: "user@host", want: `"user@host"`},
		{name: "plain slug", input: "platform-team", want: "platform-team"},
		{name: "plain alphanumeric", input: "service123", want: "service123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yamlQuote(tt.input)
			if got != tt.want {
				t.Errorf("yamlQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStringValueOrNull_NonEmpty(t *testing.T) {
	result := stringValueOrNull("hello")
	if result.IsNull() {
		t.Error("stringValueOrNull(\"hello\") should not be null")
	}
	if result.ValueString() != "hello" {
		t.Errorf("stringValueOrNull(\"hello\") = %q, want %q", result.ValueString(), "hello")
	}
}

func TestStringValueOrNull_Empty(t *testing.T) {
	result := stringValueOrNull("")
	if !result.IsNull() {
		t.Error("stringValueOrNull(\"\") should be null")
	}
}
