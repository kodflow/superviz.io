// Package services provides predefined error types for service layer operations
package services

import "errors"

// Predefined error variables for common service-layer error conditions.
//
// These errors provide standardized error types that can be used across
// different services for consistent error handling and classification.
var (
	// ErrInvalidTarget indicates that the target format is invalid (expected user@host)
	ErrInvalidTarget = errors.New("invalid target format, expected user@host")
	// ErrNilConfig indicates that a required configuration parameter is nil
	ErrNilConfig = errors.New("config cannot be nil")
	// ErrNilWriter indicates that a required writer parameter is nil
	ErrNilWriter = errors.New("writer cannot be nil")
)
