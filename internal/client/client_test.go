package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient_SetsFieldsCorrectly(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		apiKey     string
		timeout    time.Duration
		wantURL    string
		wantAPIKey string
	}{
		{
			name:       "basic configuration",
			baseURL:    "https://shoehorn.example.com",
			apiKey:     "shp_svc_test123",
			timeout:    30 * time.Second,
			wantURL:    "https://shoehorn.example.com",
			wantAPIKey: "shp_svc_test123",
		},
		{
			name:       "trailing slash removed",
			baseURL:    "https://shoehorn.example.com/",
			apiKey:     "shp_svc_abc",
			timeout:    60 * time.Second,
			wantURL:    "https://shoehorn.example.com",
			wantAPIKey: "shp_svc_abc",
		},
		{
			name:       "multiple trailing slashes removed",
			baseURL:    "https://shoehorn.example.com///",
			apiKey:     "key",
			timeout:    10 * time.Second,
			wantURL:    "https://shoehorn.example.com",
			wantAPIKey: "key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.baseURL, tt.apiKey, tt.timeout)
			if c.BaseURL != tt.wantURL {
				t.Errorf("BaseURL = %q, want %q", c.BaseURL, tt.wantURL)
			}
			if c.APIKey != tt.wantAPIKey {
				t.Errorf("APIKey = %q, want %q", c.APIKey, tt.wantAPIKey)
			}
			if c.HTTPClient == nil {
				t.Error("HTTPClient is nil")
			}
			if c.HTTPClient.Timeout != tt.timeout {
				t.Errorf("Timeout = %v, want %v", c.HTTPClient.Timeout, tt.timeout)
			}
			if c.UserAgent != "terraform-provider-shoehorn" {
				t.Errorf("UserAgent = %q, want %q", c.UserAgent, "terraform-provider-shoehorn")
			}
		})
	}
}

func TestClient_Get_SetsAuthHeaders(t *testing.T) {
	var gotHeaders http.Header
	var gotMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "shp_svc_test_token", 30*time.Second)
	_, err := c.Get(context.Background(), "/api/v1/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if got := gotHeaders.Get("Authorization"); got != "Bearer shp_svc_test_token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer shp_svc_test_token")
	}
	if got := gotHeaders.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	if got := gotHeaders.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want %q", got, "application/json")
	}
	if got := gotHeaders.Get("User-Agent"); got != "terraform-provider-shoehorn" {
		t.Errorf("User-Agent = %q, want %q", got, "terraform-provider-shoehorn")
	}
}

func TestClient_Post_SendsBody(t *testing.T) {
	var gotBody map[string]interface{}
	var gotMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	reqBody := map[string]string{"name": "test-team", "slug": "test"}
	resp, err := c.Post(context.Background(), "/api/v1/admin/teams", reqBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["name"] != "test-team" {
		t.Errorf("body name = %v, want %q", gotBody["name"], "test-team")
	}
	if gotBody["slug"] != "test" {
		t.Errorf("body slug = %v, want %q", gotBody["slug"], "test")
	}

	var respData map[string]string
	if err := json.Unmarshal(resp, &respData); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if respData["id"] != "123" {
		t.Errorf("response id = %q, want %q", respData["id"], "123")
	}
}

func TestClient_Put_SendsBody(t *testing.T) {
	var gotMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"updated":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.Put(context.Background(), "/api/v1/admin/teams/123", map[string]string{"name": "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
}

func TestClient_Delete_SendsRequest(t *testing.T) {
	var gotMethod string
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	err := c.Delete(context.Background(), "/api/v1/admin/teams/456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/v1/admin/teams/456" {
		t.Errorf("path = %q, want %q", gotPath, "/api/v1/admin/teams/456")
	}
}

func TestClient_ErrorResponse_4xx(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "400 bad request with JSON error",
			statusCode: http.StatusBadRequest,
			body:       `{"code":"INVALID_INPUT","message":"Name is required"}`,
			wantCode:   "INVALID_INPUT",
			wantMsg:    "Name is required",
		},
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"code":"UNAUTHORIZED","message":"Invalid API key"}`,
			wantCode:   "UNAUTHORIZED",
			wantMsg:    "Invalid API key",
		},
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       `{"code":"NOT_FOUND","message":"Team not found"}`,
			wantCode:   "NOT_FOUND",
			wantMsg:    "Team not found",
		},
		{
			name:       "500 server error with plain text",
			statusCode: http.StatusInternalServerError,
			body:       `Internal Server Error`,
			wantCode:   "",
			wantMsg:    "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := NewClient(server.URL, "key", 30*time.Second)
			_, err := c.Get(context.Background(), "/api/v1/test")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// 5xx errors are retried and may be wrapped; use errors.As to unwrap
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *APIError (possibly wrapped), got %T: %v", err, err)
			}
			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.wantCode)
			}
			if apiErr.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantMsg)
			}
		})
	}
}

func TestAPIError_ErrorString(t *testing.T) {
	tests := []struct {
		name    string
		err     APIError
		wantStr string
	}{
		{
			name:    "with code",
			err:     APIError{StatusCode: 400, Code: "INVALID_INPUT", Message: "Bad data"},
			wantStr: "shoehorn API error (HTTP 400): INVALID_INPUT - Bad data",
		},
		{
			name:    "without code",
			err:     APIError{StatusCode: 500, Message: "Internal error"},
			wantStr: "shoehorn API error (HTTP 500): Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantStr {
				t.Errorf("Error() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Get(ctx, "/api/v1/test")
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestClient_URLConstruction(t *testing.T) {
	var gotURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.Get(context.Background(), "/api/v1/entities/my-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotURL != "/api/v1/entities/my-service" {
		t.Errorf("URL = %q, want %q", gotURL, "/api/v1/entities/my-service")
	}
}

func TestClient_NilBody_NoContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("expected empty body for GET, got %q", string(body))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Retry_On5xx(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if int(n) < maxRetries {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`Bad Gateway`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	resp, err := c.Get(context.Background(), "/api/v1/test")
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}

	var result map[string]bool
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result["ok"] {
		t.Error("expected ok=true in response")
	}

	got := int(atomic.LoadInt32(&attempts))
	if got != maxRetries {
		t.Errorf("attempts = %d, want %d", got, maxRetries)
	}
}

func TestClient_Retry_On5xx_ExhaustsRetries(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`Service Unavailable`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 30*time.Second)
	_, err := c.Get(context.Background(), "/api/v1/test")
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}

	got := int(atomic.LoadInt32(&attempts))
	if got != maxRetries {
		t.Errorf("attempts = %d, want %d (maxRetries)", got, maxRetries)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError wrapped in error, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusServiceUnavailable)
	}

	// Verify the wrapping message indicates retries were exhausted
	if !strings.Contains(err.Error(), "request failed after") {
		t.Errorf("error should contain 'request failed after', got: %v", err)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "EOF", err: fmt.Errorf("read: EOF"), want: true},
		{name: "connection reset", err: fmt.Errorf("read: connection reset by peer"), want: true},
		{name: "connection refused", err: fmt.Errorf("dial tcp: connection refused"), want: true},
		{name: "broken pipe", err: fmt.Errorf("write: broken pipe"), want: true},
		{name: "context deadline exceeded", err: fmt.Errorf("context deadline exceeded"), want: true},
		{name: "Client.Timeout", err: fmt.Errorf("net/http: request canceled (Client.Timeout exceeded)"), want: true},
		{name: "context canceled", err: context.Canceled, want: false},
		{name: "unrelated error", err: fmt.Errorf("permission denied"), want: false},
		{name: "API error", err: &APIError{StatusCode: 400, Message: "bad request"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryable(tt.err)
			if got != tt.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
