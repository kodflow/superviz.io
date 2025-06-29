// internal/transports/ssh/config.go - SSH configuration
package ssh

import (
	"errors"
	"fmt"
	"time"
)

// Config contains SSH connection configuration parameters.
//
// Config holds all necessary information to establish and configure an SSH connection,
// including authentication details, timeouts, and host key verification settings.
type Config struct {
	// Host is the hostname or IP address of the SSH server
	Host string
	// User is the username for SSH authentication
	User string
	// Port is the TCP port number for the SSH connection (default: 22)
	Port int
	// KeyPath is the file path to the SSH private key for authentication
	KeyPath string
	// Password is the password for SSH authentication (alternative to KeyPath)
	Password string
	// Timeout is the maximum duration to wait for connection establishment
	Timeout time.Duration
	// SkipHostKeyCheck bypasses host key verification (insecure)
	SkipHostKeyCheck bool
	// AcceptNewHostKey automatically accepts unknown host keys
	AcceptNewHostKey bool
	// address is a cached formatted address string (private field)
	address string // Cached address
}

// DefaultConfig returns a configuration with sensible defaults.
//
// Returns:
//   - Config instance with default values (port 22, 30s timeout)
func DefaultConfig() *Config {
	return &Config{
		Port:    22,
		Timeout: 30 * time.Second,
	}
}

// Validate ensures the configuration is valid and complete.
//
// Validate performs comprehensive validation of all configuration fields
// and pre-computes the network address for efficiency.
//
// Returns:
//   - Error if any configuration parameter is invalid
func (c *Config) Validate() error {
	// Single pass validation with early returns
	if c.Host == "" {
		return errors.New("host cannot be empty")
	}
	if c.User == "" {
		return errors.New("user cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	// Pre-compute address
	c.address = fmt.Sprintf("%s:%d", c.Host, c.Port)
	return nil
}

// Address returns the formatted network address for the SSH connection.
//
// Address combines the host and port into a network address string suitable
// for network dialing operations. The result is cached for performance.
//
// Returns:
//   - Formatted network address as "host:port"
func (c *Config) Address() string {
	if c.address == "" {
		c.address = fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
	return c.address
}
