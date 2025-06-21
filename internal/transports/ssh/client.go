package ssh

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Config contains SSH connection configuration.
type Config struct {
	Host    string
	User    string
	Port    int
	KeyPath string
	Timeout time.Duration
}

// Client defines the interface for SSH operations.
type Client interface {
	Connect(ctx context.Context, config *Config) error
	Execute(ctx context.Context, command string) error
	Disconnect() error
}

// client implements the SSH client.
type client struct {
	conn   *ssh.Client
	config *Config
}

// NewClient creates a new SSH client.
func NewClient() Client {
	return &client{}
}

// Connect establishes an SSH connection.
func (c *client) Connect(ctx context.Context, config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	c.config = config

	// Create SSH client config
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: implement proper host key verification
		Timeout:         config.Timeout,
	}

	// Try to connect with different authentication methods
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// First, try SSH key authentication
	if err := c.tryKeyAuth(sshConfig, address); err == nil {
		return nil
	}

	// If key auth fails, try password authentication
	return c.tryPasswordAuth(sshConfig, address)
}

// Execute runs a command on the remote host.
func (c *client) Execute(ctx context.Context, command string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			// Log the error but don't fail the operation
			// Note: session.Close() often returns an error even on successful execution
			log.Printf("Warning: failed to close SSH session: %v", err)
		}
	}()

	// Execute the command
	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// Disconnect closes the SSH connection.
func (c *client) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// tryKeyAuth attempts SSH key authentication.
func (c *client) tryKeyAuth(sshConfig *ssh.ClientConfig, address string) error {
	var authMethods []ssh.AuthMethod

	if c.config.KeyPath != "" {
		// Try explicit SSH key first if provided
		if auth, err := c.publicKeyAuth(c.config.KeyPath); err == nil {
			authMethods = append(authMethods, auth)
		}
	} else {
		// Try default SSH keys if no explicit key provided
		defaultKeys := c.getDefaultSSHKeys()
		for _, keyPath := range defaultKeys {
			if auth, err := c.publicKeyAuth(keyPath); err == nil {
				authMethods = append(authMethods, auth)
				break // Use first working key
			}
		}
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no SSH keys available")
	}

	sshConfig.Auth = authMethods
	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// tryPasswordAuth attempts password authentication.
func (c *client) tryPasswordAuth(sshConfig *ssh.ClientConfig, address string) error {
	password, err := c.promptPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	sshConfig.Auth = []ssh.AuthMethod{ssh.Password(password)}
	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c.conn = conn
	return nil
}

// getDefaultSSHKeys returns the default SSH key paths to try.
func (c *client) getDefaultSSHKeys() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	return []string{
		fmt.Sprintf("%s/.ssh/id_rsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_ed25519", homeDir),
		fmt.Sprintf("%s/.ssh/id_ecdsa", homeDir),
		fmt.Sprintf("%s/.ssh/id_dsa", homeDir),
	}
}

// promptPassword prompts the user for a password securely.
func (c *client) promptPassword() (string, error) {
	fmt.Printf("Password for %s@%s: ", c.config.User, c.config.Host)

	// Read password without echoing to terminal
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Println() // Add newline after password input
	return string(password), nil
}

// publicKeyAuth creates public key authentication from a private key file.
func (c *client) publicKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}
