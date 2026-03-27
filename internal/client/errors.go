package client

import (
	"errors"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("resource not found")

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
