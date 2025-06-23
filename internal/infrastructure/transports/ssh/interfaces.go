// internal/transports/ssh/interfaces.go - All interfaces for dependency injection
package ssh

import (
	"context"
	"net"

	"golang.org/x/crypto/ssh"
)

// Core interfaces for SSH operations

// Session represents an SSH session
type Session interface {
	Run(cmd string) error
	Close() error
}

// Connection represents an SSH connection
type Connection interface {
	NewSession() (Session, error)
	Close() error
}

// Dialer establishes SSH connections
type Dialer interface {
	DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (Connection, error)
}

// Authenticator handles SSH authentication
type Authenticator interface {
	GetAuthMethods(ctx context.Context, config *Config) ([]ssh.AuthMethod, error)
}

// HostKeyManager handles host key verification
type HostKeyManager interface {
	GetHostKeyCallback(ctx context.Context, config *Config) (ssh.HostKeyCallback, error)
}

// PasswordReader reads passwords from the user
type PasswordReader interface {
	ReadPassword(prompt string) (string, error)
}

// KeyLoader loads SSH private keys
type KeyLoader interface {
	LoadKey(path string) (ssh.Signer, error)
}

// HostKeyStore manages known host keys
type HostKeyStore interface {
	IsKnown(hostname string, remote net.Addr, key ssh.PublicKey) bool
	Add(hostname string, key ssh.PublicKey) error
	GetCallback() ssh.HostKeyCallback
}

// UserPrompter prompts users for decisions
type UserPrompter interface {
	PromptYesNo(message string) (bool, error)
}

// Client provides SSH operations
type Client interface {
	Connect(ctx context.Context, config *Config) error
	Execute(ctx context.Context, command string) error
	Close() error
}
