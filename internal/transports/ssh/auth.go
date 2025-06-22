// internal/transports/ssh/auth.go - SSH authentication implementations
package ssh

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// PasswordReader reads passwords from the user
type PasswordReader interface {
	// ReadPassword reads a password from the user
	ReadPassword(prompt string) (string, error)
}

// KeyLoader loads SSH private keys
type KeyLoader interface {
	// LoadKey loads a private key from the given path
	LoadKey(path string) (ssh.Signer, error)
}

// defaultAuthenticator implements the Authenticator interface
type defaultAuthenticator struct {
	passwordReader PasswordReader
	keyLoader      KeyLoader
	keyCache       sync.Map // Cache for loaded private keys
}

// NewDefaultAuthenticator creates a new default authenticator
func NewDefaultAuthenticator() Authenticator {
	return &defaultAuthenticator{
		passwordReader: &terminalPasswordReader{},
		keyLoader:      &fileKeyLoader{},
	}
}

// NewAuthenticator creates a new authenticator with custom implementations
func NewAuthenticator(passwordReader PasswordReader, keyLoader KeyLoader) Authenticator {
	return &defaultAuthenticator{
		passwordReader: passwordReader,
		keyLoader:      keyLoader,
	}
}

// GetAuthMethods returns the authentication methods based on config
func (a *defaultAuthenticator) GetAuthMethods(ctx context.Context, config *Config) ([]ssh.AuthMethod, error) {
	// Use SSH key if path is provided
	if config.KeyPath != "" {
		signer, err := a.loadKeyWithCache(config.KeyPath)
		if err != nil {
			return nil, ErrAuthFailed.Wrap(err).WithContext("key_path", config.KeyPath)
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	}
	
	
	// Otherwise use password authentication
	prompt := fmt.Sprintf("Password for %s@%s: ", config.User, config.Host)
	password, err := a.passwordReader.ReadPassword(prompt)
	if err != nil {
		return nil, ErrAuthFailed.Wrap(err).WithMessage("failed to read password")
	}
	
	return []ssh.AuthMethod{ssh.Password(password)}, nil
}

// loadKeyWithCache loads a key with caching for performance
func (a *defaultAuthenticator) loadKeyWithCache(keyPath string) (ssh.Signer, error) {
	// Check cache first
	if cached, ok := a.keyCache.Load(keyPath); ok {
		if signer, ok := cached.(ssh.Signer); ok {
			return signer, nil
		}
	}
	
	// Load key
	signer, err := a.keyLoader.LoadKey(keyPath)
	if err != nil {
		return nil, err
	}
	
	// Cache the signer
	a.keyCache.Store(keyPath, signer)
	return signer, nil
}

// terminalPasswordReader reads passwords from terminal
type terminalPasswordReader struct{}

// ReadPassword reads a password from the terminal
func (t *terminalPasswordReader) ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	
	// Read password securely
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	
	fmt.Println() // Add newline after password input
	
	// Convert to string and clear the byte slice
	result := string(password)
	clearBytes(password)
	
	return result, nil
}

// fileKeyLoader loads keys from files
type fileKeyLoader struct{}

// LoadKey loads a private key from file
func (f *fileKeyLoader) LoadKey(path string) (ssh.Signer, error) {
	// Read key file
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key file: %w", err)
	}
	
	// Clear key data after parsing
	defer clearBytes(keyData)
	
	// Parse private key
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}
	
	return signer, nil
}

// clearBytes securely clears a byte slice
func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}