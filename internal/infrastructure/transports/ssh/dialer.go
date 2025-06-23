// internal/transports/ssh/dialer.go - SSH connection dialing implementations
package ssh

import (
	"context"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Pre-compiled error patterns for performance
var (
	authErrors = []string{"permission denied", "unable to authenticate"}
	netErrors  = []string{"connection refused", "no route to host", "host is unreachable"}
)

// defaultDialer implements the Dialer interface
type defaultDialer struct {
	netDialer *net.Dialer
}

// NewDefaultDialer creates a new default SSH dialer
func NewDefaultDialer() Dialer {
	return &defaultDialer{
		netDialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
}

// NewDialer creates a new SSH dialer with a custom net.Dialer
func NewDialer(netDialer *net.Dialer) Dialer {
	return &defaultDialer{netDialer: netDialer}
}

// DialContext establishes an SSH connection with context support
func (d *defaultDialer) DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (Connection, error) {
	// Direct dial with context
	conn, err := d.netDialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, d.wrapError(err, addr)
	}

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close() //nolint:errcheck
		return nil, d.wrapError(err, addr)
	}

	return &sshConnection{client: ssh.NewClient(sshConn, chans, reqs)}, nil
}

// wrapError provides appropriate error wrapping based on error type
func (d *defaultDialer) wrapError(err error, addr string) error {
	// Fast path for context errors
	switch err {
	case context.Canceled:
		return NewError(ErrConnectionFailed, "connection cancelled")
	case context.DeadlineExceeded:
		return NewError(ErrConnectionFailed, "connection timeout")
	}

	// Check network errors
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return NewError(ErrConnectionFailed, "connection timeout").
			WithContext("address", addr)
	}

	// Pattern matching for SSH errors
	errStr := strings.ToLower(err.Error())

	// Check auth errors
	for _, pattern := range authErrors {
		if strings.Contains(errStr, pattern) {
			return WrapError(ErrAuthFailed, err).WithContext("address", addr)
		}
	}

	// Check network errors
	for _, pattern := range netErrors {
		if strings.Contains(errStr, pattern) {
			return WrapError(ErrConnectionFailed, err).WithContext("address", addr)
		}
	}

	// Check host key error
	if strings.Contains(errStr, "host key") {
		return WrapError(ErrHostKeyRejected, err).WithContext("address", addr)
	}

	// Default
	return WrapError(ErrConnectionFailed, err).WithContext("address", addr)
}

// sshConnection wraps an ssh.Client to implement the Connection interface
type sshConnection struct {
	client *ssh.Client
}

// NewSession creates a new SSH session
func (c *sshConnection) NewSession() (Session, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, WrapError(ErrSessionCreation, err)
	}
	return &sshSession{session: session}, nil
}

// Close closes the SSH connection
func (c *sshConnection) Close() error {
	return c.client.Close()
}

// sshSession wraps an ssh.Session to implement the Session interface
type sshSession struct {
	session *ssh.Session
}

// Run executes a command and waits for it to complete
func (s *sshSession) Run(cmd string) error {
	return s.session.Run(cmd)
}

// Close closes the SSH session
func (s *sshSession) Close() error {
	return s.session.Close()
}
