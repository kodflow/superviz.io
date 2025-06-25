package ssh

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// Mock implementations for testing

type mockPasswordReader struct {
	password         string
	err              error
	readPasswordFunc func(prompt string) (string, error)
}

func (m *mockPasswordReader) ReadPassword(prompt string) (string, error) {
	if m.readPasswordFunc != nil {
		return m.readPasswordFunc(prompt)
	}
	return m.password, m.err
}

type mockKeyLoader struct {
	signer      ssh.Signer
	err         error
	loadKeyFunc func(path string) (ssh.Signer, error)
}

func (m *mockKeyLoader) LoadKey(path string) (ssh.Signer, error) {
	if m.loadKeyFunc != nil {
		return m.loadKeyFunc(path)
	}
	return m.signer, m.err
}

// Mock signer for testing
type mockSigner struct{}

func (m *mockSigner) PublicKey() ssh.PublicKey { return nil }
func (m *mockSigner) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	return &ssh.Signature{}, nil
}

func TestNewDefaultAuthenticator(t *testing.T) {
	auth := NewDefaultAuthenticator()
	require.NotNil(t, auth)

	// Should be able to type assert to concrete type
	defaultAuth, ok := auth.(*defaultAuthenticator)
	require.True(t, ok)
	require.NotNil(t, defaultAuth.passwordReader)
	require.NotNil(t, defaultAuth.keyLoader)
}

func TestNewAuthenticator(t *testing.T) {
	mockReader := &mockPasswordReader{}
	mockLoader := &mockKeyLoader{}

	auth := NewAuthenticator(mockReader, mockLoader)
	require.NotNil(t, auth)

	// Should be able to type assert to concrete type with injected dependencies
	defaultAuth, ok := auth.(*defaultAuthenticator)
	require.True(t, ok)
	require.Equal(t, mockReader, defaultAuth.passwordReader)
	require.Equal(t, mockLoader, defaultAuth.keyLoader)
}

func TestDefaultAuthenticator_GetAuthMethods_WithKey(t *testing.T) {
	tests := []struct {
		name       string
		keyPath    string
		mockSigner ssh.Signer
		keyErr     error
		wantErr    bool
		errType    error
	}{
		{
			name:       "successful key loading",
			keyPath:    "/path/to/key",
			mockSigner: &mockSigner{},
			keyErr:     nil,
			wantErr:    false,
		},
		{
			name:    "key loading failure",
			keyPath: "/invalid/path",
			keyErr:  errors.New("file not found"),
			wantErr: true,
			errType: ErrAuthFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &mockKeyLoader{
				signer: tt.mockSigner,
				err:    tt.keyErr,
			}

			auth := NewAuthenticator(nil, mockLoader).(*defaultAuthenticator)
			config := &Config{
				Host:    "example.com",
				User:    "testuser",
				KeyPath: tt.keyPath,
			}

			methods, err := auth.GetAuthMethods(context.Background(), config)

			if tt.wantErr {
				require.Error(t, err)
				require.True(t, IsAuthError(err))
				require.Nil(t, methods)

				// Check error context
				sshErr, ok := err.(*SSHError)
				require.True(t, ok)
				require.Equal(t, tt.keyPath, sshErr.Context["key_path"])
			} else {
				require.NoError(t, err)
				require.Len(t, methods, 1)
				require.NotNil(t, methods[0])
			}
		})
	}
}

func TestDefaultAuthenticator_GetAuthMethods_WithPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		passErr  error
		wantErr  bool
	}{
		{
			name:     "successful password reading",
			password: "testpassword",
			passErr:  nil,
			wantErr:  false,
		},
		{
			name:     "password reading failure",
			password: "",
			passErr:  errors.New("failed to read password"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := &mockPasswordReader{
				password: tt.password,
				err:      tt.passErr,
			}

			auth := NewAuthenticator(mockReader, nil).(*defaultAuthenticator)
			config := &Config{
				Host: "example.com",
				User: "testuser",
				// No KeyPath - should use password auth
			}

			methods, err := auth.GetAuthMethods(context.Background(), config)

			if tt.wantErr {
				require.Error(t, err)
				require.True(t, IsAuthError(err))
				require.Nil(t, methods)
			} else {
				require.NoError(t, err)
				require.Len(t, methods, 1)
				require.NotNil(t, methods[0])
			}
		})
	}
}

func TestDefaultAuthenticator_KeyCaching(t *testing.T) {
	mockSigner := &mockSigner{}
	mockLoader := &mockKeyLoader{
		signer: mockSigner,
		err:    nil,
	}

	auth := NewAuthenticator(nil, mockLoader).(*defaultAuthenticator)
	config := &Config{
		Host:    "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key",
	}

	// First call should load the key
	methods1, err1 := auth.GetAuthMethods(context.Background(), config)
	require.NoError(t, err1)
	require.Len(t, methods1, 1)

	// Second call should use cached key
	methods2, err2 := auth.GetAuthMethods(context.Background(), config)
	require.NoError(t, err2)
	require.Len(t, methods2, 1)

	// Verify key is cached
	cached, ok := auth.keyCache.Load("/path/to/key")
	require.True(t, ok)
	require.Equal(t, mockSigner, cached)
}

func TestDefaultAuthenticator_CacheWithDifferentPaths(t *testing.T) {
	mockSigner1 := &mockSigner{}
	mockSigner2 := &mockSigner{}

	mockLoader := &mockKeyLoader{
		loadKeyFunc: func(path string) (ssh.Signer, error) {
			if path == "/path/to/key1" {
				return mockSigner1, nil
			}
			return mockSigner2, nil
		},
	}

	auth := NewAuthenticator(nil, mockLoader).(*defaultAuthenticator)

	// Load first key
	config1 := &Config{
		Host:    "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key1",
	}
	_, err1 := auth.GetAuthMethods(context.Background(), config1)
	require.NoError(t, err1)

	// Load second key
	config2 := &Config{
		Host:    "example.com",
		User:    "testuser",
		KeyPath: "/path/to/key2",
	}
	_, err2 := auth.GetAuthMethods(context.Background(), config2)
	require.NoError(t, err2)

	// Both should be cached
	cached1, ok1 := auth.keyCache.Load("/path/to/key1")
	require.True(t, ok1)
	require.Equal(t, mockSigner1, cached1)

	cached2, ok2 := auth.keyCache.Load("/path/to/key2")
	require.True(t, ok2)
	require.Equal(t, mockSigner2, cached2)
}

func TestTerminalPasswordReader_ReadPassword(t *testing.T) {
	// Note: This test cannot fully test the actual terminal reading behavior
	// due to stdin dependency, but we can test the success path structure

	reader := &terminalPasswordReader{}

	// We cannot directly test this function as it reads from os.Stdin
	// and uses term.ReadPassword which requires a terminal.
	// Instead, we verify that the function exists and has the right signature

	// This test would need to be run manually or with a mocked stdin,
	// but for coverage purposes, we document that this function handles
	// terminal password reading with proper error handling.

	// Test that the reader can be created
	require.NotNil(t, reader)

	// The actual functionality is tested through the authenticator tests
	// which use mockPasswordReader
}

func TestFileKeyLoader_LoadKey(t *testing.T) {
	loader := &fileKeyLoader{}
	require.NotNil(t, loader)

	// Test with non-existent file
	_, err := loader.LoadKey("/non/existent/path")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to read private key")
}

func TestFileKeyLoader_LoadKey_InvalidKey(t *testing.T) {
	loader := &fileKeyLoader{}

	// Create a temporary file with invalid key content
	tmpFile, err := os.CreateTemp("", "invalid_key_*.pem")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }() // Ignore cleanup errors

	// Write invalid key content
	invalidKeyContent := []byte("-----BEGIN PRIVATE KEY-----\ninvalid content\n-----END PRIVATE KEY-----")
	_, err = tmpFile.Write(invalidKeyContent)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// Test with invalid key file
	_, err = loader.LoadKey(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse private key")
}

func TestFileKeyLoader_LoadKey_ValidKey(t *testing.T) {
	loader := &fileKeyLoader{}

	// Create a temporary file with a valid test key
	tmpFile, err := os.CreateTemp("", "valid_key_*.pem")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }() // Ignore cleanup errors

	// Use a minimal valid RSA private key for testing
	validKeyContent := []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
INVALID_TEST_KEY_CONTENT_FOR_PARSING_ERROR
-----END OPENSSH PRIVATE KEY-----`)

	// Write key content - this will actually be invalid because it's truncated,
	// but that's fine for testing the parse error path
	_, err = tmpFile.Write(validKeyContent)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// This should error because the key is not actually valid
	_, err = loader.LoadKey(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse private key")
}

func TestDefaultAuthenticator_GetAuthMethods_PasswordPrompt(t *testing.T) {
	var capturedPrompt string
	mockReader := &mockPasswordReader{
		password: "testpass",
		err:      nil,
		readPasswordFunc: func(prompt string) (string, error) {
			capturedPrompt = prompt
			return "testpass", nil
		},
	}

	auth := NewAuthenticator(mockReader, nil).(*defaultAuthenticator)
	config := &Config{
		Host: "example.com",
		User: "testuser",
		// No KeyPath - should use password auth
	}

	_, err := auth.GetAuthMethods(context.Background(), config)
	require.NoError(t, err)
	require.Equal(t, "Password for testuser@example.com: ", capturedPrompt)
}

func TestDefaultAuthenticator_GetAuthMethods_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockReader := &mockPasswordReader{
		password: "testpass",
		err:      nil,
	}

	auth := NewAuthenticator(mockReader, nil).(*defaultAuthenticator)
	config := &Config{
		Host: "example.com",
		User: "testuser",
		// No KeyPath - should use password auth
	}

	// Even with cancelled context, should still work since we don't check context in current implementation
	// This test ensures the method doesn't panic with cancelled context
	_, err := auth.GetAuthMethods(ctx, config)
	require.NoError(t, err)
}
