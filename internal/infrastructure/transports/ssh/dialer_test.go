package ssh

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewDefaultDialer(t *testing.T) {
	dialer := NewDefaultDialer()
	require.NotNil(t, dialer)

	// Should be able to cast to concrete type
	defaultDialer, ok := dialer.(*defaultDialer)
	require.True(t, ok)
	require.NotNil(t, defaultDialer.netDialer)
	require.Equal(t, 30*time.Second, defaultDialer.netDialer.Timeout)
	require.Equal(t, 30*time.Second, defaultDialer.netDialer.KeepAlive)
}

func TestNewDialer(t *testing.T) {
	customNetDialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	dialer := NewDialer(customNetDialer)
	require.NotNil(t, dialer)

	defaultDialer, ok := dialer.(*defaultDialer)
	require.True(t, ok)
	require.Equal(t, customNetDialer, defaultDialer.netDialer)
}

func TestDefaultDialer_WrapError_ContextErrors(t *testing.T) {
	dialer := &defaultDialer{}

	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "context cancelled",
			err:      context.Canceled,
			expected: ErrConnectionFailed,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: ErrConnectionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrappedErr := dialer.wrapError(tt.err, "example.com:22")
			require.True(t, errors.Is(wrappedErr, tt.expected))

			// Check error message
			sshErr, ok := wrappedErr.(*SSHError)
			require.True(t, ok)
			require.Equal(t, "example.com:22", sshErr.Context["address"])
		})
	}
}

func TestDefaultDialer_WrapError_NetworkErrors(t *testing.T) {
	dialer := &defaultDialer{}

	// Mock network timeout error
	mockNetError := &mockNetError{timeout: true}
	wrappedErr := dialer.wrapError(mockNetError, "example.com:22")

	require.True(t, errors.Is(wrappedErr, ErrConnectionFailed))
	require.Contains(t, wrappedErr.Error(), "connection timeout")

	sshErr, ok := wrappedErr.(*SSHError)
	require.True(t, ok)
	require.Equal(t, "example.com:22", sshErr.Context["address"])
}

func TestDefaultDialer_WrapError_AuthErrors(t *testing.T) {
	dialer := &defaultDialer{}

	authErrors := []string{
		"permission denied",
		"unable to authenticate",
		"Permission Denied", // Test case insensitive
	}

	for _, errMsg := range authErrors {
		t.Run(errMsg, func(t *testing.T) {
			err := errors.New(errMsg)
			wrappedErr := dialer.wrapError(err, "example.com:22")

			require.True(t, errors.Is(wrappedErr, ErrAuthFailed))

			sshErr, ok := wrappedErr.(*SSHError)
			require.True(t, ok)
			require.Equal(t, "example.com:22", sshErr.Context["address"])
		})
	}
}

func TestDefaultDialer_WrapError_NetworkConnectionErrors(t *testing.T) {
	dialer := &defaultDialer{}

	netErrors := []string{
		"connection refused",
		"no route to host",
		"host is unreachable",
		"Connection Refused", // Test case insensitive
	}

	for _, errMsg := range netErrors {
		t.Run(errMsg, func(t *testing.T) {
			err := errors.New(errMsg)
			wrappedErr := dialer.wrapError(err, "example.com:22")

			require.True(t, errors.Is(wrappedErr, ErrConnectionFailed))

			sshErr, ok := wrappedErr.(*SSHError)
			require.True(t, ok)
			require.Equal(t, "example.com:22", sshErr.Context["address"])
		})
	}
}

func TestDefaultDialer_WrapError_HostKeyErrors(t *testing.T) {
	dialer := &defaultDialer{}

	hostKeyErrors := []string{
		"host key verification failed",
		"host key mismatch",
		"Host Key rejected", // Test case insensitive
	}

	for _, errMsg := range hostKeyErrors {
		t.Run(errMsg, func(t *testing.T) {
			err := errors.New(errMsg)
			wrappedErr := dialer.wrapError(err, "example.com:22")

			require.True(t, errors.Is(wrappedErr, ErrHostKeyRejected))

			sshErr, ok := wrappedErr.(*SSHError)
			require.True(t, ok)
			require.Equal(t, "example.com:22", sshErr.Context["address"])
		})
	}
}

func TestDefaultDialer_WrapError_DefaultCase(t *testing.T) {
	dialer := &defaultDialer{}

	unknownErr := errors.New("some unknown error")
	wrappedErr := dialer.wrapError(unknownErr, "example.com:22")

	require.True(t, errors.Is(wrappedErr, ErrConnectionFailed))
	require.True(t, errors.Is(wrappedErr, unknownErr))

	sshErr, ok := wrappedErr.(*SSHError)
	require.True(t, ok)
	require.Equal(t, "example.com:22", sshErr.Context["address"])
}

func TestSSHConnection_NewSession_Success(t *testing.T) {
	// Note: We can't easily test the real SSH connection without a real SSH server
	// But we can test the interface compliance
	conn := &sshConnection{}
	require.NotNil(t, conn)

	// Test that methods exist (interface compliance)
	_ = conn.NewSession
	_ = conn.Close
}

func TestSSHSession_Methods(t *testing.T) {
	// Test interface compliance
	session := &sshSession{}
	require.NotNil(t, session)

	// Test that methods exist (interface compliance)
	_ = session.Run
	_ = session.Close
}

// Mock network error for testing
type mockNetError struct {
	timeout   bool
	temporary bool
}

func (e *mockNetError) Error() string {
	return "mock network error"
}

func (e *mockNetError) Timeout() bool {
	return e.timeout
}

func (e *mockNetError) Temporary() bool {
	return e.temporary
}

func TestErrorPatterns(t *testing.T) {
	// Test that pre-compiled error patterns are correct
	require.Contains(t, authErrors, "permission denied")
	require.Contains(t, authErrors, "unable to authenticate")
	require.Contains(t, netErrors, "connection refused")
	require.Contains(t, netErrors, "no route to host")
	require.Contains(t, netErrors, "host is unreachable")
}

func TestDefaultDialer_WrapError_CaseInsensitive(t *testing.T) {
	dialer := &defaultDialer{}

	tests := []struct {
		name        string
		errMsg      string
		expectedErr error
	}{
		{
			name:        "uppercase auth error",
			errMsg:      "PERMISSION DENIED",
			expectedErr: ErrAuthFailed,
		},
		{
			name:        "mixed case network error",
			errMsg:      "Connection Refused",
			expectedErr: ErrConnectionFailed,
		},
		{
			name:        "mixed case host key error",
			errMsg:      "Host Key verification failed",
			expectedErr: ErrHostKeyRejected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			wrappedErr := dialer.wrapError(err, "example.com:22")
			require.True(t, errors.Is(wrappedErr, tt.expectedErr))
		})
	}
}

func TestDefaultDialer_WrapError_PartialMatches(t *testing.T) {
	dialer := &defaultDialer{}

	tests := []struct {
		name        string
		errMsg      string
		expectedErr error
	}{
		{
			name:        "auth error in longer message",
			errMsg:      "SSH authentication failed: permission denied (publickey)",
			expectedErr: ErrAuthFailed,
		},
		{
			name:        "network error in longer message",
			errMsg:      "dial tcp 192.168.1.1:22: connection refused",
			expectedErr: ErrConnectionFailed,
		},
		{
			name:        "host key error in longer message",
			errMsg:      "SSH server host key verification failed",
			expectedErr: ErrHostKeyRejected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			wrappedErr := dialer.wrapError(err, "example.com:22")
			require.True(t, errors.Is(wrappedErr, tt.expectedErr))
		})
	}
}
