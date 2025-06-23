// internal/transports/ssh/errors.go - Error types and handling
package ssh

import (
	"errors"
	"fmt"
)

// Error types for SSH operations - using lightweight error values
var (
	ErrInvalidConfig    = errors.New("invalid_config")
	ErrNotConnected     = errors.New("not_connected")
	ErrSessionCreation  = errors.New("session_creation_failed")
	ErrCommandFailed    = errors.New("command_failed")
	ErrCommandTimeout   = errors.New("command_timeout")
	ErrHostKeyRejected  = errors.New("host_key_rejected")
	ErrAuthFailed       = errors.New("auth_failed")
	ErrConnectionFailed = errors.New("connection_failed")
)

// SSHError provides detailed error information with context
type SSHError struct {
	Type    error
	Message string
	Cause   error
	Context map[string]interface{} // Direct access for performance
}

// Error implements the error interface
func (e *SSHError) Error() string {
	if len(e.Context) > 0 {
		return fmt.Sprintf("%v: %s (context: %+v)", e.Type, e.Message, e.Context)
	}
	return fmt.Sprintf("%v: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *SSHError) Unwrap() error {
	return e.Cause
}

// Is supports error comparison
func (e *SSHError) Is(target error) bool {
	return errors.Is(e.Type, target) || errors.Is(e.Cause, target)
}

// WrapError creates a new SSHError wrapping the given error
func WrapError(errType error, cause error) *SSHError {
	return &SSHError{
		Type:    errType,
		Message: cause.Error(),
		Cause:   cause,
	}
}

// NewError creates a new SSHError with message
func NewError(errType error, message string) *SSHError {
	return &SSHError{
		Type:    errType,
		Message: message,
	}
}

// WithContext adds context to error (returns self for chaining)
func (e *SSHError) WithContext(key string, value interface{}) *SSHError {
	if e.Context == nil {
		e.Context = make(map[string]interface{}, 2) // Pre-allocate small map
	}
	e.Context[key] = value
	return e
}

// Error checking helpers for better performance

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
