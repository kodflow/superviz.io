// internal/transports/ssh/config.go - SSH configuration
package ssh

import (
	"errors"
	"fmt"
	"time"
)

// Config contains SSH connection configuration
type Config struct {
	Host             string
	User             string
	Port             int
	KeyPath          string
	Timeout          time.Duration
	SkipHostKeyCheck bool
	AcceptNewHostKey bool
	address          string // Cached address
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Port:    22,
		Timeout: 30 * time.Second,
	}
}

// Validate ensures the configuration is valid
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

// Address returns the formatted network address
func (c *Config) Address() string {
	if c.address == "" {
		c.address = fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
	return c.address
}
