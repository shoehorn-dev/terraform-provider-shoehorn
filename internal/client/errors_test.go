package client

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestParseAPIError_NamesTheFieldsAValidationErrorRejected(t *testing.T) {
	body := []byte(`{"error":{"code":"VALIDATION_ERROR","message":"Request validation failed","details":[` +
		`{"field":"appearance.platform_name_color_light","message":"must be a hex color, a named CSS color, or an rgb()/rgba()/hsl()/hsla() value"},` +
		`{"field":"appearance.logo_url_dark","message":"must be a valid URL"}]}}`)

	got := parseAPIError(400, body).Error()

	for _, want := range []string{
		"VALIDATION_ERROR - Request validation failed",
		"appearance.platform_name_color_light: must be a hex color",
		"appearance.logo_url_dark: must be a valid URL",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() = %q, want it to contain %q", got, want)
		}
	}
}

func TestParseAPIError_DetailEdgeCasesKeepThePlainMessage(t *testing.T) {
	for name, details := range map[string]string{
		"null":                `null`,
		"empty object entry":  `[{}]`,
		"non-object entries":  `[1, 2]`,
		"entry without field": `[{"message":"orphan"}]`,
	} {
		t.Run(name, func(t *testing.T) {
			body := []byte(`{"error":{"code":"VALIDATION_ERROR","message":"Request validation failed","details":` + details + `}}`)
			got := parseAPIError(400, body).Error()
			if got != "shoehorn API error (HTTP 400): VALIDATION_ERROR - Request validation failed" {
				t.Errorf("Error() = %q, want the plain code and message", got)
			}
		})
	}
}

func TestParseAPIError_IgnoresDetailsThatAreNotFieldErrors(t *testing.T) {
	body := []byte(`{"error":{"code":"CONFLICT","message":"Already exists","details":{"existing_id":"abc"}}}`)

	got := parseAPIError(409, body).Error()

	if got != "shoehorn API error (HTTP 409): CONFLICT - Already exists" {
		t.Errorf("Error() = %q, want the plain code and message", got)
	}
}

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
