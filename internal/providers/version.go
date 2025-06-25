// Package providers contains version information providers for superviz.io
package providers

import (
	"runtime"
	"sync"
)

// Build-time variables injected by the build system.
//
// These variables are set during compilation via ldflags and contain
// information about the build version, commit, and build environment.
var (
	// version contains the application version (set via ldflags)
	version = "dev"
	// commit contains the git commit hash (set via ldflags)
	commit = "none"
	// date contains the build timestamp (set via ldflags)
	date = "unknown"
	// builtBy contains the build system identifier (set via ldflags)
	builtBy = "unknown"
	// goVersion contains the Go compiler version used for the build
	goVersion = runtime.Version()
	// osArch contains the target operating system and architecture
	osArch = runtime.GOOS + "/" + runtime.GOARCH
)

// VersionInfo contains metadata about the compiled binary.
//
// VersionInfo provides structured access to build-time information
// including version numbers, build details, and runtime environment.
type VersionInfo struct {
	// Version is the application version string
	Version string
	// Commit is the git commit hash of the build
	Commit string
	// BuiltAt is the timestamp when the binary was built
	BuiltAt string
	// BuiltBy identifies the build system or user who created the binary
	BuiltBy string
	// GoVersion is the Go compiler version used for the build
	GoVersion string
	// OSArch is the target operating system and architecture
	OSArch string
}

var (
	// cachedVersionInfo holds the cached version information
	cachedVersionInfo VersionInfo
	// once ensures version info is initialized only once
	once sync.Once
)

// initVersionInfo initializes the cached version info with current build values.
//
// initVersionInfo is called once via sync.Once to populate the cached
// version information with values from the build-time variables.
//
// Example:
//
//	initVersionInfo() // Called automatically via sync.Once
//
// Parameters:
//   - None
//
// Returns:
//   - None (updates global cachedVersionInfo)
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

// Format returns a formatted string representation for CLI display.
//
// Format creates a human-readable, multi-line string containing all
// version information suitable for command-line output.
//
// Example:
//
//	info := GetVersionInfo()
//	fmt.Println(info.Format())
//	// Output:
//	// Version:       v1.0.0
//	// Commit:        abc123
//	// Built at:      2023-01-01T12:00:00Z
//	// Built by:      builder
//	// Go version:    go1.21
//	// OS/Arch:       linux/amd64
//
// Parameters:
//   - None (method receiver)
//
// Returns:
//   - formatted: string representation of version information
func (vi VersionInfo) Format() string {
	return "Version:       " + vi.Version +
		"\nCommit:        " + vi.Commit +
		"\nBuilt at:      " + vi.BuiltAt +
		"\nBuilt by:      " + vi.BuiltBy +
		"\nGo version:    " + vi.GoVersion +
		"\nOS/Arch:       " + vi.OSArch + "\n"
}

// VersionProvider defines the interface for providing version data.
//
// VersionProvider abstracts access to version information, enabling
// dependency injection and testing of version-related functionality.
//
// Example:
//
//	provider := NewVersionProvider()
//	info := provider.GetVersionInfo()
//	fmt.Println(info.Format())
type VersionProvider interface {
	// GetVersionInfo returns complete version information.
	//
	// GetVersionInfo provides access to all build-time metadata including
	// version numbers, commit hashes, and build environment details.
	//
	// Example:
	//	provider := NewVersionProvider()
	//	info := provider.GetVersionInfo()
	//	fmt.Printf("Version: %s\n", info.Version)
	//
	// Parameters:
	//   - None
	//
	// Returns:
	//   - info: VersionInfo containing all build and version metadata
	GetVersionInfo() VersionInfo
}

// versionProvider is the concrete provider for version data.
//
// versionProvider implements the VersionProvider interface using cached
// version information initialized once at startup.
type versionProvider struct{}

// GetVersionInfo returns the cached version information.
//
// GetVersionInfo ensures the version info is initialized and returns
// the complete metadata about the current build.
//
// Example:
//
//	provider := &versionProvider{}
//	info := provider.GetVersionInfo()
//	fmt.Printf("Version: %s\n", info.Version)
//
// Parameters:
//   - None (method receiver)
//
// Returns:
//   - info: VersionInfo containing all build and version metadata
func (p *versionProvider) GetVersionInfo() VersionInfo {
	once.Do(initVersionInfo)
	return cachedVersionInfo
}

// DefaultVersionProvider returns the singleton instance of the default version provider.
//
// DefaultVersionProvider provides access to the standard version provider
// instance for use throughout the application.
//
// Example:
//
//	provider := DefaultVersionProvider()
//	info := provider.GetVersionInfo()
//	fmt.Println(info.Format())
//
// Parameters:
//   - None
//
// Returns:
//   - provider: VersionProvider instance with default configuration
func DefaultVersionProvider() VersionProvider {
	return NewVersionProvider()
}

// NewVersionProvider creates a new default version data provider.
//
// NewVersionProvider initializes a new instance of the version provider.
// Note: Returns the same singleton instance for better performance.
//
// Example:
//
//	provider := NewVersionProvider()
//	info := provider.GetVersionInfo()
//	fmt.Printf("Version: %s\n", info.Version)
//
// Parameters:
//   - None
//
// Returns:
//   - provider: VersionProvider instance ready for use
func NewVersionProvider() VersionProvider {
	return &versionProvider{}
}

// Reset resets the singleton state for testing purposes.
//
// Reset clears the cached version information and resets the initialization
// state, allowing fresh initialization in test scenarios.
// WARNING: This should ONLY be used in tests!
//
// Example:
//
//	func TestVersionProvider(t *testing.T) {
//		defer Reset() // Clean up after test
//		// Test logic here
//	}
//
// Parameters:
//   - None
//
// Returns:
//   - None (resets global state)
func Reset() {
	once = sync.Once{}
	cachedVersionInfo = VersionInfo{}
}
