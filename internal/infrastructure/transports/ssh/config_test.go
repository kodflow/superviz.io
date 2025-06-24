package ssh

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	require.NotNil(t, config)
	require.Equal(t, 22, config.Port)
	require.Equal(t, 30*time.Second, config.Timeout)
	require.Empty(t, config.Host)
	require.Empty(t, config.User)
	require.Empty(t, config.KeyPath)
	require.False(t, config.SkipHostKeyCheck)
	require.False(t, config.AcceptNewHostKey)
	require.Empty(t, config.address)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Host:    "example.com",
				User:    "testuser",
				Port:    22,
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host:    "",
				User:    "testuser",
				Port:    22,
				Timeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "host cannot be empty",
		},
		{
			name: "empty user",
			config: &Config{
				Host:    "example.com",
				User:    "",
				Port:    22,
				Timeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "user cannot be empty",
		},
		{
			name: "port too low",
			config: &Config{
				Host:    "example.com",
				User:    "testuser",
				Port:    0,
				Timeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "port too high",
			config: &Config{
				Host:    "example.com",
				User:    "testuser",
				Port:    65536,
				Timeout: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "negative timeout",
			config: &Config{
				Host:    "example.com",
				User:    "testuser",
				Port:    22,
				Timeout: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "zero timeout",
			config: &Config{
				Host:    "example.com",
				User:    "testuser",
				Port:    22,
				Timeout: 0,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "valid configuration with custom port",
			config: &Config{
				Host:    "192.168.1.100",
				User:    "admin",
				Port:    2222,
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				// Address should be cached after validation
				require.NotEmpty(t, tt.config.address)
				require.Equal(t, tt.config.Address(), tt.config.address)
			}
		})
	}
}

func TestConfig_Address(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "default port",
			config: &Config{
				Host: "example.com",
				Port: 22,
			},
			expected: "example.com:22",
		},
		{
			name: "custom port",
			config: &Config{
				Host: "192.168.1.100",
				Port: 2222,
			},
			expected: "192.168.1.100:2222",
		},
		{
			name: "high port number",
			config: &Config{
				Host: "localhost",
				Port: 65535,
			},
			expected: "localhost:65535",
		},
		{
			name: "IPv6 address",
			config: &Config{
				Host: "::1",
				Port: 22,
			},
			expected: "::1:22",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test first call (computation)
			address1 := tt.config.Address()
			require.Equal(t, tt.expected, address1)

			// Test second call (cached)
			address2 := tt.config.Address()
			require.Equal(t, tt.expected, address2)
			require.Equal(t, address1, address2)

			// Verify address was cached
			require.Equal(t, tt.expected, tt.config.address)
		})
	}
}

func TestConfig_AddressCaching(t *testing.T) {
	config := &Config{
		Host: "example.com",
		Port: 22,
	}

	// Initially address should be empty
	require.Empty(t, config.address)

	// First call should compute and cache
	addr1 := config.Address()
	require.Equal(t, "example.com:22", addr1)
	require.Equal(t, "example.com:22", config.address)

	// Manually modify the cached value to test caching
	config.address = "cached:value"
	addr2 := config.Address()
	require.Equal(t, "cached:value", addr2)
}

func TestConfig_ValidateAddressCaching(t *testing.T) {
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		Port:    22,
		Timeout: 30 * time.Second,
	}

	// Validate should cache the address
	err := config.Validate()
	require.NoError(t, err)
	require.Equal(t, "example.com:22", config.address)

	// Address() should return the cached value
	addr := config.Address()
	require.Equal(t, "example.com:22", addr)
}
