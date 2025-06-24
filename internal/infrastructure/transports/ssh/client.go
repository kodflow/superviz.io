// internal/transports/ssh/client.go - Main SSH client implementation
package ssh

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// client implements the SSH client interface with dependency injection support.
//
// client provides a complete SSH client implementation that combines authentication,
// host key management, and connection handling into a single cohesive interface.
type client struct {
	// conn holds the active SSH connection
	conn Connection
	// config contains the SSH connection configuration
	config *Config
	// authenticator handles SSH authentication methods
	authenticator Authenticator
	// hostKeyManager manages host key verification
	hostKeyManager HostKeyManager
	// dialer establishes network connections
	dialer Dialer
}

// ClientOptions contains options for creating a new SSH client.
//
// ClientOptions allows customization of client behavior through dependency injection
// of authenticators, host key managers, and dialers.
type ClientOptions struct {
	// Authenticator handles SSH authentication (optional, defaults to standard implementation)
	Authenticator Authenticator
	// HostKeyManager handles host key verification (optional, defaults to standard implementation)
	HostKeyManager HostKeyManager
	// Dialer handles network connection establishment (optional, defaults to standard implementation)
	Dialer Dialer
}

// NewClient creates a new SSH client with the given options.
//
// NewClient initializes an SSH client with the provided options, using default
// implementations for any nil components. This design enables dependency injection
// for testing and customization while providing sensible defaults.
//
// Parameters:
//   - opts: Client options for dependency injection (can be nil for all defaults)
//
// Returns:
//   - Client instance ready for use
func NewClient(opts *ClientOptions) Client {
	if opts == nil {
		return &client{
			authenticator:  NewDefaultAuthenticator(),
			hostKeyManager: NewDefaultHostKeyManager(),
			dialer:         NewDefaultDialer(),
		}
	}

	// Initialize only nil fields
	c := &client{
		authenticator:  opts.Authenticator,
		hostKeyManager: opts.HostKeyManager,
		dialer:         opts.Dialer,
	}

	if c.authenticator == nil {
		c.authenticator = NewDefaultAuthenticator()
	}
	if c.hostKeyManager == nil {
		c.hostKeyManager = NewDefaultHostKeyManager()
	}
	if c.dialer == nil {
		c.dialer = NewDefaultDialer()
	}

	return c
}

// Connect establishes an SSH connection using the provided configuration.
//
// Connect validates the configuration, sets up authentication and host key verification,
// then establishes the SSH connection. The connection remains active until Close is called.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - config: SSH configuration for the connection
//
// Returns:
//   - Error if connection establishment fails
func (c *client) Connect(ctx context.Context, config *Config) error {
	// Validate and cache config
	if err := config.Validate(); err != nil {
		return err
	}
	c.config = config

	// Get host key callback
	hostKeyCallback, err := c.hostKeyManager.GetHostKeyCallback(ctx, config)
	if err != nil {
		return WrapError(ErrHostKeyRejected, err)
	}

	// Get authentication methods
	authMethods, err := c.authenticator.GetAuthMethods(ctx, config)
	if err != nil {
		return WrapError(ErrAuthFailed, err)
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
		return err // Already wrapped by dialer
	}

	c.conn = conn
	return nil
}

// Execute runs a command on the remote SSH server.
//
// Execute creates a new SSH session, runs the specified command, and handles
// the session lifecycle. The command execution respects the provided context
// for timeout and cancellation.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - command: Command string to execute on the remote server
//
// Returns:
//   - Error if command execution fails or times out
func (c *client) Execute(ctx context.Context, command string) error {
	if c.conn == nil {
		return ErrNotConnected
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return WrapError(ErrSessionCreation, err)
	}

	defer func() {
		if cerr := session.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close SSH session: %v\n", cerr)
		}
	}()

	// Execute command with context
	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		return WrapError(ErrCommandTimeout, ctx.Err())
	case err := <-done:
		if err != nil {
			return WrapError(ErrCommandFailed, err)
		}
		return nil
	}
}

// Close closes the SSH connection and releases associated resources.
//
// Close terminates the active SSH connection if one exists. It is safe to call
// Close multiple times or on a client that was never connected.
//
// Returns:
//   - Error if connection closure fails
func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
