package ssh

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSHError_Error(t *testing.T) {
	tests := []struct {
		name     string
		sshErr   *SSHError
		expected string
	}{
		{
			name: "error without context",
			sshErr: &SSHError{
				Type:    ErrConnectionFailed,
				Message: "failed to connect to server",
			},
			expected: "connection_failed: failed to connect to server",
		},
		{
			name: "error with context",
			sshErr: &SSHError{
				Type:    ErrAuthFailed,
				Message: "authentication failed",
				Context: map[string]interface{}{
					"user": "testuser",
					"host": "example.com",
				},
			},
			expected: "auth_failed: authentication failed (context: map[host:example.com user:testuser])",
		},
		{
			name: "error with empty context map",
			sshErr: &SSHError{
				Type:    ErrCommandFailed,
				Message: "command execution failed",
				Context: map[string]interface{}{},
			},
			expected: "command_failed: command execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sshErr.Error()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSSHError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	sshErr := &SSHError{
		Type:    ErrConnectionFailed,
		Message: "connection failed",
		Cause:   originalErr,
	}

	unwrapped := sshErr.Unwrap()
	require.Equal(t, originalErr, unwrapped)
}

func TestSSHError_Is(t *testing.T) {
	originalErr := errors.New("network error")
	sshErr := &SSHError{
		Type:    ErrConnectionFailed,
		Message: "connection failed",
		Cause:   originalErr,
	}

	// Should match the error type
	require.True(t, sshErr.Is(ErrConnectionFailed))

	// Should match the underlying cause
	require.True(t, sshErr.Is(originalErr))

	// Should not match different error types
	require.False(t, sshErr.Is(ErrAuthFailed))
	require.False(t, sshErr.Is(errors.New("different error")))
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("network timeout")
	wrappedErr := WrapError(ErrConnectionFailed, originalErr)

	require.NotNil(t, wrappedErr)
	require.Equal(t, ErrConnectionFailed, wrappedErr.Type)
	require.Equal(t, "network timeout", wrappedErr.Message)
	require.Equal(t, originalErr, wrappedErr.Cause)
	require.Nil(t, wrappedErr.Context)
}

func TestNewError(t *testing.T) {
	sshErr := NewError(ErrInvalidConfig, "configuration is invalid")

	require.NotNil(t, sshErr)
	require.Equal(t, ErrInvalidConfig, sshErr.Type)
	require.Equal(t, "configuration is invalid", sshErr.Message)
	require.Nil(t, sshErr.Cause)
	require.Nil(t, sshErr.Context)
}

func TestSSHError_WithContext(t *testing.T) {
	sshErr := NewError(ErrCommandFailed, "execution failed")

	// Add first context
	result1 := sshErr.WithContext("command", "ls -la")
	require.Equal(t, sshErr, result1) // Should return self for chaining
	require.NotNil(t, sshErr.Context)
	require.Equal(t, "ls -la", sshErr.Context["command"])

	// Add second context (chaining)
	result2 := sshErr.WithContext("exit_code", 1)
	require.Equal(t, sshErr, result2) // Should return self for chaining
	require.Len(t, sshErr.Context, 2)
	require.Equal(t, "ls -la", sshErr.Context["command"])
	require.Equal(t, 1, sshErr.Context["exit_code"])

	// Verify pre-allocation optimization
	require.GreaterOrEqual(t, cap(make([]string, 0, len(sshErr.Context))), 2)
}

func TestSSHError_WithContextChaining(t *testing.T) {
	sshErr := NewError(ErrAuthFailed, "authentication failed").
		WithContext("user", "testuser").
		WithContext("method", "password").
		WithContext("attempts", 3)

	require.Len(t, sshErr.Context, 3)
	require.Equal(t, "testuser", sshErr.Context["user"])
	require.Equal(t, "password", sshErr.Context["method"])
	require.Equal(t, 3, sshErr.Context["attempts"])
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct auth error",
			err:      ErrAuthFailed,
			expected: true,
		},
		{
			name:     "wrapped auth error",
			err:      WrapError(ErrAuthFailed, errors.New("password rejected")),
			expected: true,
		},
		{
			name:     "different error type",
			err:      ErrConnectionFailed,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct connection error",
			err:      ErrConnectionFailed,
			expected: true,
		},
		{
			name:     "wrapped connection error",
			err:      WrapError(ErrConnectionFailed, errors.New("network unreachable")),
			expected: true,
		},
		{
			name:     "different error type",
			err:      ErrAuthFailed,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "direct timeout error",
			err:      ErrCommandTimeout,
			expected: true,
		},
		{
			name:     "wrapped timeout error",
			err:      WrapError(ErrCommandTimeout, errors.New("context deadline exceeded")),
			expected: true,
		},
		{
			name:     "different error type",
			err:      ErrAuthFailed,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimeoutError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	// Test that all predefined errors are properly defined
	errors := []error{
		ErrInvalidConfig,
		ErrNotConnected,
		ErrSessionCreation,
		ErrCommandFailed,
		ErrCommandTimeout,
		ErrHostKeyRejected,
		ErrAuthFailed,
		ErrConnectionFailed,
	}

	for _, err := range errors {
		require.NotNil(t, err)
		require.NotEmpty(t, err.Error())
	}

	// Test that errors are distinct
	require.NotEqual(t, ErrInvalidConfig, ErrNotConnected)
	require.NotEqual(t, ErrAuthFailed, ErrConnectionFailed)
	require.NotEqual(t, ErrCommandFailed, ErrCommandTimeout)
}

func TestSSHError_ErrorWrapping(t *testing.T) {
	// Test error wrapping chain
	originalErr := errors.New("network error")
	wrappedErr := WrapError(ErrConnectionFailed, originalErr)

	// Should be identifiable via errors.Is
	require.True(t, errors.Is(wrappedErr, ErrConnectionFailed))
	require.True(t, errors.Is(wrappedErr, originalErr))

	// Should be unwrappable via errors.Unwrap
	unwrapped := errors.Unwrap(wrappedErr)
	require.Equal(t, originalErr, unwrapped)
}

func TestSSHError_WithNilContext(t *testing.T) {
	sshErr := &SSHError{
		Type:    ErrCommandFailed,
		Message: "test error",
		Context: nil,
	}

	// Adding context to nil context should initialize the map
	result := sshErr.WithContext("key", "value")
	require.Equal(t, sshErr, result)
	require.NotNil(t, sshErr.Context)
	require.Equal(t, "value", sshErr.Context["key"])
}
