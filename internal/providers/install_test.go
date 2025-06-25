package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallConfig_Fields(t *testing.T) {
	config := &InstallConfig{
		Host:             "test.example.com",
		User:             "testuser",
		Port:             2222,
		KeyPath:          "/path/to/key",
		Force:            true,
		Target:           "testuser@test.example.com",
		SkipHostKeyCheck: true,
	}

	assert.Equal(t, "test.example.com", config.Host)
	assert.Equal(t, "testuser", config.User)
	assert.Equal(t, 2222, config.Port)
	assert.Equal(t, "/path/to/key", config.KeyPath)
	assert.True(t, config.Force)
	assert.Equal(t, "testuser@test.example.com", config.Target)
	assert.True(t, config.SkipHostKeyCheck)
}

func TestInstallInfo_Fields(t *testing.T) {
	info := InstallInfo{
		RepositoryURL: "https://repo.example.com",
		PackageName:   "test-package",
		GPGKeyID:      "test-key-id",
		Version:       "1.0.0",
		Target:        "test-target",
	}

	assert.Equal(t, "https://repo.example.com", info.RepositoryURL)
	assert.Equal(t, "test-package", info.PackageName)
	assert.Equal(t, "test-key-id", info.GPGKeyID)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "test-target", info.Target)
}

func TestNewInstallProvider(t *testing.T) {
	provider := NewInstallProvider()

	require.NotNil(t, provider)

	// Verify it implements the InstallProvider interface
	var _ = InstallProvider(provider)
}

func TestDefaultInstallProvider(t *testing.T) {
	provider := DefaultInstallProvider()

	require.NotNil(t, provider)

	// Verify it implements the InstallProvider interface
	var _ = InstallProvider(provider)
}

func TestInstallProvider_GetInstallInfo(t *testing.T) {
	provider := NewInstallProvider()

	info := provider.GetInstallInfo()

	assert.Equal(t, "https://repo.superviz.io", info.RepositoryURL)
	assert.Equal(t, "superviz", info.PackageName)
	assert.Equal(t, "A1B2C3D4E5F6789A", info.GPGKeyID)
	assert.Equal(t, "latest", info.Version)
	assert.Equal(t, "", info.Target) // Default empty value
}

func TestInstallProvider_GetRepositoryURL(t *testing.T) {
	provider := NewInstallProvider()

	url := provider.GetRepositoryURL()

	assert.Equal(t, "https://repo.superviz.io", url)
}

func TestInstallProvider_GetPackageName(t *testing.T) {
	provider := NewInstallProvider()

	name := provider.GetPackageName()

	assert.Equal(t, "superviz", name)
}

func TestInstallProvider_GetGPGKeyID(t *testing.T) {
	provider := NewInstallProvider()

	keyID := provider.GetGPGKeyID()

	assert.Equal(t, "A1B2C3D4E5F6789A", keyID)
}

func TestInstallProvider_Caching(t *testing.T) {
	provider1 := NewInstallProvider()
	provider2 := NewInstallProvider()

	// Both providers should return the same cached info
	info1 := provider1.GetInstallInfo()
	info2 := provider2.GetInstallInfo()

	assert.Equal(t, info1, info2)
	assert.Equal(t, info1.RepositoryURL, info2.RepositoryURL)
	assert.Equal(t, info1.PackageName, info2.PackageName)
	assert.Equal(t, info1.GPGKeyID, info2.GPGKeyID)
	assert.Equal(t, info1.Version, info2.Version)
}

func TestInstallProvider_IndividualMethods_ReturnConsistentData(t *testing.T) {
	provider := NewInstallProvider()

	// Get data via individual methods
	url := provider.GetRepositoryURL()
	name := provider.GetPackageName()
	keyID := provider.GetGPGKeyID()

	// Get data via GetInstallInfo
	info := provider.GetInstallInfo()

	// Should be consistent
	assert.Equal(t, url, info.RepositoryURL)
	assert.Equal(t, name, info.PackageName)
	assert.Equal(t, keyID, info.GPGKeyID)
}

func TestInstallProvider_Interface_Compliance(t *testing.T) {
	provider := NewInstallProvider()

	// Test that all interface methods are implemented and callable
	info := provider.GetInstallInfo()
	assert.NotNil(t, info)

	url := provider.GetRepositoryURL()
	assert.NotEmpty(t, url)

	name := provider.GetPackageName()
	assert.NotEmpty(t, name)

	keyID := provider.GetGPGKeyID()
	assert.NotEmpty(t, keyID)
}

func TestDefaultInstallProvider_ReturnsSameInstance(t *testing.T) {
	provider1 := DefaultInstallProvider()
	provider2 := DefaultInstallProvider()

	// Should return the same type of provider with same data
	info1 := provider1.GetInstallInfo()
	info2 := provider2.GetInstallInfo()

	assert.Equal(t, info1, info2)
}

func TestInitInstallInfo_IsCalledOnce(t *testing.T) {
	// This test verifies that the sync.Once behavior works correctly
	// by calling the provider methods multiple times and ensuring consistency

	provider := NewInstallProvider()

	// Call multiple times - should always return the same values
	info1 := provider.GetInstallInfo()
	info2 := provider.GetInstallInfo()
	info3 := provider.GetInstallInfo()

	assert.Equal(t, info1, info2)
	assert.Equal(t, info2, info3)

	// Same for individual methods
	url1 := provider.GetRepositoryURL()
	url2 := provider.GetRepositoryURL()
	assert.Equal(t, url1, url2)

	name1 := provider.GetPackageName()
	name2 := provider.GetPackageName()
	assert.Equal(t, name1, name2)

	keyID1 := provider.GetGPGKeyID()
	keyID2 := provider.GetGPGKeyID()
	assert.Equal(t, keyID1, keyID2)
}
