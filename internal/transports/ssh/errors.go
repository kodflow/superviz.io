// internal/transports/ssh/errors.go - Error types and handling
package ssh

import (
	"errors"
	"fmt"
)

// Error types for SSH operations
var (
	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig = &ErrorType{name: "invalid_config"}

	// ErrNotConnected indicates no active connection
	ErrNotConnected = &ErrorType{name: "not_connected"}

	// ErrSessionCreation indicates session creation failure
	ErrSessionCreation = &ErrorType{name: "session_creation_failed"}

	// ErrCommandFailed indicates command execution failure
	ErrCommandFailed = &ErrorType{name: "command_failed"}

	// ErrCommandTimeout indicates command execution timeout
	ErrCommandTimeout = &ErrorType{name: "command_timeout"}

	// ErrHostKeyRejected indicates host key verification failure
	ErrHostKeyRejected = &ErrorType{name: "host_key_rejected"}

	// ErrAuthFailed indicates authentication failure
	ErrAuthFailed = &ErrorType{name: "auth_failed"}

	// ErrConnectionFailed indicates connection failure
	ErrConnectionFailed = &ErrorType{name: "connection_failed"}
)

// ErrorType represents a category of SSH errors
type ErrorType struct {
	name string
}

// Error implements the error interface
func (e *ErrorType) Error() string {
	return e.name
}

// WithMessage creates a new SSHError with the given message
func (e *ErrorType) WithMessage(message string) *SSHError {
	return &SSHError{
		Type:    e,
		Message: message,
		context: make(map[string]interface{}),
	}
}

// Wrap creates a new SSHError wrapping the given error
func (e *ErrorType) Wrap(err error) *SSHError {
	return &SSHError{
		Type:    e,
		Cause:   err,
		Message: err.Error(),
		context: make(map[string]interface{}),
	}
}

// SSHError provides detailed error information with context
type SSHError struct {
	Type    *ErrorType             // The type of error
	Message string                 // Human-readable message
	Cause   error                  // Wrapped error, if any
	context map[string]interface{} // Additional context
}

// Error implements the error interface
func (e *SSHError) Error() string {
	if len(e.context) > 0 {
		return fmt.Sprintf("%s: %s (context: %+v)", e.Type.name, e.Message, e.context)
	}
	return fmt.Sprintf("%s: %s", e.Type.name, e.Message)
}

// Unwrap returns the wrapped error for Go 1.13+ error handling
func (e *SSHError) Unwrap() error {
	return e.Cause
}

// Is supports Go 1.13+ error comparison
func (e *SSHError) Is(target error) bool {
	// Check if target is an ErrorType
	if errType, ok := target.(*ErrorType); ok {
		return e.Type == errType
	}

	// Check if target is an SSHError with the same type
	if sshErr, ok := target.(*SSHError); ok {
		return e.Type == sshErr.Type
	}

	// Check wrapped error
	return errors.Is(e.Cause, target)
}

// WithMessage adds or updates the error message
func (e *SSHError) WithMessage(message string) *SSHError {
	e.Message = message
	return e
}

// WithContext adds context information to the error
func (e *SSHError) WithContext(key string, value interface{}) *SSHError {
	if e.context == nil {
		e.context = make(map[string]interface{})
	}
	e.context[key] = value
	return e
}

// Wrap wraps another error
func (e *SSHError) Wrap(err error) *SSHError {
	e.Cause = err
	if e.Message == "" {
		e.Message = err.Error()
	}
	return e
}

// GetContext returns the value for a context key
func (e *SSHError) GetContext(key string) (interface{}, bool) {
	if e.context == nil {
		return nil, false
	}
	val, ok := e.context[key]
	return val, ok
}

// IsAuthError checks if the error is an authentication error
func IsAuthError(err error) bool {
	return errors.Is(err, ErrAuthFailed)
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrCommandTimeout)
}

// NewSSHError creates a new SSH error with formatting (for backward compatibility)
func NewSSHError(errType *ErrorType, format string, args ...interface{}) *SSHError {
	// Extract cause if last argument is an error
	var cause error
	if len(args) > 0 {
		if err, ok := args[len(args)-1].(error); ok {
			cause = err
			args = args[:len(args)-1]
		}
	}

	// Format message
	message := fmt.Sprintf(format, args...)

	// Create error
	err := &SSHError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		context: make(map[string]interface{}),
	}

	return err
}
