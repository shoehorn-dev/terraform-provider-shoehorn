package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

func TestNotificationSubscriptionResource_Metadata(t *testing.T) {
	r := NewNotificationSubscriptionResource()
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "shoehorn"}, resp)

	if resp.TypeName != "shoehorn_notification_subscription" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "shoehorn_notification_subscription")
	}
}

func TestNotificationSubscriptionResource_Schema_HasAttributes(t *testing.T) {
	r := NewNotificationSubscriptionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	wantAttrs := []string{"id", "scope", "scope_id", "name", "enabled", "event_types",
		"min_severity", "entity_filter", "channel_type", "cadence", "cadence_config",
		"created_at", "updated_at"}
	for _, name := range wantAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}

	wantBlocks := []string{"slack", "webhook", "email"}
	for _, name := range wantBlocks {
		if _, ok := resp.Schema.Blocks[name]; !ok {
			t.Errorf("schema missing block %q", name)
		}
	}
}

func TestNotificationSubscriptionResource_Schema_IDIsComputed(t *testing.T) {
	r := NewNotificationSubscriptionResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	idAttr := resp.Schema.Attributes["id"]
	if idAttr == nil || !idAttr.IsComputed() {
		t.Error("id should be computed")
	}
}

// TestSecretRefValidator covers the reference-only secret rule: literal secrets
// are rejected, secret:// and env:// refs are accepted.
func TestSecretRefValidator(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"reject https literal", "https://hooks.slack.com/services/T000/B000/XXXX", true},
		{"reject bare token", "xoxb-1234567890-abcdef", true},
		{"reject empty-ish literal", "not-a-ref", true},
		{"accept secret ref", "secret://slack", false},
		{"accept env ref", "env://NOTIFICATIONS_X", false},
		{"accept secret ref with path", "secret://notifications/slack-webhook", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := secretRefValidator{}
			req := validator.StringRequest{
				Path:        path.Root("url_secret"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			gotError := resp.Diagnostics.HasError()
			if gotError != tt.wantError {
				t.Errorf("ValidateString(%q): hasError = %v, want %v (%v)",
					tt.value, gotError, tt.wantError, resp.Diagnostics)
			}
		})
	}
}

// TestSecretRefValidator_NullAndUnknown verifies null/unknown values pass without error.
func TestSecretRefValidator_NullAndUnknown(t *testing.T) {
	v := secretRefValidator{}

	for _, cv := range []types.String{types.StringNull(), types.StringUnknown()} {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("token_secret"),
			ConfigValue: cv,
		}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("null/unknown value should not error: %v", resp.Diagnostics)
		}
	}
}

func TestIsSecretRef(t *testing.T) {
	cases := map[string]bool{
		"secret://x":      true,
		"env://X":         true,
		"https://x":       false,
		"":                false,
		"secretsomething": false,
	}
	for in, want := range cases {
		if got := isSecretRef(in); got != want {
			t.Errorf("isSecretRef(%q) = %v, want %v", in, got, want)
		}
	}
}

// TestBuildChannelConfig_SlackWrapsSecretRefs verifies that a Slack url_secret ref
// is wrapped into the {"mode":"ref","value":"<ref>"} envelope.
func TestBuildChannelConfig_SlackWrapsSecretRefs(t *testing.T) {
	plan := &NotificationSubscriptionResourceModel{
		ChannelType: types.StringValue("slack"),
		Slack: &slackConfigModel{
			Mode:      types.StringValue("webhook"),
			URLSecret: types.StringValue("secret://slack-webhook"),
		},
	}

	cfg, diags := buildChannelConfig(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(cfg, &parsed); err != nil {
		t.Fatalf("channel_config not valid JSON: %v", err)
	}

	urlSecret, ok := parsed["url_secret"].(map[string]interface{})
	if !ok {
		t.Fatalf("url_secret not an object: %v", parsed["url_secret"])
	}
	if urlSecret["mode"] != "ref" {
		t.Errorf("url_secret.mode = %v, want %q", urlSecret["mode"], "ref")
	}
	if urlSecret["value"] != "secret://slack-webhook" {
		t.Errorf("url_secret.value = %v, want %q", urlSecret["value"], "secret://slack-webhook")
	}
}

// TestRawSecretField_UnwrapsEnvelope verifies the read path unwraps the envelope.
func TestRawSecretField_UnwrapsEnvelope(t *testing.T) {
	cfg := map[string]json.RawMessage{
		"url_secret": json.RawMessage(`{"mode":"ref","value":"secret://slack-webhook"}`),
	}
	got := rawSecretField(cfg, "url_secret")
	if got.ValueString() != "secret://slack-webhook" {
		t.Errorf("rawSecretField = %q, want %q", got.ValueString(), "secret://slack-webhook")
	}

	if missing := rawSecretField(cfg, "token_secret"); !missing.IsNull() {
		t.Errorf("missing field should be null, got %q", missing.ValueString())
	}
}

func TestNotificationSubscriptionResource_ImportState_RejectsBadID(t *testing.T) {
	r := NewNotificationSubscriptionResource().(*NotificationSubscriptionResource)
	resp := &resource.ImportStateResponse{}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "just-an-id"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for non-3-part import ID")
	}
}

// TestNotificationSubscriptionClient_CRUD exercises the client against a mock server.
func TestNotificationSubscriptionClient_CRUD(t *testing.T) {
	store := map[string]map[string]interface{}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/notifications/subscriptions":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			sub := map[string]interface{}{
				"id":             "sub-123",
				"scope":          req["scope"],
				"scope_id":       req["scope_id"],
				"name":           req["name"],
				"enabled":        req["enabled"],
				"event_types":    req["event_types"],
				"min_severity":   req["min_severity"],
				"channel_type":   req["channel_type"],
				"channel_config": req["channel_config"],
				"cadence":        req["cadence"],
				"created_at":     "2026-05-21T10:00:00Z",
				"updated_at":     "2026-05-21T10:00:00Z",
			}
			store["sub-123"] = sub
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(sub)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/notifications/subscriptions":
			subs := []map[string]interface{}{}
			for _, s := range store {
				subs = append(subs, s)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"subscriptions": subs, "total": len(subs)})

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/notifications/subscriptions/sub-123":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			sub := store["sub-123"]
			sub["name"] = req["name"]
			sub["updated_at"] = "2026-05-21T11:00:00Z"
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(sub)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/notifications/subscriptions/sub-123":
			delete(store, "sub-123")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"%s %s"}`, r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(server.URL, "key", 30*time.Second)
	ctx := context.Background()

	// CREATE
	sub, err := c.CreateNotificationSubscription(ctx, client.NotificationSubscriptionRequest{
		Scope:         "team",
		ScopeID:       "team-uuid",
		Name:          "CVE alerts",
		Enabled:       true,
		EventTypes:    []string{"cve.detected"},
		MinSeverity:   "warning",
		ChannelType:   "slack",
		ChannelConfig: json.RawMessage(`{"mode":"webhook","url_secret":{"mode":"ref","value":"secret://slack"}}`),
		Cadence:       "realtime",
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if sub.ID != "sub-123" {
		t.Errorf("CREATE: ID = %q, want %q", sub.ID, "sub-123")
	}

	// READ (list-and-filter)
	subs, err := c.ListNotificationSubscriptions(ctx, "team", "team-uuid")
	if err != nil {
		t.Fatalf("LIST failed: %v", err)
	}
	if len(subs) != 1 || subs[0].Name != "CVE alerts" {
		t.Errorf("LIST: got %d subs, want 1 named 'CVE alerts'", len(subs))
	}

	// UPDATE
	updated, err := c.UpdateNotificationSubscription(ctx, "sub-123", client.NotificationSubscriptionRequest{
		Scope: "team", ScopeID: "team-uuid", Name: "CVE alerts v2",
		EventTypes: []string{"cve.detected"}, MinSeverity: "warning",
		ChannelType: "slack", Cadence: "realtime", Enabled: true,
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updated.Name != "CVE alerts v2" {
		t.Errorf("UPDATE: Name = %q, want %q", updated.Name, "CVE alerts v2")
	}

	// DELETE
	if err := c.DeleteNotificationSubscription(ctx, "sub-123"); err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	subs, _ = c.ListNotificationSubscriptions(ctx, "team", "team-uuid")
	if len(subs) != 0 {
		t.Errorf("after DELETE: got %d subs, want 0", len(subs))
	}
}

// --- Acceptance test (gated by TF_ACC) ---

// TestAccNotificationSubscriptionResource covers the create -> read -> update ->
// delete -> import lifecycle against a live Shoehorn API. It is gated by TF_ACC
// (set by `make testacc`) and skipped during the normal unit-test run.
func TestAccNotificationSubscriptionResource(t *testing.T) {
	if strings.TrimSpace(os.Getenv("TF_ACC")) == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	host := os.Getenv("SHOEHORN_HOST")
	apiKey := os.Getenv("SHOEHORN_API_KEY")
	scopeID := os.Getenv("SHOEHORN_TEST_TEAM_ID")
	if host == "" || apiKey == "" || scopeID == "" {
		t.Skip("SHOEHORN_HOST, SHOEHORN_API_KEY and SHOEHORN_TEST_TEAM_ID required for acceptance test")
	}

	c := client.NewClient(host, apiKey, 30*time.Second)
	ctx := context.Background()

	// CREATE: a Slack webhook-mode subscription with a secret:// ref.
	sub, err := c.CreateNotificationSubscription(ctx, client.NotificationSubscriptionRequest{
		Scope:       "team",
		ScopeID:     scopeID,
		Name:        "tf-acc CVE alerts",
		Enabled:     true,
		EventTypes:  []string{"cve.detected"},
		MinSeverity: "warning",
		ChannelType: "slack",
		ChannelConfig: json.RawMessage(
			`{"mode":"webhook","url_secret":{"mode":"ref","value":"secret://slack-webhook"}}`),
		Cadence: "realtime",
	})
	if err != nil {
		t.Fatalf("acc CREATE failed: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteNotificationSubscription(ctx, sub.ID) })

	// READ: list within scope and find the row.
	subs, err := c.ListNotificationSubscriptions(ctx, "team", scopeID)
	if err != nil {
		t.Fatalf("acc LIST failed: %v", err)
	}
	var found bool
	for _, s := range subs {
		if s.ID == sub.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("acc READ: created subscription %s not found in list", sub.ID)
	}

	// UPDATE
	if _, err := c.UpdateNotificationSubscription(ctx, sub.ID, client.NotificationSubscriptionRequest{
		Scope: "team", ScopeID: scopeID, Name: "tf-acc CVE alerts v2",
		EventTypes: []string{"cve.detected"}, MinSeverity: "critical",
		ChannelType: "slack", Cadence: "realtime", Enabled: true,
		ChannelConfig: json.RawMessage(
			`{"mode":"webhook","url_secret":{"mode":"ref","value":"secret://slack-webhook"}}`),
	}); err != nil {
		t.Fatalf("acc UPDATE failed: %v", err)
	}

	// DELETE
	if err := c.DeleteNotificationSubscription(ctx, sub.ID); err != nil {
		t.Fatalf("acc DELETE failed: %v", err)
	}
}
