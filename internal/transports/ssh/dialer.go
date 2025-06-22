// internal/transports/ssh/dialer.go - SSH connection dialing implementations
package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// defaultDialer implements the Dialer interface
type defaultDialer struct {
	// Optional: custom net.Dialer for advanced use cases
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
	return &defaultDialer{
		netDialer: netDialer,
	}
}

// DialContext establishes an SSH connection with context support
func (d *defaultDialer) DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (Connection, error) {
	// Use custom dialer if provided, otherwise use ssh.Dial
	var conn net.Conn
	var err error

	if d.netDialer != nil {
		// Use context-aware dialing with custom dialer
		conn, err = d.netDialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, d.enhanceError(err, addr)
		}

		// Perform SSH handshake
		sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close underlying TCP connection: %v\n", cerr)
			}

			return nil, d.enhanceError(err, addr)
		}

		// Create SSH client
		client := ssh.NewClient(sshConn, chans, reqs)
		return &sshConnection{client: client}, nil
	}

	// Use standard SSH dial with context support
	type dialResult struct {
		client *ssh.Client
		err    error
	}

	resultCh := make(chan dialResult, 1)

	go func() {
		// Create a connection that can be cancelled
		dialer := &net.Dialer{
			Timeout: config.Timeout,
		}
		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			resultCh <- dialResult{err: err}
			return
		}

		// Perform SSH handshake
		sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close connection after handshake failure: %v\n", cerr)
			}
			resultCh <- dialResult{err: err}
			return
		}

		client := ssh.NewClient(sshConn, chans, reqs)
		resultCh <- dialResult{client: client, err: nil}
	}()

	// Wait for connection or context cancellation
	select {
	case <-ctx.Done():
		return nil, ErrConnectionFailed.Wrap(ctx.Err()).WithMessage("connection cancelled")
	case result := <-resultCh:
		if result.err != nil {
			return nil, d.enhanceError(result.err, addr)
		}
		return &sshConnection{client: result.client}, nil
	}
}

// enhanceError provides better error messages based on the error type
func (d *defaultDialer) enhanceError(err error, addr string) error {
	if err == nil {
		return nil
	}

	// Context errors
	if err == context.Canceled {
		return ErrConnectionFailed.Wrap(err).WithMessage("connection cancelled")
	}
	if err == context.DeadlineExceeded {
		return ErrConnectionFailed.Wrap(err).WithMessage("connection timeout")
	}

	// Network errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return ErrConnectionFailed.Wrap(err).
				WithMessage("connection timeout").
				WithContext("address", addr)
		}
	}

	// SSH-specific errors
	errStr := err.Error()
	switch {
	case contains(errStr, "permission denied", "unable to authenticate"):
		return ErrAuthFailed.Wrap(err).
			WithMessage("authentication failed").
			WithContext("address", addr)
	case contains(errStr, "connection refused"):
		return ErrConnectionFailed.Wrap(err).
			WithMessage("connection refused (SSH service may not be running)").
			WithContext("address", addr)
	case contains(errStr, "no route to host", "host is unreachable"):
		return ErrConnectionFailed.Wrap(err).
			WithMessage("host unreachable (check network connectivity)").
			WithContext("address", addr)
	case contains(errStr, "host key verification failed"):
		return ErrHostKeyRejected.Wrap(err).
			WithContext("address", addr)
	default:
		return ErrConnectionFailed.Wrap(err).
			WithContext("address", addr)
	}
}

// sshConnection wraps an ssh.Client to implement the Connection interface
type sshConnection struct {
	client *ssh.Client
}

// NewSession creates a new SSH session
func (c *sshConnection) NewSession() (Session, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, ErrSessionCreation.Wrap(err)
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

func contains(s string, substrs ...string) bool {
	lowerS := strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(lowerS, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
