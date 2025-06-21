package providers

import (
	"sync"
	"time"
)

// InstallConfig contains configuration for installation operations.
type InstallConfig struct {
	Host    string
	User    string
	Port    int
	KeyPath string
	Timeout time.Duration
	Force   bool
	Target  string // parsed from user@host
}

// InstallInfo contains metadata about installation operations.
type InstallInfo struct {
	RepositoryURL string
	PackageName   string
	Version       string
	Target        string
}

var (
	cachedInstallInfo InstallInfo
	installOnce       sync.Once
)

// initInstallInfo initializes the cached install info once
func initInstallInfo() {
	cachedInstallInfo = InstallInfo{
		RepositoryURL: "https://repo.superviz.io",
		PackageName:   "superviz",
		Version:       "latest",
	}
}

// InstallProvider defines the interface for providing installation data.
type InstallProvider interface {
	GetInstallInfo() InstallInfo
	GetRepositoryURL() string
	GetPackageName() string
}

// installProvider is the concrete provider for installation data.
type installProvider struct{}

// GetInstallInfo returns the cached installation information.
func (p *installProvider) GetInstallInfo() InstallInfo {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo
}

// GetRepositoryURL returns the repository URL.
func (p *installProvider) GetRepositoryURL() string {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo.RepositoryURL
}

// GetPackageName returns the package name.
func (p *installProvider) GetPackageName() string {
	installOnce.Do(initInstallInfo)
	return cachedInstallInfo.PackageName
}

// DefaultInstallProvider returns the singleton instance of the default install provider.
func DefaultInstallProvider() InstallProvider {
	return NewInstallProvider()
}

// NewInstallProvider creates a new default installation data provider.
func NewInstallProvider() InstallProvider {
	return &installProvider{}
}
