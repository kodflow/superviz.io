package providers

import (
	"runtime"
	"sync"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
	goVersion = runtime.Version()
	osArch    = runtime.GOOS + "/" + runtime.GOARCH
)

// VersionInfo contains metadata about the compiled binary.
type VersionInfo struct {
	Version   string
	Commit    string
	BuiltAt   string
	BuiltBy   string
	GoVersion string
	OSArch    string
}

var (
	cachedVersionInfo VersionInfo
	once              sync.Once
)

// initVersionInfo initializes the cached version info once
func initVersionInfo() {
	cachedVersionInfo = VersionInfo{
		Version:   version,
		Commit:    commit,
		BuiltAt:   date,
		BuiltBy:   builtBy,
		GoVersion: goVersion,
		OSArch:    osArch,
	}
}

// Format returns a formatted string for CLI display.
func (vi VersionInfo) Format() string {
	return "Version:       " + vi.Version +
		"\nCommit:        " + vi.Commit +
		"\nBuilt at:      " + vi.BuiltAt +
		"\nBuilt by:      " + vi.BuiltBy +
		"\nGo version:    " + vi.GoVersion +
		"\nOS/Arch:       " + vi.OSArch + "\n"
}

// VersionProvider defines the interface for providing version data.
type VersionProvider interface {
	GetVersionInfo() VersionInfo
}

// versionProvider is the concrete provider for version data.
type versionProvider struct{}

// GetVersionInfo returns the cached version information.
func (p *versionProvider) GetVersionInfo() VersionInfo {
	once.Do(initVersionInfo)
	return cachedVersionInfo
}

// DefaultVersionProvider returns the singleton instance of the default version provider.
func DefaultVersionProvider() VersionProvider {
	return NewVersionProvider()
}

// NewVersionProvider creates a new default version data provider.
// Note: Returns the same singleton instance for better performance.
func NewVersionProvider() VersionProvider {
	return &versionProvider{}
}

// Reset resets the singleton state for testing purposes.
// WARNING: This should ONLY be used in tests!
func Reset() {
	once = sync.Once{}
	cachedVersionInfo = VersionInfo{}
}
