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

func TestListGovernanceActions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/governance/actions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"actions": []map[string]interface{}{
				{
					"id": "act-1", "entity_id": "svc-web", "entity_name": "web-app",
					"title": "Fix critical CVE", "priority": "critical", "status": "open",
					"source_type": "security", "source_id": "scan-123",
				},
				{
					"id": "act-2", "entity_id": "svc-api", "entity_name": "api-service",
					"title": "Improve test coverage", "priority": "medium", "status": "in_progress",
					"source_type": "scorecard", "source_id": "sc-456",
				},
			},
			"total": 2,
			"summary": map[string]interface{}{
				"critical": 1, "high": 0, "medium": 1, "low": 0,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	actions, total, err := c.ListGovernanceActions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListGovernanceActions() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(actions) != 2 {
		t.Fatalf("action count = %d, want 2", len(actions))
	}
	if actions[0].ID != "act-1" {
		t.Errorf("ID = %q, want %q", actions[0].ID, "act-1")
	}
	if actions[0].Title != "Fix critical CVE" {
		t.Errorf("Title = %q, want %q", actions[0].Title, "Fix critical CVE")
	}
	if actions[0].Priority != "critical" {
		t.Errorf("Priority = %q, want %q", actions[0].Priority, "critical")
	}
	if actions[1].Status != "in_progress" {
		t.Errorf("Status = %q, want %q", actions[1].Status, "in_progress")
	}
}

func TestListGovernanceActions_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/governance/actions" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		// Verify query parameters were passed
		if r.URL.Query().Get("entity_id") != "svc-web" {
			t.Errorf("entity_id = %q, want %q", r.URL.Query().Get("entity_id"), "svc-web")
		}
		if r.URL.Query().Get("status") != "open" {
			t.Errorf("status = %q, want %q", r.URL.Query().Get("status"), "open")
		}
		if r.URL.Query().Get("priority") != "critical" {
			t.Errorf("priority = %q, want %q", r.URL.Query().Get("priority"), "critical")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"actions": []map[string]interface{}{
				{
					"id": "act-1", "entity_id": "svc-web", "entity_name": "web-app",
					"title": "Fix critical CVE", "priority": "critical", "status": "open",
					"source_type": "security", "source_id": "scan-123",
				},
			},
			"total":   1,
			"summary": map[string]interface{}{"critical": 1},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	filters := &GovernanceActionFilters{
		EntityID: "svc-web",
		Status:   "open",
		Priority: "critical",
	}
	actions, total, err := c.ListGovernanceActions(context.Background(), filters)
	if err != nil {
		t.Fatalf("ListGovernanceActions() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(actions) != 1 {
		t.Fatalf("action count = %d, want 1", len(actions))
	}
}

func TestGetGovernanceAction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/governance/actions/act-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "act-1", "entity_id": "svc-web", "entity_name": "web-app",
			"title": "Fix critical CVE", "description": "Address CVE-2025-1234",
			"priority": "critical", "status": "open",
			"source_type": "security", "source_id": "scan-123",
			"assigned_to": "user-1", "sla_days": 7,
			"created_by": "system", "created_at": "2025-06-01T10:00:00Z",
			"updated_at": "2025-06-01T10:00:00Z",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	action, err := c.GetGovernanceAction(context.Background(), "act-1")
	if err != nil {
		t.Fatalf("GetGovernanceAction() error = %v", err)
	}
	if action.ID != "act-1" {
		t.Errorf("ID = %q, want %q", action.ID, "act-1")
	}
	if action.Title != "Fix critical CVE" {
		t.Errorf("Title = %q, want %q", action.Title, "Fix critical CVE")
	}
	if action.Description != "Address CVE-2025-1234" {
		t.Errorf("Description = %q, want %q", action.Description, "Address CVE-2025-1234")
	}
	if action.Priority != "critical" {
		t.Errorf("Priority = %q, want %q", action.Priority, "critical")
	}
	if action.AssignedTo != "user-1" {
		t.Errorf("AssignedTo = %q, want %q", action.AssignedTo, "user-1")
	}
	if action.SLADays == nil || *action.SLADays != 7 {
		t.Errorf("SLADays = %v, want 7", action.SLADays)
	}
}

func TestGetGovernanceAction_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "NOT_FOUND", "message": "governance action not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetGovernanceAction(context.Background(), "act-nonexistent")
	if err == nil {
		t.Fatal("GetGovernanceAction() expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestCreateGovernanceAction_Success(t *testing.T) {
	// Create returns {"id": "...", "message": "..."} with 201, then we do a GET for the full object
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/governance/actions":
			callCount++
			body, _ := io.ReadAll(r.Body)
			var req CreateGovernanceActionRequest
			json.Unmarshal(body, &req)

			if req.EntityID != "svc-web" {
				t.Errorf("EntityID = %q, want %q", req.EntityID, "svc-web")
			}
			if req.Title != "Fix critical CVE" {
				t.Errorf("Title = %q, want %q", req.Title, "Fix critical CVE")
			}
			if req.Priority != "critical" {
				t.Errorf("Priority = %q, want %q", req.Priority, "critical")
			}
			if req.SourceType != "security" {
				t.Errorf("SourceType = %q, want %q", req.SourceType, "security")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      "act-new-1",
				"message": "Governance action created successfully",
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/governance/actions/act-new-1":
			callCount++
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "act-new-1", "entity_id": "svc-web", "entity_name": "web-app",
				"title": "Fix critical CVE", "description": "Address CVE-2025-1234",
				"priority": "critical", "status": "open",
				"source_type": "security", "source_id": "scan-123",
				"created_at": "2025-06-01T10:00:00Z",
				"updated_at": "2025-06-01T10:00:00Z",
			})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	slaDays := 7
	action, err := c.CreateGovernanceAction(context.Background(), CreateGovernanceActionRequest{
		EntityID:    "svc-web",
		Title:       "Fix critical CVE",
		Description: "Address CVE-2025-1234",
		Priority:    "critical",
		SourceType:  "security",
		SourceID:    "scan-123",
		SLADays:     &slaDays,
	})
	if err != nil {
		t.Fatalf("CreateGovernanceAction() error = %v", err)
	}
	if action.ID != "act-new-1" {
		t.Errorf("ID = %q, want %q", action.ID, "act-new-1")
	}
	if action.Title != "Fix critical CVE" {
		t.Errorf("Title = %q, want %q", action.Title, "Fix critical CVE")
	}
	if action.Status != "open" {
		t.Errorf("Status = %q, want %q", action.Status, "open")
	}
	// Verify both POST and GET were called
	if callCount != 2 {
		t.Errorf("expected 2 API calls (POST + GET), got %d", callCount)
	}
}

func TestUpdateGovernanceAction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/governance/actions/act-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		// Verify only changed fields are sent
		if req["status"] != "resolved" {
			t.Errorf("status = %v, want %q", req["status"], "resolved")
		}
		if req["resolution_note"] != "Fixed in PR #42" {
			t.Errorf("resolution_note = %v, want %q", req["resolution_note"], "Fixed in PR #42")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Governance action updated successfully",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	status := "resolved"
	note := "Fixed in PR #42"
	err := c.UpdateGovernanceAction(context.Background(), "act-1", UpdateGovernanceActionRequest{
		Status:         &status,
		ResolutionNote: &note,
	})
	if err != nil {
		t.Fatalf("UpdateGovernanceAction() error = %v", err)
	}
}

func TestDeleteGovernanceAction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/governance/actions/act-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteGovernanceAction(context.Background(), "act-1")
	if err != nil {
		t.Fatalf("DeleteGovernanceAction() error = %v", err)
	}
}

func TestCreateGovernanceAction_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "",
			"message": "Governance action created successfully",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.CreateGovernanceAction(context.Background(), CreateGovernanceActionRequest{
		EntityID:   "svc-web",
		Title:      "Fix CVE",
		Priority:   "critical",
		SourceType: "security",
	})
	if err == nil {
		t.Fatal("CreateGovernanceAction() expected error for empty ID, got nil")
	}
}

func TestListGovernanceActions_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"actions": []map[string]interface{}{},
			"total":   0,
			"summary": map[string]interface{}{},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	actions, total, err := c.ListGovernanceActions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListGovernanceActions() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(actions) != 0 {
		t.Errorf("action count = %d, want 0", len(actions))
	}
}

func TestGetGovernanceAction_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"code":"INTERNAL","message":"internal server error"}`)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.GetGovernanceAction(context.Background(), "act-1")
	if err == nil {
		t.Fatal("GetGovernanceAction() expected error for 500 response, got nil")
	}
}

func TestDeleteGovernanceAction_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": "NOT_FOUND", "message": "governance action not found",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.DeleteGovernanceAction(context.Background(), "act-nonexistent")
	if err == nil {
		t.Fatal("DeleteGovernanceAction() expected error for 404, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestGovernanceAction_Lifecycle_Integration(t *testing.T) {
	actions := make(map[string]map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// CREATE
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/governance/actions":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			id := "act-lifecycle-1"
			actions[id] = map[string]interface{}{
				"id":          id,
				"entity_id":   req["entity_id"],
				"entity_name": req["entity_name"],
				"title":       req["title"],
				"description": req["description"],
				"priority":    req["priority"],
				"status":      "open",
				"source_type": req["source_type"],
				"source_id":   req["source_id"],
				"created_at":  "2025-06-01T10:00:00Z",
				"updated_at":  "2025-06-01T10:00:00Z",
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      id,
				"message": "Governance action created successfully",
			})

		// GET single
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/governance/actions/act-lifecycle-1":
			if a, ok := actions["act-lifecycle-1"]; ok {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(a)
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "NOT_FOUND", "message": "governance action not found",
				})
			}

		// LIST
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/governance/actions":
			actionList := make([]map[string]interface{}, 0, len(actions))
			for _, a := range actions {
				actionList = append(actionList, a)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"actions": actionList,
				"total":   len(actionList),
				"summary": map[string]interface{}{},
			})

		// UPDATE
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/governance/actions/act-lifecycle-1":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)

			if a, ok := actions["act-lifecycle-1"]; ok {
				if s, ok := req["status"].(string); ok {
					a["status"] = s
				}
				if p, ok := req["priority"].(string); ok {
					a["priority"] = p
				}
				if rn, ok := req["resolution_note"].(string); ok {
					a["resolution_note"] = rn
				}
				a["updated_at"] = "2025-06-01T11:00:00Z"
				actions["act-lifecycle-1"] = a

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"message": "Governance action updated successfully",
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code": "NOT_FOUND", "message": "governance action not found",
				})
			}

		// DELETE
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/governance/actions/act-lifecycle-1":
			delete(actions, "act-lifecycle-1")
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code":"NOT_FOUND","message":"not found"}`)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	// CREATE
	created, err := c.CreateGovernanceAction(context.Background(), CreateGovernanceActionRequest{
		EntityID:    "svc-web",
		EntityName:  "web-app",
		Title:       "Fix critical CVE",
		Description: "Address CVE-2025-1234",
		Priority:    "critical",
		SourceType:  "security",
		SourceID:    "scan-123",
	})
	if err != nil {
		t.Fatalf("CREATE failed: %v", err)
	}
	if created.ID != "act-lifecycle-1" {
		t.Errorf("CREATE: ID = %q, want %q", created.ID, "act-lifecycle-1")
	}
	if created.Status != "open" {
		t.Errorf("CREATE: Status = %q, want %q", created.Status, "open")
	}

	// READ
	read, err := c.GetGovernanceAction(context.Background(), "act-lifecycle-1")
	if err != nil {
		t.Fatalf("READ failed: %v", err)
	}
	if read.Title != "Fix critical CVE" {
		t.Errorf("READ: Title = %q, want %q", read.Title, "Fix critical CVE")
	}
	if read.EntityID != "svc-web" {
		t.Errorf("READ: EntityID = %q, want %q", read.EntityID, "svc-web")
	}

	// LIST
	list, total, err := c.ListGovernanceActions(context.Background(), nil)
	if err != nil {
		t.Fatalf("LIST failed: %v", err)
	}
	if total != 1 {
		t.Errorf("LIST: total = %d, want 1", total)
	}
	if len(list) != 1 {
		t.Errorf("LIST: count = %d, want 1", len(list))
	}

	// UPDATE
	status := "resolved"
	note := "Fixed in PR #42"
	err = c.UpdateGovernanceAction(context.Background(), "act-lifecycle-1", UpdateGovernanceActionRequest{
		Status:         &status,
		ResolutionNote: &note,
	})
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}

	// Verify update via GET
	updated, err := c.GetGovernanceAction(context.Background(), "act-lifecycle-1")
	if err != nil {
		t.Fatalf("GET after UPDATE failed: %v", err)
	}
	if updated.Status != "resolved" {
		t.Errorf("UPDATE: Status = %q, want %q", updated.Status, "resolved")
	}
	if updated.ResolutionNote != "Fixed in PR #42" {
		t.Errorf("UPDATE: ResolutionNote = %q, want %q", updated.ResolutionNote, "Fixed in PR #42")
	}

	// DELETE
	err = c.DeleteGovernanceAction(context.Background(), "act-lifecycle-1")
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	// Verify deletion via GET
	_, err = c.GetGovernanceAction(context.Background(), "act-lifecycle-1")
	if err == nil {
		t.Fatal("GET after DELETE: expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("GET after DELETE: expected not-found error, got: %v", err)
	}
}
