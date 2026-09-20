package client

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("resource not found")

// errorEnvelope is Shoehorn's error shape: code and message sit under "error".
type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// parseAPIError reads Shoehorn's envelope, then a flat {"code","message"}, then the raw body.
func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && (env.Error.Code != "" || env.Error.Message != "") {
		apiErr.Code = env.Error.Code
		apiErr.Message = env.Error.Message
		return apiErr
	}

	var flat struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &flat); err == nil {
		apiErr.Code = flat.Code
		apiErr.Message = flat.Message
	}

	if apiErr.Code == "" && apiErr.Message == "" {
		if len(body) > 0 {
			apiErr.Message = string(body)
		} else {
			apiErr.Message = http.StatusText(statusCode)
		}
	}
	return apiErr
}

// IsAlreadyExists returns true if the error indicates a resource already exists (HTTP 409).
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

// IsNotFound returns true if the error indicates a resource was not found.
// It unwraps error chains, so it works with errors wrapped via fmt.Errorf %w.
// It checks for the ErrNotFound sentinel (used by list-and-filter methods)
// and for *APIError with a 404 status code (returned by direct HTTP endpoints).
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}
