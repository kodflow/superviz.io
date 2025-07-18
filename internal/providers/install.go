// Package providers contains data providers for superviz.io installation operations
package providers

import (
	"sync"
	"time"
)

// InstallConfig contains configuration parameters for installation operations.
//
// InstallConfig holds all necessary configuration for performing remote installations,
// including connection details, authentication settings, and installation options.
type InstallConfig struct {
	// Host is the hostname or IP address of the target server
	Host string
	// User is the username for SSH authentication
	User string
	// Port is the SSH port number (default: 22)
	Port int
	// KeyPath is the file path to the SSH private key
	KeyPath string
	// Timeout is the maximum duration for installation operations
	Timeout time.Duration
	// Force bypasses confirmation prompts and overwrites existing installations
	Force bool
	// Target is the parsed user@host combination
	Target string // parsed from user@host
	// SkipHostKeyCheck bypasses host key verification (development only)
	SkipHostKeyCheck bool // Skip host key verification (development only)
}

// InstallInfo contains metadata about superviz.io installation operations.
//
// InstallInfo provides package information, repository details, and security
// credentials needed for installing superviz.io on target systems.
type InstallInfo struct {
	// RepositoryURL is the base URL for the superviz.io package repository
	RepositoryURL string
	// PackageName is the name of the superviz.io package
	PackageName string
	// GPGKeyID is the identifier for the GPG key used to sign packages
	GPGKeyID string
	// Version specifies the package version to install
	Version string
	// Target identifies the installation target system
	Target string
}

var (
	// cachedInstallInfo holds the cached installation information
	cachedInstallInfo InstallInfo
	// installOnce ensures installation info is initialized only once
	installOnce sync.Once
)

// initInstallInfo initializes the cached install info with default values.
//
// initInstallInfo is called once via sync.Once to populate the cached
// installation information with the current configuration values.
//
// Example:
//
//	initInstallInfo() // Called automatically via sync.Once
//
// Parameters:
//   - None
//
// Returns:
//   - None (updates global cachedInstallInfo)
func initInstallInfo() {
	cachedInstallInfo = InstallInfo{
		RepositoryURL: "https://repo.superviz.io",
		PackageName:   "superviz",
		GPGKeyID:      "A1B2C3D4E5F6789A", // Replace with actual GPG key ID
		Version:       "latest",
	}
}

// InstallProvider defines the interface for providing installation data.
//
// InstallProvider abstracts access to installation metadata and configuration,
// enabling dependency injection and testing of installation components.
//
// Example:
//
//	provider := NewInstallProvider()
//	info := provider.GetInstallInfo()
//	fmt.Printf("Repository: %s\n", info.RepositoryURL)
type InstallProvider interface {
	// GetInstallInfo returns complete installation information.
	//
	// GetInstallInfo provides access to all metadata required for
	// superviz.io installation including repository URLs and package details.
	//
	// Example:
	//	provider := NewInstallProvider()
	//	info := provider.GetInstallInfo()
	//	fmt.Printf("Package: %s\n", info.PackageName)
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - info: InstallInfo containing all installation metadata
	GetInstallInfo() InstallInfo

	// GetRepositoryURL returns the package repository URL.
	//
	// GetRepositoryURL provides the base URL for accessing the superviz.io
	// package repository for downloads and updates.
	//
	// Example:
	//	provider := NewInstallProvider()
	//	url := provider.GetRepositoryURL()
	//	fmt.Printf("Repository: %s\n", url)
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - url: string URL for the superviz.io package repository
	GetRepositoryURL() string

	// GetPackageName returns the superviz.io package name.
	//
	// GetPackageName provides the canonical package name used across
	// different package managers and distributions.
	//
	// Example:
	//	provider := NewInstallProvider()
	//	name := provider.GetPackageName()
	//	fmt.Printf("Package: %s\n", name)
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - name: string name of the superviz.io package
	GetPackageName() string

	// GetGPGKeyID returns the GPG key ID for package verification.
	//
	// GetGPGKeyID provides the identifier for the GPG key used to
	// sign superviz.io packages for integrity verification.
	//
	// Example:
	//	provider := NewInstallProvider()
	//	keyID := provider.GetGPGKeyID()
	//	fmt.Printf("GPG Key: %s\n", keyID)
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - keyID: string identifier for the package signing GPG key
	GetGPGKeyID() string
}

// installProvider is the concrete provider for installation data.
//
// installProvider implements the InstallProvider interface using cached
// installation information initialized once at startup.
type installProvider struct{}

// GetInstallInfo returns the cached installation information.
//
// GetInstallInfo ensures the installation info is initialized and returns
// the complete metadata needed for superviz.io installation.
//
// Example:
//
//	provider := &installProvider{}
//	info := provider.GetInstallInfo()
//	fmt.Printf("Repository: %s\n", info.RepositoryURL)
//
// Parameters:
//   - None (method receiver)
//
// Returns:
//   - info: InstallInfo containing all installation metadata
func (p *installProvider) GetInstallInfo() InstallInfo {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo
}

// GetRepositoryURL returns the package repository URL.
//
// GetRepositoryURL provides the base URL for the superviz.io package repository
// where installation packages can be downloaded.
//
// Example:
//
//	provider := &installProvider{}
//	url := provider.GetRepositoryURL()
//	fmt.Printf("Repository: %s\n", url)
//
// Parameters:
//   - None (method receiver)
//
// Returns:
//   - url: string URL for the superviz.io package repository
func (p *installProvider) GetRepositoryURL() string {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo.RepositoryURL
}

// GetPackageName returns the superviz.io package name.
//
// GetPackageName provides the canonical name of the superviz.io package
// as it appears in package managers.
//
// Example:
//
//	provider := &installProvider{}
//	name := provider.GetPackageName()
//	fmt.Printf("Package: %s\n", name)
//
// Parameters:
//   - None (method receiver)
//
// Returns:
//   - name: string name of the superviz.io package
func (p *installProvider) GetPackageName() string {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo.PackageName
}

// GetGPGKeyID returns the GPG key ID for package verification.
//
// GetGPGKeyID provides the identifier for the GPG key used to sign
// superviz.io packages for security verification.
//
// Example:
//
//	provider := &installProvider{}
//	keyID := provider.GetGPGKeyID()
//	fmt.Printf("GPG Key: %s\n", keyID)
//
// Parameters:
//   - None (method receiver)
//
// Returns:
//   - keyID: string identifier for the package signing GPG key
func (p *installProvider) GetGPGKeyID() string {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo.GPGKeyID
}

// DefaultInstallProvider returns the singleton instance of the default install provider.
//
// DefaultInstallProvider provides access to the standard installation provider
// instance for use throughout the application.
//
// Example:
//
//	provider := DefaultInstallProvider()
//	info := provider.GetInstallInfo()
//	fmt.Printf("Repository: %s\n", info.RepositoryURL)
//
// Parameters:
//   - None
//
// Returns:
//   - provider: InstallProvider instance with default configuration
func DefaultInstallProvider() InstallProvider {
	return NewInstallProvider()
}

// NewInstallProvider creates a new default installation data provider.
//
// NewInstallProvider initializes a new instance of the installation provider
// with standard configuration and caching behavior.
//
// Example:
//
//	provider := NewInstallProvider()
//	info := provider.GetInstallInfo()
//	fmt.Printf("Package: %s\n", info.PackageName)
//
// Parameters:
//   - None
//
// Returns:
//   - provider: InstallProvider instance ready for use
func NewInstallProvider() InstallProvider {
	return &installProvider{}
}
