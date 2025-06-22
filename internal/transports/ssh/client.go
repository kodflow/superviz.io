package ssh

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// Config contains SSH connection configuration.
type Config struct {
	Host             string
	User             string
	Port             int
	KeyPath          string // Path to private key (optional)
	Timeout          time.Duration
	SkipHostKeyCheck bool // Skip host key verification (development only)
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

	// Create SSH client config with secure host key checking
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Timeout:         config.Timeout,
		HostKeyCallback: c.getHostKeyCallback(),
	}

	// Set up authentication: try key first, then password
	var authMethods []ssh.AuthMethod

	// 1. Try SSH key authentication if provided
	if config.KeyPath != "" {
		if auth, err := c.loadPrivateKey(config.KeyPath); err == nil {
			authMethods = append(authMethods, auth)
		} else {
			fmt.Printf("Warning: failed to load SSH key %s: %v\n", config.KeyPath, err)
		}
	}

	// 2. Try default SSH keys if no explicit key provided
	if config.KeyPath == "" {
		if auth := c.tryDefaultKeys(); auth != nil {
			authMethods = append(authMethods, auth)
		}
	}

	// 3. Add password authentication as fallback
	if len(authMethods) == 0 {
		// No keys worked, prompt for password
		password, err := c.promptPassword()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		authMethods = append(authMethods, ssh.Password(password))
	} else {
		// Keys available, but add password as fallback in case keys fail
		authMethods = append(authMethods, ssh.PasswordCallback(func() (string, error) {
			return c.promptPassword()
		}))
	}

	sshConfig.Auth = authMethods

	// Connect to the remote host
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c.conn = conn
	return nil
}

// getHostKeyCallback returns a secure host key callback using known_hosts.
func (c *client) getHostKeyCallback() ssh.HostKeyCallback {
	// Development mode: skip host key checking
	if c.config.SkipHostKeyCheck {
		fmt.Println("WARNING: Host key verification disabled (development mode)")
		return ssh.InsecureIgnoreHostKey()
	}

	// Try to load known_hosts file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Warning: Cannot determine home directory, skipping host key verification")
		return ssh.InsecureIgnoreHostKey()
	}

	knownHostsFile := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Try to create host key callback from known_hosts
	if _, err := os.Stat(knownHostsFile); err == nil {
		if callback, err := knownhosts.New(knownHostsFile); err == nil {
			return callback
		}
	}

	// If no known_hosts, use InsecureIgnoreHostKey with warning
	fmt.Printf("Warning: No known_hosts file found at %s. Host key verification disabled.\n", knownHostsFile)
	fmt.Println("This makes the connection vulnerable to man-in-the-middle attacks.")
	return ssh.InsecureIgnoreHostKey()
}

// loadPrivateKey loads a private key from file.
func (c *client) loadPrivateKey(keyPath string) (ssh.AuthMethod, error) {
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

// tryDefaultKeys tries to load default SSH keys.
func (c *client) tryDefaultKeys() ssh.AuthMethod {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	// Try common key files in order of preference
	keyFiles := []string{
		filepath.Join(homeDir, ".ssh", "id_ed25519"),
		filepath.Join(homeDir, ".ssh", "id_rsa"),
		filepath.Join(homeDir, ".ssh", "id_ecdsa"),
		filepath.Join(homeDir, ".ssh", "id_dsa"),
	}

	for _, keyFile := range keyFiles {
		if _, err := os.Stat(keyFile); err == nil {
			if auth, err := c.loadPrivateKey(keyFile); err == nil {
				return auth
			}
		}
	}

	return nil
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
