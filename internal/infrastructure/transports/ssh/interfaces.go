// internal/transports/ssh/interfaces.go - All interfaces for dependency injection
package ssh

import (
	"context"
	"net"

	"golang.org/x/crypto/ssh"
)

// Core interfaces for SSH operations

// Session represents an SSH session for executing commands.
//
// Session provides methods to run commands and manage the session lifecycle.
type Session interface {
	// Run executes a command on the remote server.
	//
	// Parameters:
	//   - cmd: Command string to execute
	//
	// Returns:
	//   - Error if command execution fails
	Run(cmd string) error

	// Close terminates the SSH session.
	//
	// Returns:
	//   - Error if session closure fails
	Close() error
}

// Connection represents an SSH connection to a remote server.
//
// Connection manages the SSH connection lifecycle and session creation.
type Connection interface {
	// NewSession creates a new SSH session on this connection.
	//
	// Returns:
	//   - Session instance for command execution
	//   - Error if session creation fails
	NewSession() (Session, error)

	// Close terminates the SSH connection.
	//
	// Returns:
	//   - Error if connection closure fails
	Close() error
}

// Dialer establishes SSH connections to remote servers.
//
// Dialer provides methods to create new SSH connections with proper configuration.
type Dialer interface {
	// DialContext establishes an SSH connection with context support.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - network: Network type (typically "tcp")
	//   - addr: Remote address to connect to
	//   - config: SSH client configuration
	//
	// Returns:
	//   - Connection instance for the established connection
	//   - Error if connection establishment fails
	DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (Connection, error)
}

// Authenticator handles SSH authentication methods.
//
// Authenticator provides authentication methods based on configuration.
type Authenticator interface {
	// GetAuthMethods returns SSH authentication methods for the given configuration.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - config: SSH configuration containing authentication details
	//
	// Returns:
	//   - Slice of SSH authentication methods
	//   - Error if authentication setup fails
	GetAuthMethods(ctx context.Context, config *Config) ([]ssh.AuthMethod, error)
}

// HostKeyManager handles SSH host key verification.
//
// HostKeyManager provides host key validation and callback functions.
type HostKeyManager interface {
	// GetHostKeyCallback returns a callback function for host key verification.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - config: SSH configuration containing host key settings
	//
	// Returns:
	//   - Host key callback function
	//   - Error if callback setup fails
	GetHostKeyCallback(ctx context.Context, config *Config) (ssh.HostKeyCallback, error)
}

// PasswordReader reads passwords from user input.
//
// PasswordReader provides secure password input functionality.
type PasswordReader interface {
	// ReadPassword prompts for and reads a password securely.
	//
	// Parameters:
	//   - prompt: Text to display to the user
	//
	// Returns:
	//   - Password string entered by user
	//   - Error if password reading fails
	ReadPassword(prompt string) (string, error)
}

// KeyLoader loads SSH private keys from storage.
//
// KeyLoader provides functionality to load and parse SSH private keys.
type KeyLoader interface {
	// LoadKey loads an SSH private key from the specified path.
	//
	// Parameters:
	//   - path: File path to the private key
	//
	// Returns:
	//   - SSH signer instance for the loaded key
	//   - Error if key loading or parsing fails
	LoadKey(path string) (ssh.Signer, error)
}

// HostKeyStore manages known SSH host keys.
//
// HostKeyStore provides storage and verification of known host keys.
type HostKeyStore interface {
	// IsKnown checks if a host key is already known.
	//
	// Parameters:
	//   - hostname: Name or address of the host
	//   - remote: Remote network address
	//   - key: Public key to verify
	//
	// Returns:
	//   - True if the key is known and valid, false otherwise
	IsKnown(hostname string, remote net.Addr, key ssh.PublicKey) bool

	// Add stores a new host key as known.
	//
	// Parameters:
	//   - hostname: Name or address of the host
	//   - key: Public key to store
	//
	// Returns:
	//   - Error if key storage fails
	Add(hostname string, key ssh.PublicKey) error

	// GetCallback returns a host key callback function.
	//
	// Returns:
	//   - Host key callback function for SSH client
	GetCallback() ssh.HostKeyCallback
}

// UserPrompter prompts users for interactive decisions.
//
// UserPrompter provides methods for user interaction and confirmation.
type UserPrompter interface {
	// PromptYesNo displays a message and prompts for yes/no response.
	//
	// Parameters:
	//   - message: Text to display to the user
	//
	// Returns:
	//   - True if user responds yes, false if no
	//   - Error if prompt interaction fails
	PromptYesNo(message string) (bool, error)
}

// Client provides high-level SSH operations.
//
// Client combines all SSH functionality into a single interface for ease of use.
type Client interface {
	// Connect establishes an SSH connection using the provided configuration.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - config: SSH configuration for the connection
	//
	// Returns:
	//   - Error if connection establishment fails
	Connect(ctx context.Context, config *Config) error

	// Execute runs a command on the connected SSH server.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - command: Command string to execute
	//
	// Returns:
	//   - Error if command execution fails
	Execute(ctx context.Context, command string) error

	// Close terminates the SSH connection.
	//
	// Returns:
	//   - Error if connection closure fails
	Close() error
}
