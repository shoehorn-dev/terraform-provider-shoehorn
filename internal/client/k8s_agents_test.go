package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListK8sAgents_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/k8s/agents" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"agents": []map[string]interface{}{
				{"clusterId": "prod-east", "name": "Prod US East", "status": "active", "tokenPrefix": "shp_agent_"},
				{"clusterId": "staging", "name": "Staging", "status": "active", "tokenPrefix": "shp_agent_"},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	agents, err := c.ListK8sAgents(context.Background())
	if err != nil {
		t.Fatalf("ListK8sAgents() error = %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("agent count = %d, want 2", len(agents))
	}
	if agents[0].ClusterID != "prod-east" {
		t.Errorf("ClusterID = %q, want %q", agents[0].ClusterID, "prod-east")
	}
}

func TestGetK8sAgent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/k8s/agents/prod-east" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"clusterId": "prod-east", "name": "Prod US East", "status": "active",
			"tokenPrefix": "shp_agent_", "createdAt": "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	agent, err := c.GetK8sAgent(context.Background(), "prod-east")
	if err != nil {
		t.Fatalf("GetK8sAgent() error = %v", err)
	}
	if agent.Name != "Prod US East" {
		t.Errorf("Name = %q, want %q", agent.Name, "Prod US East")
	}
	if agent.Status != "active" {
		t.Errorf("Status = %q, want %q", agent.Status, "active")
	}
}

func TestGetK8sAgent_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "Agent not found"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetK8sAgent(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found agent, got nil")
	}
}

func TestRegisterK8sAgent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/k8s/agents/register" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req RegisterK8sAgentRequest
		json.Unmarshal(body, &req)

		if req.ClusterID != "prod-east" {
			t.Errorf("ClusterID = %q, want %q", req.ClusterID, "prod-east")
		}
		if req.Name != "Prod US East" {
			t.Errorf("Name = %q, want %q", req.Name, "Prod US East")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":       "shp_agent_secrettoken1234567890abcdef",
			"tokenPrefix": "shp_agent_",
			"clusterId":   req.ClusterID,
			"name":        req.Name,
			"createdAt":   "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resp, err := c.RegisterK8sAgent(context.Background(), RegisterK8sAgentRequest{
		ClusterID:   "prod-east",
		Name:        "Prod US East",
		Description: "Production cluster",
	})
	if err != nil {
		t.Fatalf("RegisterK8sAgent() error = %v", err)
	}
	if resp.Token == "" {
		t.Error("Token should not be empty on registration")
	}
	if resp.ClusterID != "prod-east" {
		t.Errorf("ClusterID = %q, want %q", resp.ClusterID, "prod-east")
	}
}

func TestRegisterK8sAgent_WithExpiry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req RegisterK8sAgentRequest
		json.Unmarshal(body, &req)

		if req.ExpiresIn == nil || *req.ExpiresIn != 365 {
			t.Errorf("ExpiresIn = %v, want 365", req.ExpiresIn)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":       "shp_agent_tokenwitexp",
			"tokenPrefix": "shp_agent_",
			"clusterId":   req.ClusterID,
			"name":        req.Name,
			"expiresAt":   "2026-01-15T10:00:00Z",
			"createdAt":   "2025-01-15T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	days := 365
	resp, err := c.RegisterK8sAgent(context.Background(), RegisterK8sAgentRequest{
		ClusterID: "prod-east",
		Name:      "Prod US East",
		ExpiresIn: &days,
	})
	if err != nil {
		t.Fatalf("RegisterK8sAgent() error = %v", err)
	}
	if resp.ExpiresAt != "2026-01-15T10:00:00Z" {
		t.Errorf("ExpiresAt = %q, want %q", resp.ExpiresAt, "2026-01-15T10:00:00Z")
	}
}

func TestRevokeK8sAgent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/k8s/agents/prod-east/revoke" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "revoked", "message": "Agent token revoked successfully"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.RevokeK8sAgent(context.Background(), "prod-east")
	if err != nil {
		t.Fatalf("RevokeK8sAgent() error = %v", err)
	}
}

func TestDeleteK8sAgent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/k8s/agents/prod-east" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteK8sAgent(context.Background(), "prod-east")
	if err != nil {
		t.Fatalf("DeleteK8sAgent() error = %v", err)
	}
}

func TestK8sAgentClient_Lifecycle_Integration(t *testing.T) {
	agents := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/k8s/agents/register":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			clusterID := req["clusterId"].(string)
			agents[clusterID] = map[string]interface{}{
				"clusterId":   clusterID,
				"name":        req["name"],
				"status":      "active",
				"tokenPrefix": "shp_agent_",
				"createdAt":   "2025-01-15T10:00:00Z",
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"token":       "shp_agent_secrettoken",
				"tokenPrefix": "shp_agent_",
				"clusterId":   clusterID,
				"name":        req["name"],
				"createdAt":   "2025-01-15T10:00:00Z",
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/k8s/agents/test-cluster":
			if agent, ok := agents["test-cluster"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(agent)
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND"})
			}

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/k8s/agents/test-cluster/revoke":
			if agent, ok := agents["test-cluster"]; ok {
				agent["status"] = "revoked"
				agents["test-cluster"] = agent
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "revoked"})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/k8s/agents/test-cluster":
			delete(agents, "test-cluster")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND"}`)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// REGISTER
	regResp, err := c.RegisterK8sAgent(context.Background(), RegisterK8sAgentRequest{
		ClusterID: "test-cluster", Name: "Test Cluster",
	})
	if err != nil {
		t.Fatalf("REGISTER failed: %v", err)
	}
	if regResp.Token == "" {
		t.Error("REGISTER: Token should not be empty")
	}

	// READ
	agent, err := c.GetK8sAgent(context.Background(), "test-cluster")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if agent.Name != "Test Cluster" {
		t.Errorf("READ: Name = %q, want %q", agent.Name, "Test Cluster")
	}

	// REVOKE
	err = c.RevokeK8sAgent(context.Background(), "test-cluster")
	if err != nil {
		t.Fatalf("REVOKE failed: %v", err)
	}

	// DELETE
	err = c.DeleteK8sAgent(context.Background(), "test-cluster")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
}
