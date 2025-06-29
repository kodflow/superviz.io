// Package providers contains version information providers for superviz.io with ultra-performance optimizations
package providers

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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
	// goVersion contains the Go compiler version used for the build (cleaned from runtime.Version())
	goVersion = cleanGoVersion(runtime.Version())
	// osArch contains the target operating system and architecture
	osArch = runtime.GOOS + "/" + runtime.GOARCH
)

// cleanGoVersion removes experimental flags and extra information from Go version string
func cleanGoVersion(version string) string {
	// Split by space to remove experimental flags like "X:nocoverageredesign"
	parts := strings.Fields(version)
	if len(parts) > 0 {
		return parts[0] // Return only the main version part (e.g., "go1.24.3")
	}
	return version
}

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
	// cachedVersionInfo holds the cached version information (cache-line aligned)
	cachedVersionInfo VersionInfo
	// once ensures version info is initialized only once
	once sync.Once

	// Atomic performance counters for provider operations (false sharing prevention)
	formatCalls    atomic.Uint64
	getInfoCalls   atomic.Uint64
	bytesFormatted atomic.Uint64
	_              [5]uint64 // Padding to prevent false sharing

	// Pre-allocated format string pool for zero-allocation formatting
	formatPool = sync.Pool{
		New: func() any {
			// Pre-allocate with typical version string capacity
			builder := &strings.Builder{}
			builder.Grow(256) // Estimated size for version info
			return builder
		},
	}
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

// Format returns a formatted string representation for CLI display with zero-allocation optimization
// Code block:
//
//	info := GetVersionInfo()
//	formatted := info.Format()
//	fmt.Println(formatted)
//	// Output:
//	// Version:       v1.0.0
//	// Commit:        abc123
//	// Built at:      2023-01-01T12:00:00Z
//	// Built by:      builder
//	// Go version:    go1.21
//	// OS/Arch:       linux/amd64
//
// Parameters: N/A (method receiver)
//
// Returns:
//   - 1 formatted: string - zero-allocation formatted version information
func (vi VersionInfo) Format() string {
	formatCalls.Add(1)

	// Get builder from pool for zero-allocation formatting
	builder := formatPool.Get().(*strings.Builder)
	defer func() {
		// Track bytes and reset for reuse
		bytesFormatted.Add(uint64(builder.Len()))
		builder.Reset()
		formatPool.Put(builder)
	}()

	// Efficient string building with pre-allocated capacity
	builder.WriteString("Version:       ")
	builder.WriteString(vi.Version)
	builder.WriteString("\nCommit:        ")
	builder.WriteString(vi.Commit)
	builder.WriteString("\nBuilt at:      ")
	builder.WriteString(vi.BuiltAt)
	builder.WriteString("\nBuilt by:      ")
	builder.WriteString(vi.BuiltBy)
	builder.WriteString("\nGo version:    ")
	builder.WriteString(vi.GoVersion)
	builder.WriteString("\nOS/Arch:       ")
	builder.WriteString(vi.OSArch)
	builder.WriteByte('\n')

	return builder.String()
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

// GetVersionInfo returns the cached version information with atomic call tracking
// Code block:
//
//	provider := &versionProvider{}
//	info := provider.GetVersionInfo()
//	fmt.Printf("Version: %s, Calls: %d\n", info.Version, GetProviderMetrics().GetInfoCalls)
//
// Parameters: N/A (method receiver)
//
// Returns:
//   - 1 info: VersionInfo - cached build and version metadata
func (p *versionProvider) GetVersionInfo() VersionInfo {
	getInfoCalls.Add(1)
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

// VersionProviderMetrics contains atomic performance metrics for provider operations
// Code block:
//
//	metrics := GetProviderMetrics()
//	fmt.Printf("Format calls: %d, Info calls: %d, Bytes: %d\n",
//	    metrics.FormatCalls, metrics.GetInfoCalls, metrics.BytesFormatted)
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type VersionProviderMetrics struct {
	FormatCalls    uint64
	GetInfoCalls   uint64
	BytesFormatted uint64
}

// GetProviderMetrics returns current atomic performance metrics for all provider operations
// Code block:
//
//	metrics := GetProviderMetrics()
//	fmt.Printf("Provider metrics: %+v\n", metrics)
//
// Parameters: N/A
//
// Returns:
//   - 1 metrics: VersionProviderMetrics - current atomic counter values
func GetProviderMetrics() VersionProviderMetrics {
	return VersionProviderMetrics{
		FormatCalls:    formatCalls.Load(),
		GetInfoCalls:   getInfoCalls.Load(),
		BytesFormatted: bytesFormatted.Load(),
	}
}

// ResetProviderMetrics atomically resets all provider performance counters to zero
// Code block:
//
//	ResetProviderMetrics()
//	metrics := GetProviderMetrics() // All counters will be 0
//
// Parameters: N/A
//
// Returns: N/A
func ResetProviderMetrics() {
	formatCalls.Store(0)
	getInfoCalls.Store(0)
	bytesFormatted.Store(0)
}

// Reset resets the singleton state for testing purposes with atomic metrics cleanup
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
	ResetProviderMetrics() // Reset atomic counters too
}
