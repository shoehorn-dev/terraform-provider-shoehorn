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

func TestListApprovalPolicies_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/forge/approval-policies" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policies": []map[string]interface{}{
				{
					"id": "ap-1", "name": "Security Review", "description": "Requires security team approval",
					"enabled": true,
					"approval_chain": []map[string]interface{}{
						{"name": "Security Team", "approvers": []string{"user-1", "user-2"}, "required_count": 0},
					},
					"created_at": "2025-06-01T10:00:00Z",
				},
				{
					"id": "ap-2", "name": "Change Advisory Board", "description": "CAB approval for production changes",
					"enabled": false,
					"approval_chain": []map[string]interface{}{
						{"name": "Manager Approval", "approvers": []string{"user-3"}, "required_count": 1},
						{"name": "CAB Review", "approvers": []string{"user-4", "user-5"}, "required_count": 0},
					},
					"created_at": "2025-06-02T10:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	policies, err := c.ListApprovalPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListApprovalPolicies() error = %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("policy count = %d, want 2", len(policies))
	}
	if policies[0].ID != "ap-1" {
		t.Errorf("ID = %q, want %q", policies[0].ID, "ap-1")
	}
	if policies[0].Name != "Security Review" {
		t.Errorf("Name = %q, want %q", policies[0].Name, "Security Review")
	}
	if !policies[0].Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(policies[0].ApprovalChain) != 1 {
		t.Fatalf("Steps count = %d, want 1", len(policies[0].ApprovalChain))
	}
	if policies[0].ApprovalChain[0].Name != "Security Team" {
		t.Errorf("Steps[0].Name = %q, want %q", policies[0].ApprovalChain[0].Name, "Security Team")
	}
	if len(policies[0].ApprovalChain[0].Approvers) != 2 {
		t.Errorf("Steps[0].Approvers count = %d, want 2", len(policies[0].ApprovalChain[0].Approvers))
	}
	if policies[1].Name != "Change Advisory Board" {
		t.Errorf("Name = %q, want %q", policies[1].Name, "Change Advisory Board")
	}
	if len(policies[1].ApprovalChain) != 2 {
		t.Errorf("Steps count = %d, want 2", len(policies[1].ApprovalChain))
	}
}

func TestGetApprovalPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/forge/approval-policies/ap-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policy": map[string]interface{}{
				"id": "ap-1", "name": "Security Review", "description": "Requires security team approval",
				"enabled": true,
				"approval_chain": []map[string]interface{}{
					{"name": "Security Team", "approvers": []string{"user-1", "user-2"}, "required_count": 0},
				},
				"created_at": "2025-06-01T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	policy, err := c.GetApprovalPolicy(context.Background(), "ap-1")
	if err != nil {
		t.Fatalf("GetApprovalPolicy() error = %v", err)
	}
	if policy.ID != "ap-1" {
		t.Errorf("ID = %q, want %q", policy.ID, "ap-1")
	}
	if policy.Name != "Security Review" {
		t.Errorf("Name = %q, want %q", policy.Name, "Security Review")
	}
	if policy.Description != "Requires security team approval" {
		t.Errorf("Description = %q, want %q", policy.Description, "Requires security team approval")
	}
	if !policy.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(policy.ApprovalChain) != 1 {
		t.Fatalf("Steps count = %d, want 1", len(policy.ApprovalChain))
	}
	if len(policy.ApprovalChain[0].Approvers) != 2 {
		t.Errorf("ApprovalChain[0].Approvers count = %d, want 2", len(policy.ApprovalChain[0].Approvers))
	}
	if policy.ApprovalChain[0].Name != "Security Team" {
		t.Errorf("Steps[0].Name = %q, want %q", policy.ApprovalChain[0].Name, "Security Team")
	}
	if policy.ApprovalChain[0].RequiredCount != 0 {
		t.Errorf("ApprovalChain[0].RequiredCount = %d, want 0", policy.ApprovalChain[0].RequiredCount)
	}
	if policy.CreatedAt != "2025-06-01T10:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", policy.CreatedAt, "2025-06-01T10:00:00Z")
	}
}

func TestCreateApprovalPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/forge/approval-policies" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req CreateApprovalPolicyRequest
		json.Unmarshal(body, &req)

		if req.Name != "Security Review" {
			t.Errorf("Name = %q, want %q", req.Name, "Security Review")
		}
		if !req.Enabled {
			t.Error("Enabled = false, want true")
		}
		if len(req.ApprovalChain) != 1 {
			t.Fatalf("Steps count = %d, want 1", len(req.ApprovalChain))
		}
		if req.ApprovalChain[0].Name != "Security Team" {
			t.Errorf("Steps[0].Name = %q, want %q", req.ApprovalChain[0].Name, "Security Team")
		}
		if len(req.ApprovalChain[0].Approvers) != 2 {
			t.Errorf("Steps[0].Approvers count = %d, want 2", len(req.ApprovalChain[0].Approvers))
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policy": map[string]interface{}{
				"id": "ap-new", "name": req.Name, "description": req.Description,
				"enabled": req.Enabled,
				"approval_chain": []map[string]interface{}{
					{"name": req.ApprovalChain[0].Name, "approvers": req.ApprovalChain[0].Approvers, "required_count": req.ApprovalChain[0].RequiredCount},
				},
				"created_at": "2025-06-10T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	policy, err := c.CreateApprovalPolicy(context.Background(), CreateApprovalPolicyRequest{
		Name:        "Security Review",
		Description: "Requires security team approval",
		Enabled:     true,
		ApprovalChain: []ApprovalStep{
			{Name: "Security Team", Approvers: []string{"user-1", "user-2"}, RequiredCount: 0},
		},
	})
	if err != nil {
		t.Fatalf("CreateApprovalPolicy() error = %v", err)
	}
	if policy.ID != "ap-new" {
		t.Errorf("ID = %q, want %q", policy.ID, "ap-new")
	}
	if policy.Name != "Security Review" {
		t.Errorf("Name = %q, want %q", policy.Name, "Security Review")
	}
	if !policy.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(policy.ApprovalChain) != 1 {
		t.Fatalf("Steps count = %d, want 1", len(policy.ApprovalChain))
	}
	if policy.ApprovalChain[0].Name != "Security Team" {
		t.Errorf("Steps[0].Name = %q, want %q", policy.ApprovalChain[0].Name, "Security Team")
	}
}

func TestUpdateApprovalPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/forge/approval-policies/ap-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req UpdateApprovalPolicyRequest
		json.Unmarshal(body, &req)

		if req.Name != "Updated Security Review" {
			t.Errorf("Name = %q, want %q", req.Name, "Updated Security Review")
		}
		if req.Enabled == nil || !*req.Enabled {
			t.Error("Enabled should be true")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policy": map[string]interface{}{
				"id": "ap-1", "name": "Updated Security Review", "description": "Updated description",
				"enabled": true,
				"approval_chain": []map[string]interface{}{
					{"name": "Security Team", "approvers": []string{"user-1", "user-2"}, "required_count": 1},
				},
				"updated_at": "2025-06-10T11:00:00Z",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	enabled := true
	policy, err := c.UpdateApprovalPolicy(context.Background(), "ap-1", UpdateApprovalPolicyRequest{
		Name:    "Updated Security Review",
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("UpdateApprovalPolicy() error = %v", err)
	}
	if policy.Name != "Updated Security Review" {
		t.Errorf("Name = %q, want %q", policy.Name, "Updated Security Review")
	}
	if !policy.Enabled {
		t.Error("Enabled = false, want true")
	}
	if policy.UpdatedAt != "2025-06-10T11:00:00Z" {
		t.Errorf("UpdatedAt = %q, want %q", policy.UpdatedAt, "2025-06-10T11:00:00Z")
	}
}

func TestDeleteApprovalPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/forge/approval-policies/ap-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteApprovalPolicy(context.Background(), "ap-1")
	if err != nil {
		t.Fatalf("DeleteApprovalPolicy() error = %v", err)
	}
}

func TestGetApprovalPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Approval policy not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetApprovalPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found policy, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true, got false for error: %v", err)
	}
}

func TestListApprovalPolicies_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"policies": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	policies, err := c.ListApprovalPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListApprovalPolicies() error = %v", err)
	}
	if len(policies) != 0 {
		t.Errorf("policy count = %d, want 0", len(policies))
	}
}

func TestCreateApprovalPolicy_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"code":"INTERNAL","message":"internal server error"}`)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.CreateApprovalPolicy(context.Background(), CreateApprovalPolicyRequest{
		Name:    "Test Policy",
		Enabled: true,
		ApprovalChain: []ApprovalStep{
			{Name: "Approval", Approvers: []string{"user-1"}, RequiredCount: 1},
		},
	})
	if err == nil {
		t.Fatal("CreateApprovalPolicy() expected error for 500 response, got nil")
	}
}

func TestDeleteApprovalPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "NOT_FOUND", "message": "approval policy not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteApprovalPolicy(context.Background(), "ap-nonexistent")
	if err == nil {
		t.Fatal("DeleteApprovalPolicy() expected error for 404, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestApprovalPolicyClient_Lifecycle(t *testing.T) {
	policies := make(map[string]map[string]interface{})
	nextID := 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/forge/approval-policies":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			id := fmt.Sprintf("ap-%d", nextID)
			policy := map[string]interface{}{
				"id": id, "name": req["name"], "description": req["description"],
				"enabled": req["enabled"], "approval_chain": req["approval_chain"],
				"created_at": "2025-06-10T10:00:00Z",
			}
			policies[id] = policy
			nextID++

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"policy": policy})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/forge/approval-policies/ap-1":
			if p, ok := policies["ap-1"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"policy": p})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"code": "NOT_FOUND", "message": "not found"})
			}

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/forge/approval-policies":
			list := make([]map[string]interface{}, 0, len(policies))
			for _, p := range policies {
				list = append(list, p)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"policies": list})

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/forge/approval-policies/ap-1":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			if p, ok := policies["ap-1"]; ok {
				if name, ok := req["name"].(string); ok && name != "" {
					p["name"] = name
				}
				if desc, ok := req["description"].(string); ok && desc != "" {
					p["description"] = desc
				}
				if enabled, ok := req["enabled"]; ok && enabled != nil {
					p["enabled"] = enabled
				}
				if steps, ok := req["approval_chain"]; ok && steps != nil {
					p["approval_chain"] = steps
				}
				p["updated_at"] = "2025-06-10T11:00:00Z"
				policies["ap-1"] = p
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"policy": p})
			}

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/forge/approval-policies/ap-1":
			delete(policies, "ap-1")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"not found"}`)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	created, err := c.CreateApprovalPolicy(context.Background(), CreateApprovalPolicyRequest{
		Name:        "Lifecycle Test Policy",
		Description: "Created for lifecycle test",
		Enabled:     true,
		ApprovalChain: []ApprovalStep{
			{Name: "Lead Approval", Approvers: []string{"user-lead"}, RequiredCount: 1},
		},
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if created.ID != "ap-1" {
		t.Errorf("CREATE: ID = %q, want %q", created.ID, "ap-1")
	}
	if created.Name != "Lifecycle Test Policy" {
		t.Errorf("CREATE: Name = %q, want %q", created.Name, "Lifecycle Test Policy")
	}

	// READ
	read, err := c.GetApprovalPolicy(context.Background(), "ap-1")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if read.Name != "Lifecycle Test Policy" {
		t.Errorf("READ: Name = %q, want %q", read.Name, "Lifecycle Test Policy")
	}
	if read.Description != "Created for lifecycle test" {
		t.Errorf("READ: Description = %q, want %q", read.Description, "Created for lifecycle test")
	}

	// LIST
	list, err := c.ListApprovalPolicies(context.Background())
	if err != nil {
		t.Fatalf("LIST failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("LIST: count = %d, want 1", len(list))
	}

	// UPDATE
	enabled := false
	updated, err := c.UpdateApprovalPolicy(context.Background(), "ap-1", UpdateApprovalPolicyRequest{
		Name:    "Updated Lifecycle Policy",
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	if updated.Name != "Updated Lifecycle Policy" {
		t.Errorf("UPDATE: Name = %q, want %q", updated.Name, "Updated Lifecycle Policy")
	}

	// VERIFY UPDATE
	readAgain, err := c.GetApprovalPolicy(context.Background(), "ap-1")
	if err != nil {
		t.Fatalf("VERIFY UPDATE failed: %v", err)
	}
	if readAgain.Name != "Updated Lifecycle Policy" {
		t.Errorf("VERIFY UPDATE: Name = %q, want %q", readAgain.Name, "Updated Lifecycle Policy")
	}

	// DELETE
	err = c.DeleteApprovalPolicy(context.Background(), "ap-1")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// VERIFY DELETE
	_, err = c.GetApprovalPolicy(context.Background(), "ap-1")
	if err == nil {
		t.Fatal("VERIFY DELETE: expected error after deletion, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("VERIFY DELETE: expected IsNotFound, got: %v", err)
	}
}
