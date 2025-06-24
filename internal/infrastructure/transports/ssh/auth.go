// internal/transports/ssh/auth.go - SSH authentication implementations
package ssh

import (
	"context"
	"fmt"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// defaultAuthenticator implements the Authenticator interface with caching support.
//
// defaultAuthenticator provides SSH authentication using either SSH keys or passwords,
// with private key caching for improved performance on repeated connections.
type defaultAuthenticator struct {
	// passwordReader handles secure password input from users
	passwordReader PasswordReader
	// keyLoader handles loading SSH private keys from files
	keyLoader KeyLoader
	// keyCache stores loaded private keys to avoid repeated file I/O
	keyCache sync.Map // Cache for loaded private keys
}

// NewDefaultAuthenticator creates a new default authenticator with standard implementations.
//
// NewDefaultAuthenticator initializes an authenticator with terminal-based password reading
// and file-based key loading capabilities.
//
// Returns:
//   - Authenticator instance ready for use
func NewDefaultAuthenticator() Authenticator {
	return &defaultAuthenticator{
		passwordReader: &terminalPasswordReader{},
		keyLoader:      &fileKeyLoader{},
	}
}

// NewAuthenticator creates a new authenticator with custom implementations.
//
// NewAuthenticator allows injection of custom password readers and key loaders
// for testing or alternative authentication methods.
//
// Parameters:
//   - passwordReader: Custom implementation for password input
//   - keyLoader: Custom implementation for key loading
//
// Returns:
//   - Authenticator instance with injected dependencies
func NewAuthenticator(passwordReader PasswordReader, keyLoader KeyLoader) Authenticator {
	return &defaultAuthenticator{
		passwordReader: passwordReader,
		keyLoader:      keyLoader,
	}
}

// GetAuthMethods returns the authentication methods based on configuration.
//
// GetAuthMethods prioritizes SSH key authentication when a key path is provided,
// falling back to password authentication. Private keys are cached to improve
// performance on subsequent connections.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - config: SSH configuration containing authentication details
//
// Returns:
//   - Slice of SSH authentication methods
//   - Error if authentication setup fails
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

// terminalPasswordReader reads passwords from terminal input securely.
//
// terminalPasswordReader implements the PasswordReader interface using
// the system terminal for secure password input without echo.
type terminalPasswordReader struct{}

// ReadPassword reads a password from the terminal without displaying it.
//
// ReadPassword prompts the user with the given message and securely reads
// the password input, ensuring the password is not echoed to the terminal.
//
// Parameters:
//   - prompt: Text to display to prompt the user
//
// Returns:
//   - Password string entered by the user
//   - Error if password reading fails
func (t *terminalPasswordReader) ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Read password securely
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
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

// fileKeyLoader loads SSH private keys from the filesystem.
//
// fileKeyLoader implements the KeyLoader interface for loading SSH private keys
// from files, with secure memory handling to clear key data after parsing.
type fileKeyLoader struct{}

// LoadKey loads a private key from the specified file path.
//
// LoadKey reads an SSH private key file, parses it into a usable signer,
// and securely clears the key data from memory after parsing.
//
// Parameters:
//   - path: File system path to the SSH private key
//
// Returns:
//   - SSH signer instance for the loaded key
//   - Error if key loading or parsing fails
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
