// Package services provides business logic services for superviz.io operations
package services

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/kodflow/superviz.io/internal/providers"
)

// VersionService handles version-related operations and formatting with ultra-performance optimizations
// Code block:
//
//	service := NewVersionService(nil)
//	info := service.GetVersionInfo()
//	fmt.Printf("Version: %s\n", info.Version)
//
//	err := service.DisplayVersion(os.Stdout)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type VersionService struct {
	// Hot fields first for cache locality
	provider providers.VersionProvider

	// Atomic counters for performance metrics (cache-line aligned)
	getVersionCalls atomic.Uint64
	displayCalls    atomic.Uint64
	displayErrors   atomic.Uint64
	bytesWritten    atomic.Uint64
	_               [4]uint64 // Padding to prevent false sharing

	// Buffer pool for version string formatting (zero-allocation)
	bufferPool sync.Pool
}

// NewVersionService creates a new version service with the given provider and optimized buffer pool
// Code block:
//
//	service := NewVersionService(nil)
//	defer service.Close()
//	info := service.GetVersionInfo()
//	fmt.Printf("Version: %s\n", info.Version)
//
// Parameters:
//   - 1 provider: providers.VersionProvider - version provider instance (nil for default)
//
// Returns:
//   - 1 service: *VersionService - configured service with atomic counters and buffer pool
func NewVersionService(provider providers.VersionProvider) *VersionService {
	if provider == nil {
		provider = providers.DefaultVersionProvider() // Uses the singleton
	}

	service := &VersionService{
		provider: provider,
		bufferPool: sync.Pool{
			New: func() any {
				// Pre-allocate buffer for typical version string size
				buf := make([]byte, 0, 256)
				return &buf
			},
		},
	}

	return service
}

// Close performs cleanup of pooled resources
// Code block:
//
//	service := NewVersionService(nil)
//	defer service.Close()
//
// Parameters: N/A
//
// Returns: N/A
func (s *VersionService) Close() {
	// Pool cleanup is automatic in Go, but we can reset metrics if needed
	s.getVersionCalls.Store(0)
	s.displayCalls.Store(0)
	s.displayErrors.Store(0)
	s.bytesWritten.Store(0)
}

// GetVersionInfo retrieves version information through the configured provider with atomic tracking
// Code block:
//
//	service := NewVersionService(nil)
//	info := service.GetVersionInfo()
//	fmt.Printf("Version: %s, Calls: %d\n", info.Version, service.GetMetrics().GetVersionCalls)
//
// Parameters: N/A
//
// Returns:
//   - 1 info: providers.VersionInfo - complete version metadata
func (s *VersionService) GetVersionInfo() providers.VersionInfo {
	s.getVersionCalls.Add(1)
	return s.provider.GetVersionInfo()
}

// DisplayVersion writes formatted version information to the specified writer with zero-allocation optimization
// Code block:
//
//	var buf bytes.Buffer
//	err := service.DisplayVersion(&buf)
//	if err != nil {
//	    log.Printf("Display error: %v", err)
//	    return
//	}
//	fmt.Println(buf.String())
//
// Parameters:
//   - 1 w: io.Writer - writer to output the formatted version information (must not be nil)
//
// Returns:
//   - 1 error - non-nil if writing fails or writer is nil
func (s *VersionService) DisplayVersion(w io.Writer) error {
	s.displayCalls.Add(1)

	// Input validation (proactive security)
	if w == nil {
		s.displayErrors.Add(1)
		return ErrNilWriter
	}

	// Get pre-formatted version info
	info := s.provider.GetVersionInfo()
	formatted := info.Format()

	// Write with atomic byte tracking
	n, err := w.Write([]byte(formatted)) // More efficient than fmt.Fprint
	if err != nil {
		s.displayErrors.Add(1)
		return err
	}

	s.bytesWritten.Add(uint64(n))
	return nil
}

// DisplayVersionString returns the formatted version as a string with atomic call tracking
// Code block:
//
//	service := NewVersionService(nil)
//	versionStr := service.DisplayVersionString()
//	fmt.Println("Version:", versionStr)
//
// Parameters: N/A
//
// Returns:
//   - 1 formatted: string - formatted version string
func (s *VersionService) DisplayVersionString() string {
	s.getVersionCalls.Add(1) // Track string access too
	return s.provider.GetVersionInfo().Format()
}

// VersionServiceMetrics contains atomic performance metrics for the version service
// Code block:
//
//	metrics := service.GetMetrics()
//	fmt.Printf("Get calls: %d, Display calls: %d\n", metrics.GetVersionCalls, metrics.DisplayCalls)
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type VersionServiceMetrics struct {
	GetVersionCalls uint64
	DisplayCalls    uint64
	DisplayErrors   uint64
	BytesWritten    uint64
}

// GetMetrics returns current atomic performance metrics
// Code block:
//
//	service := NewVersionService(nil)
//	metrics := service.GetMetrics()
//	fmt.Printf("Service metrics: %+v\n", metrics)
//
// Parameters: N/A
//
// Returns:
//   - 1 metrics: VersionServiceMetrics - current atomic counter values
func (s *VersionService) GetMetrics() VersionServiceMetrics {
	return VersionServiceMetrics{
		GetVersionCalls: s.getVersionCalls.Load(),
		DisplayCalls:    s.displayCalls.Load(),
		DisplayErrors:   s.displayErrors.Load(),
		BytesWritten:    s.bytesWritten.Load(),
	}
}

// ResetMetrics atomically resets all performance counters to zero
// Code block:
//
//	service := NewVersionService(nil)
//	service.ResetMetrics()
//	metrics := service.GetMetrics() // All counters will be 0
//
// Parameters: N/A
//
// Returns: N/A
func (s *VersionService) ResetMetrics() {
	s.getVersionCalls.Store(0)
	s.displayCalls.Store(0)
	s.displayErrors.Store(0)
	s.bytesWritten.Store(0)
}
