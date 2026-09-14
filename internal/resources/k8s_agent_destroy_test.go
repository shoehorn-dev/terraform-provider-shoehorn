package resources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/shoehorn-dev/terraform-provider-shoehorn/internal/client"
)

// k8sAgentStateFor builds a tfsdk.State populated with one agent, which is what the framework
// hands Delete at runtime.
func k8sAgentStateFor(t *testing.T, clusterID string) tfsdk.State {
	t.Helper()
	ctx := context.Background()

	r := NewK8sAgentResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	st := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
	}

	diags := st.Set(ctx, K8sAgentResourceModel{
		ID:          types.StringValue(clusterID),
		ClusterID:   types.StringValue(clusterID),
		Name:        types.StringValue("test-agent"),
		Description: types.StringNull(),
		ExpiresIn:   types.Int64Null(),
		Token:       types.StringNull(),
		TokenPrefix: types.StringNull(),
		Status:      types.StringNull(),
		ExpiresAt:   types.StringNull(),
		CreatedAt:   types.StringNull(),
	})
	if diags.HasError() {
		t.Fatalf("failed to build state: %v", diags)
	}
	return st
}

// Shoehorn v0.7.0 made the agent DELETE tenant-scoped, so it resolves ownership first and answers
// 404 when the caller's tenant owns no token for that cluster. That branch was unreachable before:
// the old implementation deleted unconditionally and returned 204. Destroy therefore has to tolerate
// a 404, exactly as every sibling resource already does (forge_approval_policy_resource.go,
// forge_mold_resource.go, governance_action_resource.go, marketplace_installation_resource.go).
//
// Without that tolerance, `terraform destroy` run twice, or run after the cluster was removed from
// the UI, fails the whole destroy and strands state.
func TestK8sAgentResource_Delete_Tolerates404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/revoke"):
			// Already revoked alongside the deletion; Delete only warns on this.
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Agent not found"}`))
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"Agent not found"}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	r := &K8sAgentResource{client: client.NewClient(server.URL, "test-key", 10*time.Second)}

	resp := &resource.DeleteResponse{State: k8sAgentStateFor(t, "gone-cluster")}
	r.Delete(context.Background(), resource.DeleteRequest{
		State: k8sAgentStateFor(t, "gone-cluster"),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("destroy failed on an agent the API reports as already gone: %v", resp.Diagnostics.Errors())
	}
}

// A delete that fails for any other reason must still surface as an error, so a real outage is not
// mistaken for "already deleted" and silently dropped from state.
func TestK8sAgentResource_Delete_StillFailsOnServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/revoke") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer server.Close()

	r := &K8sAgentResource{client: client.NewClient(server.URL, "test-key", 10*time.Second)}

	resp := &resource.DeleteResponse{State: k8sAgentStateFor(t, "broken-cluster")}
	r.Delete(context.Background(), resource.DeleteRequest{
		State: k8sAgentStateFor(t, "broken-cluster"),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("destroy reported success even though the API returned 500")
	}
}
