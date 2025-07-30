package models

import "errors"

// Common error types for the LLM dispatcher
var (
	ErrNoVendorsRegistered = errors.New("no vendors registered")
	ErrVendorNotFound      = errors.New("vendor not found")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrVendorUnavailable   = errors.New("vendor unavailable")
	ErrTimeout             = errors.New("request timeout")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrInvalidConfig       = errors.New("invalid configuration")
)
