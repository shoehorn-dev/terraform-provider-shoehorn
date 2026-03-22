package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testCtx returns a background context for tests.
func testCtx() context.Context {
	return context.Background()
}

// setupClientWithServer creates a test client pointing at the given server.
func setupClientWithServer(server *httptest.Server) *Client {
	return NewClient(server.URL, "test-key", 30*time.Second)
}

// newTestServer creates an httptest.Server and registers cleanup on t.
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// setupClientWithEmptyKeys returns a client whose API returns an empty key list.
func setupClientWithEmptyKeys(t *testing.T) *Client {
	t.Helper()
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"keys": []any{}, "total": 0})
	})
	return setupClientWithServer(server)
}

// setupClientWithEmptyFlags returns a client whose API returns an empty flag list.
func setupClientWithEmptyFlags(t *testing.T) *Client {
	t.Helper()
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"flags": []any{}})
	})
	return setupClientWithServer(server)
}

// setupClientWithEmptyPolicies returns a client whose API returns an empty policy list.
func setupClientWithEmptyPolicies(t *testing.T) *Client {
	t.Helper()
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"policies": []any{}})
	})
	return setupClientWithServer(server)
}

// setupClientWithEmptyRoles returns a client whose API returns an empty role list.
func setupClientWithEmptyRoles(t *testing.T) *Client {
	t.Helper()
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"roles": []any{}, "count": 0})
	})
	return setupClientWithServer(server)
}

// setupClientWith404Server returns a client whose API always returns 404.
func setupClientWith404Server(t *testing.T) *Client {
	t.Helper()
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"code": "NOT_FOUND", "message": "not found"})
	})
	return setupClientWithServer(server)
}
