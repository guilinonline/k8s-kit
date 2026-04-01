package pod

import (
	"errors"
	"strings"
)

// Common error messages
var (
	ErrContainerNotFound = errors.New("container not found in pod")
	ErrPodNotFound       = errors.New("pod not found")
	ErrForbidden         = errors.New("forbidden")
	ErrTimeout           = errors.New("operation timed out")
	ErrConnectionLost    = errors.New("connection lost")
)

// IsContainerNotFound checks if the error indicates container not found.
func IsContainerNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "container") && strings.Contains(err.Error(), "not found")
}

// IsPodNotFound checks if the error indicates pod not found.
func IsPodNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "pods") && strings.Contains(err.Error(), "not found")
}

// IsForbidden checks if the error indicates permission denied.
func IsForbidden(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "403")
}

// IsTimeout checks if the error indicates timeout.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded")
}

// IsConnectionLost checks if the error indicates connection lost.
func IsConnectionLost(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connection") && (strings.Contains(err.Error(), "lost") || strings.Contains(err.Error(), "broken"))
}

// IsNotFound is a convenience function that checks for any not found error.
func IsNotFound(err error) bool {
	return IsPodNotFound(err) || IsContainerNotFound(err)
}

// IsServerError checks if the error is a server-side error (5xx).
func IsServerError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "502") ||
		strings.Contains(err.Error(), "503") || strings.Contains(err.Error(), "504")
}