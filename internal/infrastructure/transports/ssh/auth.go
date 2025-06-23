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
	// Fast path: SSH key authentication
	if config.KeyPath != "" {
		// Check cache first
		if cached, ok := a.keyCache.Load(config.KeyPath); ok {
			return []ssh.AuthMethod{ssh.PublicKeys(cached.(ssh.Signer))}, nil
		}

		// Load and cache key
		signer, err := a.keyLoader.LoadKey(config.KeyPath)
		if err != nil {
			return nil, NewError(ErrAuthFailed, err.Error()).
				WithContext("key_path", config.KeyPath)
		}

		a.keyCache.Store(config.KeyPath, signer)
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	}

	// Fallback: password authentication
	password, err := a.passwordReader.ReadPassword(
		fmt.Sprintf("Password for %s@%s: ", config.User, config.Host))
	if err != nil {
		return nil, WrapError(ErrAuthFailed, err)
	}

	return []ssh.AuthMethod{ssh.Password(password)}, nil
}

// terminalPasswordReader reads passwords from terminal
type terminalPasswordReader struct{}

// ReadPassword reads a password from the terminal
func (t *terminalPasswordReader) ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Read password securely
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	fmt.Println() // Add newline after password input

	// Convert and clear in one step
	result := string(password)
	for i := range password {
		password[i] = 0
	}

	return result, nil
}

// fileKeyLoader loads keys from files
type fileKeyLoader struct{}

// LoadKey loads a private key from file
func (f *fileKeyLoader) LoadKey(path string) (ssh.Signer, error) {
	// Read key file
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(keyData)

	// Clear key data immediately
	for i := range keyData {
		keyData[i] = 0
	}

	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return signer, nil
}
