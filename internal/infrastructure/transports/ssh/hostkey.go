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
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Key type display names
var keyTypeNames = map[string]string{
	"ssh-ed25519":         "ED25519",
	"ssh-rsa":             "RSA",
	"ecdsa-sha2-nistp256": "ECDSA",
	"ecdsa-sha2-nistp384": "ECDSA",
	"ecdsa-sha2-nistp521": "ECDSA",
	"ssh-dss":             "DSA",
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
func (m *defaultHostKeyManager) GetHostKeyCallback(ctx context.Context, config *Config) (ssh.HostKeyCallback, error) {
	// Fast path: skip verification if configured
	if config.SkipHostKeyCheck {
		fmt.Fprintln(os.Stderr, "WARNING: Host key verification disabled")
		return ssh.InsecureIgnoreHostKey(), nil
	}

	// Get callback from store
	if callback := m.store.GetCallback(); callback != nil {
		return callback, nil
	}

	// Need to handle unknown host key
	return m.createInteractiveCallback(config), nil
}

// createInteractiveCallback creates a callback that handles unknown hosts
func (m *defaultHostKeyManager) createInteractiveCallback(config *Config) ssh.HostKeyCallback {
	var once sync.Once
	var storedErr error

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Check if already known
		if m.store.IsKnown(hostname, remote, key) {
			return nil
		}

		// Handle unknown key once
		once.Do(func() {
			storedErr = m.handleUnknownKey(config, hostname, key)
		})

		return storedErr
	}
}

// handleUnknownKey processes an unknown host key
func (m *defaultHostKeyManager) handleUnknownKey(config *Config, hostname string, key ssh.PublicKey) error {
	// Display key info
	keyType := getKeyTypeDisplay(key.Type())
	fingerprint := ssh.FingerprintSHA256(key)

	fmt.Printf("The authenticity of host '%s' can't be established.\n", hostname)
	fmt.Printf("%s key fingerprint is %s.\n", keyType, fingerprint)

	// Auto-accept if configured
	if config.AcceptNewHostKey {
		if err := m.store.Add(hostname, key); err != nil {
			return err
		}
		fmt.Printf("Warning: Permanently added '%s' to known hosts.\n", hostname)
		return nil
	}

	// Interactive prompt
	accepted, err := m.prompter.PromptYesNo("Continue connecting (yes/no)? ")
	if err != nil {
		return err
	}

	if !accepted {
		return NewError(ErrHostKeyRejected, "user rejected host key")
	}

	// Save the accepted key
	if err := m.store.Add(hostname, key); err != nil {
		return err
	}
	fmt.Printf("Warning: Permanently added '%s' to known hosts.\n", hostname)
	return nil
}

// fileHostKeyStore implements HostKeyStore using the filesystem
type fileHostKeyStore struct {
	path     string
	callback ssh.HostKeyCallback
	mu       sync.RWMutex
}

// IsKnown checks if a host key is already known
func (s *fileHostKeyStore) IsKnown(hostname string, remote net.Addr, key ssh.PublicKey) bool {
	callback := s.GetCallback()
	if callback == nil {
		return false
	}
	return callback(hostname, remote, key) == nil
}

// Add adds a new host key to the store
func (s *fileHostKeyStore) Add(hostname string, key ssh.PublicKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	knownHostsPath := s.getPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(knownHostsPath), 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Append to known_hosts
	file, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts: %w", err)
	}
	defer file.Close() //nolint:errcheck

	// Write the host key entry
	line := knownhosts.Line([]string{hostname}, key)
	if _, err = fmt.Fprintln(file, line); err != nil {
		return fmt.Errorf("failed to write host key: %w", err)
	}

	// Invalidate cached callback
	s.callback = nil
	return nil
}

// GetCallback returns an ssh.HostKeyCallback for verification
func (s *fileHostKeyStore) GetCallback() ssh.HostKeyCallback {
	s.mu.RLock()
	if s.callback != nil {
		s.mu.RUnlock()
		return s.callback
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.callback != nil {
		return s.callback
	}

	knownHostsPath := s.getPath()

	// Check if file exists
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil
	}

	// Create and cache callback with timeout protection
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		// Log warning but don't fail completely - allow fallback to interactive mode
		fmt.Fprintf(os.Stderr, "warning: failed to load known_hosts file: %v\n", err)
		return nil
	}

	s.callback = callback
	return callback
}

// getPath returns the path to the known_hosts file
func (s *fileHostKeyStore) getPath() string {
	if s.path != "" {
		return s.path
	}

	// Cache the path
	homeDir, _ := os.UserHomeDir()
	s.path = filepath.Join(homeDir, ".ssh", "known_hosts")
	return s.path
}

// terminalPrompter implements UserPrompter using terminal input
type terminalPrompter struct {
	reader *bufio.Reader
}

// PromptYesNo prompts the user with a yes/no question
func (p *terminalPrompter) PromptYesNo(message string) (bool, error) {
	fmt.Print(message)

	if p.reader == nil {
		p.reader = bufio.NewReader(os.Stdin)
	}

	response, err := p.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y", nil
}

// getKeyTypeDisplay converts SSH key type to display format
func getKeyTypeDisplay(keyType string) string {
	if name, ok := keyTypeNames[keyType]; ok {
		return name
	}
	return strings.ToUpper(keyType)
}
