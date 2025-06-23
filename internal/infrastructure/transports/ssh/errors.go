// internal/transports/ssh/errors.go - Error types and handling for SSH operations
package ssh

import (
	"errors"
	"fmt"
)

// Predefined error types for SSH operations using lightweight error values.
//
// These errors provide a standardized set of error conditions that can occur
// during SSH operations, enabling consistent error handling and classification.
var (
	// ErrInvalidConfig indicates that the SSH configuration is invalid
	ErrInvalidConfig = errors.New("invalid_config")
	// ErrNotConnected indicates that no SSH connection is established
	ErrNotConnected = errors.New("not_connected")
	// ErrSessionCreation indicates that SSH session creation failed
	ErrSessionCreation = errors.New("session_creation_failed")
	// ErrCommandFailed indicates that a command execution failed
	ErrCommandFailed = errors.New("command_failed")
	// ErrCommandTimeout indicates that a command execution timed out
	ErrCommandTimeout = errors.New("command_timeout")
	// ErrHostKeyRejected indicates that the host key was rejected or invalid
	ErrHostKeyRejected = errors.New("host_key_rejected")
	// ErrAuthFailed indicates that SSH authentication failed
	ErrAuthFailed = errors.New("auth_failed")
	// ErrConnectionFailed indicates that the SSH connection establishment failed
	ErrConnectionFailed = errors.New("connection_failed")
)

// SSHError provides detailed error information with context and cause chaining.
//
// SSHError implements a rich error type that supports error wrapping, context
// information, and type classification for comprehensive error handling.
type SSHError struct {
	// Type categorizes the error for programmatic handling
	Type error
	// Message provides a human-readable description of the error
	Message string
	// Cause contains the underlying error that caused this error
	Cause error
	// Context provides additional key-value pairs for debugging and logging
	Context map[string]interface{} // Direct access for performance
}

// Error implements the error interface with formatted output.
//
// Error formats the error message including type, message, and context
// information for comprehensive error reporting.
//
// Returns:
//   - Formatted error string with type, message, and context
func (e *SSHError) Error() string {
	if len(e.Context) > 0 {
		return fmt.Sprintf("%v: %s (context: %+v)", e.Type, e.Message, e.Context)
	}
	return fmt.Sprintf("%v: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error for error chain traversal.
//
// Unwrap enables Go's error unwrapping functionality, allowing
// errors.Is and errors.As to work with the error chain.
//
// Returns:
//   - The underlying cause error
func (e *SSHError) Unwrap() error {
	return e.Cause
}

// Is supports error comparison for error type checking.
//
// Is enables comparison with both the error type and the underlying
// cause, supporting Go's standard error checking patterns.
//
// Parameters:
//   - target: Error to compare against
//
// Returns:
//   - True if this error matches the target type or cause
func (e *SSHError) Is(target error) bool {
	return errors.Is(e.Type, target) || errors.Is(e.Cause, target)
}

// WrapError creates a new SSHError wrapping the given error.
//
// WrapError constructs an SSHError that wraps an existing error
// with a specific error type classification.
//
// Parameters:
//   - errType: Error type for classification
//   - cause: Underlying error to wrap
//
// Returns:
//   - SSHError instance wrapping the cause
func WrapError(errType error, cause error) *SSHError {
	return &SSHError{
		Type:    errType,
		Message: cause.Error(),
		Cause:   cause,
	}
}

// NewError creates a new SSHError with a custom message.
//
// NewError constructs an SSHError with a specific type and message
// without wrapping an underlying error.
//
// Parameters:
//   - errType: Error type for classification
//   - message: Human-readable error message
//
// Returns:
//   - SSHError instance with the specified type and message
func NewError(errType error, message string) *SSHError {
	return &SSHError{
		Type:    errType,
		Message: message,
	}
}

// WithContext adds context information to the error.
//
// WithContext enables method chaining by adding key-value context
// information to the error for enhanced debugging and logging.
//
// Parameters:
//   - key: Context key identifier
//   - value: Context value (any type)
//
// Returns:
//   - Self reference for method chaining
func (e *SSHError) WithContext(key string, value interface{}) *SSHError {
	if e.Context == nil {
		e.Context = make(map[string]interface{}, 2) // Pre-allocate small map
	}
	e.Context[key] = value
	return e
}

// Error checking helpers for improved performance and convenience

// IsAuthError checks if the error is an authentication error.
//
// IsAuthError provides a convenient way to check for authentication
// failures without manual error type comparison.
//
// Parameters:
//   - err: Error to check
//
// Returns:
//   - True if the error is an authentication error
func IsAuthError(err error) bool {
	return errors.Is(err, ErrAuthFailed)
}

// IsConnectionError checks if the error is a connection error.
//
// IsConnectionError provides a convenient way to check for connection
// failures without manual error type comparison.
//
// Parameters:
//   - err: Error to check
//
// Returns:
//   - True if the error is a connection error
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

// IsTimeoutError checks if the error is a timeout error.
//
// IsTimeoutError provides a convenient way to check for timeout
// conditions without manual error type comparison.
//
// Parameters:
//   - err: Error to check
//
// Returns:
//   - True if the error is a timeout error
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrCommandTimeout)
}
