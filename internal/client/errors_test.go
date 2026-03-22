package client

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFound_WithErrNotFound(t *testing.T) {
	t.Parallel()
	if !IsNotFound(ErrNotFound) {
		t.Error("IsNotFound(ErrNotFound) = false, want true")
	}
}

func TestIsNotFound_WithWrappedErrNotFound(t *testing.T) {
	t.Parallel()
	wrapped := fmt.Errorf("get api key: %w", ErrNotFound)
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound(wrapped ErrNotFound) = false, want true")
	}
}

func TestIsNotFound_WithAPIError404(t *testing.T) {
	t.Parallel()
	apiErr := &APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "not found"}
	if !IsNotFound(apiErr) {
		t.Error("IsNotFound(APIError{404}) = false, want true")
	}
}

func TestIsNotFound_WithWrappedAPIError404(t *testing.T) {
	t.Parallel()
	apiErr := &APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "not found"}
	wrapped := fmt.Errorf("get integration: %w", apiErr)
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound(wrapped APIError{404}) = false, want true")
	}
}

func TestIsNotFound_WithAPIError500(t *testing.T) {
	t.Parallel()
	apiErr := &APIError{StatusCode: 500, Message: "server error"}
	if IsNotFound(apiErr) {
		t.Error("IsNotFound(APIError{500}) = true, want false")
	}
}

func TestIsNotFound_WithRegularError(t *testing.T) {
	t.Parallel()
	err := errors.New("connection refused")
	if IsNotFound(err) {
		t.Error("IsNotFound(regular error) = true, want false")
	}
}

func TestIsNotFound_WithNil(t *testing.T) {
	t.Parallel()
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true, want false")
	}
}

func TestGetAPIKey_NotFound_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	_, err := setupClientWithEmptyKeys(t).GetAPIKey(testCtx(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("GetAPIKey not-found error should satisfy IsNotFound, got: %v", err)
	}
}

func TestGetFeatureFlag_NotFound_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	_, err := setupClientWithEmptyFlags(t).GetFeatureFlag(testCtx(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("GetFeatureFlag not-found error should satisfy IsNotFound, got: %v", err)
	}
}

func TestGetPolicy_NotFound_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	_, err := setupClientWithEmptyPolicies(t).GetPolicy(testCtx(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("GetPolicy not-found error should satisfy IsNotFound, got: %v", err)
	}
}

func TestGetUserRole_NotFound_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	_, err := setupClientWithEmptyRoles(t).GetUserRole(testCtx(), "user-1", "admin")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("GetUserRole not-found error should satisfy IsNotFound, got: %v", err)
	}
}

// TestIsNotFound_HTTP404ViaAPIError verifies the full path: HTTP 404 response ->
// APIError wrapping in doRequest -> fmt.Errorf wrapping in client method -> IsNotFound.
// This covers the direct-endpoint Get methods (GetTeam, GetIntegration, etc.)
// where 404 comes from the HTTP layer, not from list-and-filter logic.
func TestIsNotFound_HTTP404ViaAPIError(t *testing.T) {
	t.Parallel()
	c := setupClientWith404Server(t)

	// Simulate a direct-endpoint call that returns 404
	_, err := c.Get(testCtx(), "/api/v1/admin/teams/nonexistent")
	if err == nil {
		t.Fatal("expected error from 404 response, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("HTTP 404 error should satisfy IsNotFound, got: %v", err)
	}

	// Verify a wrapped 404 also works (simulating client method wrapping)
	wrapped := fmt.Errorf("get team: %w", err)
	if !IsNotFound(wrapped) {
		t.Errorf("wrapped HTTP 404 error should satisfy IsNotFound, got: %v", wrapped)
	}
}
