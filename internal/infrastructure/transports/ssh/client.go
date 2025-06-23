// internal/transports/ssh/client.go - Main SSH client implementation
package ssh

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

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

// Connect establishes an SSH connection
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

// Execute runs a command on the remote host
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

// Close closes the SSH connection
func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
