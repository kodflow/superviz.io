// internal/transports/ssh/client.go - Main SSH client implementation
package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config contains SSH connection configuration
type Config struct {
	Host             string
	User             string
	Port             int
	KeyPath          string // Path to private key (optional)
	Timeout          time.Duration
	SkipHostKeyCheck bool // Skip host key verification (development only)
	AcceptNewHostKey bool // Auto-accept new host keys (development only)
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
	if c.Host == "" {
		return ErrInvalidConfig.WithMessage("host cannot be empty")
	}
	if c.User == "" {
		return ErrInvalidConfig.WithMessage("user cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidConfig.WithMessage("port must be between 1 and 65535")
	}
	if c.Timeout <= 0 {
		return ErrInvalidConfig.WithMessage("timeout must be positive")
	}
	return nil
}

// Address returns the formatted network address
func (c *Config) Address() string {
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))
}

// HostKeyCallback is an alias for ssh.HostKeyCallback for better abstraction
type HostKeyCallback = ssh.HostKeyCallback

// Authenticator handles SSH authentication
type Authenticator interface {
	// GetAuthMethods returns the authentication methods based on config
	GetAuthMethods(ctx context.Context, config *Config) ([]ssh.AuthMethod, error)
}

// HostKeyManager handles host key verification
type HostKeyManager interface {
	// GetHostKeyCallback returns the appropriate host key callback
	GetHostKeyCallback(ctx context.Context, config *Config) (HostKeyCallback, error)
}

// Session represents an SSH session
type Session interface {
	// Run executes a command and waits for it to complete
	Run(cmd string) error
	// Close closes the session
	Close() error
}

// Connection represents an SSH connection
type Connection interface {
	// NewSession creates a new session
	NewSession() (Session, error)
	// Close closes the connection
	Close() error
}

// Dialer establishes SSH connections
type Dialer interface {
	// DialContext establishes an SSH connection with context
	DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (Connection, error)
}

// Client provides SSH operations
type Client interface {
	// Connect establishes an SSH connection
	Connect(ctx context.Context, config *Config) error
	// Execute runs a command on the remote host
	Execute(ctx context.Context, command string) error
	// Close closes the SSH connection
	Close() error
}

// client implements the SSH client
type client struct {
	conn           Connection
	config         *Config
	authenticator  Authenticator
	hostKeyManager HostKeyManager
	dialer         Dialer
}

// ClientOptions contains options for creating a new client
type ClientOptions struct {
	Authenticator  Authenticator
	HostKeyManager HostKeyManager
	Dialer         Dialer
}

// NewClient creates a new SSH client with the given options
func NewClient(opts *ClientOptions) Client {
	if opts == nil {
		opts = &ClientOptions{}
	}

	// Use default implementations if not provided
	if opts.Authenticator == nil {
		opts.Authenticator = NewDefaultAuthenticator()
	}
	if opts.HostKeyManager == nil {
		opts.HostKeyManager = NewDefaultHostKeyManager()
	}
	if opts.Dialer == nil {
		opts.Dialer = NewDefaultDialer()
	}

	return &client{
		authenticator:  opts.Authenticator,
		hostKeyManager: opts.HostKeyManager,
		dialer:         opts.Dialer,
	}
}

// Connect establishes an SSH connection
func (c *client) Connect(ctx context.Context, config *Config) error {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return err
	}

	c.config = config

	// Get host key callback
	hostKeyCallback, err := c.hostKeyManager.GetHostKeyCallback(ctx, config)
	if err != nil {
		return ErrHostKeyRejected.Wrap(err)
	}

	// Get authentication methods
	authMethods, err := c.authenticator.GetAuthMethods(ctx, config)
	if err != nil {
		return ErrAuthFailed.Wrap(err)
	}

	// Create SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         config.Timeout,
	}

	// Establish connection
	conn, err := c.dialer.DialContext(ctx, "tcp", config.Address(), sshConfig)
	if err != nil {
		return ErrConnectionFailed.Wrap(err)
	}

	c.conn = conn
	return nil
}

// Execute runs a command on the remote host
func (c *client) Execute(ctx context.Context, command string) error {
	if c.conn == nil {
		return ErrNotConnected
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return ErrSessionCreation.Wrap(err)
	}

	type result struct {
		err error
	}
	done := make(chan result, 1)

	go func() {
		err := session.Run(command)
		if cerr := session.Close(); cerr != nil {
			// Logue l'erreur si la fermeture échoue
			fmt.Fprintf(os.Stderr, "warning: failed to close SSH session: %v\n", cerr)
		}
		done <- result{err: err}
	}()

	select {
	case <-ctx.Done():
		return ErrCommandTimeout.Wrap(ctx.Err())
	case res := <-done:
		if res.err != nil {
			return ErrCommandFailed.Wrap(res.err)
		}
		return nil
	}
}

// Close closes the SSH connection
func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
