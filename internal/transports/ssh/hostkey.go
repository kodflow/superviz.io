// internal/transports/ssh/hostkey.go - Host key verification implementations
package ssh

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyStore manages known host keys
type HostKeyStore interface {
	// IsKnown checks if a host key is already known
	IsKnown(hostname string, remote net.Addr, key ssh.PublicKey) bool
	// Add adds a new host key to the store
	Add(hostname string, key ssh.PublicKey) error
	// GetCallback returns an ssh.HostKeyCallback for verification
	GetCallback() ssh.HostKeyCallback
}

// UserPrompter prompts users for decisions
type UserPrompter interface {
	// PromptYesNo prompts the user with a yes/no question
	PromptYesNo(message string) (bool, error)
}

// defaultHostKeyManager implements the HostKeyManager interface
type defaultHostKeyManager struct {
	store    HostKeyStore
	prompter UserPrompter
}

// NewDefaultHostKeyManager creates a new default host key manager
func NewDefaultHostKeyManager() HostKeyManager {
	return &defaultHostKeyManager{
		store:    &fileHostKeyStore{},
		prompter: &terminalPrompter{},
	}
}

// NewHostKeyManager creates a new host key manager with custom implementations
func NewHostKeyManager(store HostKeyStore, prompter UserPrompter) HostKeyManager {
	return &defaultHostKeyManager{
		store:    store,
		prompter: prompter,
	}
}

// GetHostKeyCallback returns the appropriate host key callback
func (m *defaultHostKeyManager) GetHostKeyCallback(ctx context.Context, config *Config) (HostKeyCallback, error) {
	// Skip verification if configured
	if config.SkipHostKeyCheck {
		fmt.Println("WARNING: Host key verification disabled (development mode)")
		return ssh.InsecureIgnoreHostKey(), nil
	}

	// Get the server's host key for verification
	hostKey, err := m.getServerHostKey(ctx, config)
	if err != nil {
		return nil, err
	}

	// Create remote address
	remote := &net.TCPAddr{
		IP:   net.ParseIP(config.Host),
		Port: config.Port,
	}

	// Check if key is already known
	hostname := formatHostPort(config.Host, config.Port)
	if m.store.IsKnown(hostname, remote, hostKey) {
		return m.store.GetCallback(), nil
	}

	// Handle unknown host key
	if config.AcceptNewHostKey {
		// Auto-accept mode
		m.displayHostKeyInfo(config, hostKey)
		if err := m.store.Add(hostname, hostKey); err != nil {
			return nil, err
		}
		fmt.Printf("Warning: Permanently added '%s' to the list of known hosts.\n", hostname)
		return m.store.GetCallback(), nil
	}

	// Interactive mode - prompt user
	m.displayHostKeyInfo(config, hostKey)

	prompt := "Are you sure you want to continue connecting (yes/no/[fingerprint])? "
	accepted, err := m.prompter.PromptYesNo(prompt)
	if err != nil {
		return nil, err
	}

	if !accepted {
		return nil, ErrHostKeyRejected.WithMessage("user rejected host key")
	}

	// Save the accepted key
	if err := m.store.Add(hostname, hostKey); err != nil {
		return nil, err
	}
	fmt.Printf("Warning: Permanently added '%s' to the list of known hosts.\n", hostname)

	return m.store.GetCallback(), nil
}

// getServerHostKey retrieves the host key from the server
func (m *defaultHostKeyManager) getServerHostKey(ctx context.Context, config *Config) (ssh.PublicKey, error) {
	var capturedKey ssh.PublicKey

	// Create a temporary config to capture the host key
	tempConfig := &ssh.ClientConfig{
		User:    config.User,
		Timeout: config.Timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			capturedKey = key
			return fmt.Errorf("key captured") // Intentional error to stop connection
		},
		Auth: []ssh.AuthMethod{
			// Dummy auth method to avoid auth errors
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				return nil, fmt.Errorf("not used")
			}),
		},
	}

	// Try to connect just to capture the key
	type result struct {
		key ssh.PublicKey
		err error
	}

	resultCh := make(chan result, 1)

	go func() {
		conn, err := ssh.Dial("tcp", config.Address(), tempConfig)
		if conn != nil {
			if cerr := conn.Close(); cerr != nil {
				// On logue l'erreur de fermeture sans remplacer l'erreur principale
				fmt.Fprintf(os.Stderr, "warning: failed to close SSH connection: %v\n", cerr)
			}
		}

		if capturedKey != nil {
			resultCh <- result{key: capturedKey, err: nil}
			return
		}

		if err != nil && strings.Contains(err.Error(), "key captured") {
			resultCh <- result{key: capturedKey, err: nil}
			return
		}

		resultCh <- result{err: fmt.Errorf("failed to capture host key: %w", err)}
	}()

	select {
	case <-ctx.Done():
		return nil, ErrConnectionFailed.Wrap(ctx.Err()).WithMessage("host key retrieval cancelled")
	case res := <-resultCh:
		if res.err != nil {
			return nil, res.err
		}
		return res.key, nil
	}
}

// displayHostKeyInfo displays information about an unknown host key
func (m *defaultHostKeyManager) displayHostKeyInfo(config *Config, key ssh.PublicKey) {
	hostname := formatHostPort(config.Host, config.Port)
	keyType := getKeyTypeDisplay(key.Type())
	fingerprint := ssh.FingerprintSHA256(key)

	fmt.Printf("The authenticity of host '%s (%s)' can't be established.\n", hostname, hostname)
	fmt.Printf("%s key fingerprint is %s.\n", keyType, fingerprint)
	fmt.Println("This key is not known by any other names.")
}

// fileHostKeyStore implements HostKeyStore using the filesystem
type fileHostKeyStore struct {
	knownHostsPath string
}

// IsKnown checks if a host key is already known
func (s *fileHostKeyStore) IsKnown(hostname string, remote net.Addr, key ssh.PublicKey) bool {
	callback := s.GetCallback()
	if callback == nil {
		return false
	}

	err := callback(hostname, remote, key)
	return err == nil
}

// Add adds a new host key to the store
func (s *fileHostKeyStore) Add(hostname string, key ssh.PublicKey) (err error) {
	knownHostsPath := s.getKnownHostsPath()

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(knownHostsPath)
	if err = os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Open or create known_hosts file
	file, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close known_hosts file: %w", cerr)
		}
	}()

	// Format and write the host key entry
	line := knownhosts.Line([]string{hostname}, key)
	if _, err = fmt.Fprintln(file, line); err != nil {
		return fmt.Errorf("failed to write host key: %w", err)
	}

	return nil
}

// GetCallback returns an ssh.HostKeyCallback for verification
func (s *fileHostKeyStore) GetCallback() ssh.HostKeyCallback {
	knownHostsPath := s.getKnownHostsPath()

	// Check if known_hosts file exists
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil
	}

	// Create callback from known_hosts file
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil
	}

	return callback
}

// getKnownHostsPath returns the path to the known_hosts file
func (s *fileHostKeyStore) getKnownHostsPath() string {
	if s.knownHostsPath != "" {
		return s.knownHostsPath
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".ssh", "known_hosts")
}

// terminalPrompter implements UserPrompter using terminal input
type terminalPrompter struct{}

// PromptYesNo prompts the user with a yes/no question
func (p *terminalPrompter) PromptYesNo(message string) (bool, error) {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y", nil
}

// formatHostPort formats host and port for known_hosts format
func formatHostPort(host string, port int) string {
	if port == 22 {
		return host
	}
	return fmt.Sprintf("[%s]:%d", host, port)
}

// getKeyTypeDisplay converts SSH key type to display format
func getKeyTypeDisplay(keyType string) string {
	switch keyType {
	case "ssh-ed25519":
		return "ED25519"
	case "ssh-rsa":
		return "RSA"
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		return "ECDSA"
	case "ssh-dss":
		return "DSA"
	default:
		return strings.ToUpper(keyType)
	}
}
