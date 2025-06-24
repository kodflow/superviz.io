package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPredefinedErrors(t *testing.T) {
	// Test that all predefined errors are properly defined
	errors := []error{
		ErrInvalidTarget,
		ErrNilConfig,
		ErrNilWriter,
	}

	for _, err := range errors {
		require.NotNil(t, err)
		require.NotEmpty(t, err.Error())
	}

	// Test that errors are distinct
	require.NotEqual(t, ErrInvalidTarget, ErrNilConfig)
	require.NotEqual(t, ErrNilConfig, ErrNilWriter)
	require.NotEqual(t, ErrInvalidTarget, ErrNilWriter)
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "invalid target error",
			err:      ErrInvalidTarget,
			expected: "invalid target format, expected user@host",
		},
		{
			name:     "nil config error",
			err:      ErrNilConfig,
			expected: "config cannot be nil",
		},
		{
			name:     "nil writer error",
			err:      ErrNilWriter,
			expected: "writer cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}
