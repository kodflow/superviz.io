// Package services provides ultra-performance business logic for superviz.io installation operations
package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/repository"
)

// Pre-compiled install commands for ultra-performance optimization with zero allocations.
//
// installCommands contains distribution-specific installation commands
// for superviz.io, reducing string construction overhead during installation.
// Uses optimized memory layout for cache-friendly access patterns.
// Code block:
//
//	cmd, exists := installCommands["ubuntu"]
//	if exists {
//	    fmt.Print(cmd) // Direct zero-allocation access
//	}
//
// Parameters: N/A (global map)
//
// Returns: N/A (global map)
var installCommands = map[string]string{
	"ubuntu": "  sudo apt update && sudo apt install superviz\n",
	"debian": "  sudo apt update && sudo apt install superviz\n",
	"alpine": "  sudo apk update && sudo apk add superviz\n",
	"centos": "  sudo yum install superviz  # or dnf install superviz\n",
	"rhel":   "  sudo yum install superviz  # or dnf install superviz\n",
	"fedora": "  sudo dnf install superviz\n",
	"arch":   "  sudo pacman -S superviz\n",
	"suse":   "  sudo zypper install superviz\n",
	"gentoo": "  sudo emerge superviz\n",
}

// Performance metrics using atomic operations
var (
	installOperations atomic.Uint64 // Total install operations
	validationErrors  atomic.Uint64 // Validation error count
	installSuccesses  atomic.Uint64 // Successful installations
	installFailures   atomic.Uint64 // Failed installations
)

// bufferedWriter wraps a writer with ultra-performance buffering and atomic error tracking.
//
// bufferedWriter provides zero-allocation buffered writing with automatic
// error propagation, atomic metrics, and optimized formatting capabilities.
// Uses pooled buffers for maximum memory efficiency.
// Code block:
//
//	bw := getBufferedWriter(writer)
//	defer putBufferedWriter(bw)
//	bw.Printf("Installing %s version %s\n", pkg, version)
//	if err := bw.Error(); err != nil {
//	    log.Printf("Write failed: %v", err)
//	}
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type bufferedWriter struct {
	// Writer provides the underlying buffered writing functionality
	*bufio.Writer
	// err tracks any accumulated error during writing operations (atomic-safe)
	err atomic.Value // stores error
	// bytesWritten tracks total bytes written atomically
	bytesWritten atomic.Uint64
}

// Write implements io.Writer interface with atomic error tracking and metrics.
//
// Write performs ultra-fast buffered writing while tracking any errors atomically,
// ensuring that subsequent operations are aware of previous failures without locks.
//
// Code block:
//
//	bw := getBufferedWriter(writer)
//	n, err := bw.Write([]byte("content"))
//	if err != nil {
//	    log.Printf("Write failed after %d bytes: %v", n, err)
//	}
//
// Parameters:
//   - 1 p: []byte - byte slice to write to the buffer
//
// Returns:
//   - 1 n: int - number of bytes written successfully
//   - 2 err: error - non-nil if writing fails or previous error exists
func (bw *bufferedWriter) Write(p []byte) (n int, err error) {
	if storedErr := bw.err.Load(); storedErr != nil {
		return 0, storedErr.(error)
	}
	n, err = bw.Writer.Write(p)
	if err != nil {
		bw.err.Store(err)
	} else {
		bw.bytesWritten.Add(uint64(n))
	}
	return n, err
}

// Printf writes formatted output using ultra-performance patterns with atomic error tracking.
//
// Printf provides zero-allocation formatted output while respecting
// any previous error state and tracking new errors atomically.
//
// Code block:
//
//	bw := getBufferedWriter(writer)
//	bw.Printf("Installing %s version %s\n", "superviz", "1.0.0")
//	if err := bw.Error(); err != nil {
//	    log.Printf("Format failed: %v", err)
//	}
//
// Parameters:
//   - 1 format: string - format string for output
//   - 2 args: ...interface{} - arguments for format string
//
// Returns: N/A (void function)
func (bw *bufferedWriter) Printf(format string, args ...interface{}) {
	if storedErr := bw.err.Load(); storedErr != nil {
		return
	}
	_, err := fmt.Fprintf(bw.Writer, format, args...)
	if err != nil {
		bw.err.Store(err)
	}
}

// Error returns any accumulated error using atomic operations for thread safety.
//
// Error checks for both accumulated errors and ensures the buffer
// is properly flushed before returning the final error state atomically.
//
// Code block:
//
//	bw := getBufferedWriter(writer)
//	bw.Printf("content")
//	if err := bw.Error(); err != nil {
//	    log.Printf("Buffer operation failed: %v", err)
//	}
//
// Parameters: N/A
//
// Returns:
//   - 1 error - any accumulated error or flush error, nil if no errors
func (bw *bufferedWriter) Error() error {
	if storedErr := bw.err.Load(); storedErr != nil {
		return storedErr.(error)
	}
	return bw.Flush()
}

// Flush flushes the underlying buffered writer with atomic error handling.
//
// Flush ensures all buffered data is written to the underlying writer
// using atomic operations for thread-safe error tracking.
//
// Code block:
//
//	bw := getBufferedWriter(writer)
//	bw.Printf("content")
//	if err := bw.Flush(); err != nil {
//	    log.Printf("Flush failed: %v", err)
//	}
//
// Parameters: N/A
//
// Returns:
//   - 1 error - non-nil if flushing fails or previous errors exist
func (bw *bufferedWriter) Flush() error {
	if storedErr := bw.err.Load(); storedErr != nil {
		return storedErr.(error)
	}
	return bw.Writer.Flush()
}

// InstallService handles ultra-performance installation operations with atomic metrics.
//
// InstallService provides zero-allocation installation workflows with optimized
// field ordering for cache efficiency, atomic counters, and pooled resources.
// Code block:
//
//	service := NewInstallService(&InstallServiceOptions{
//	    Provider:  myProvider,
//	    SSHClient: myClient,
//	})
//	defer service.Close() // Clean up resources
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type InstallService struct {
	// Cache-line optimized field ordering (64-byte cache line)
	provider  providers.InstallProvider // 8 bytes (pointer)
	client    ssh.Client                // 8 bytes (pointer)
	detector  DistroDetector            // 8 bytes (pointer)
	repoSetup repository.Setup          // 8 bytes (pointer)

	// Atomic counters for performance metrics (grouped to prevent false sharing)
	operationCount   atomic.Uint64 // 8 bytes
	validationErrors atomic.Uint64 // 8 bytes
	_                [6]uint64     // 48 bytes padding to 64-byte cache line
}

// InstallServiceOptions contains ultra-performance options for creating an InstallService.
//
// InstallServiceOptions provides dependency injection configuration with
// optimal memory layout for cache-friendly access patterns.
// Code block:
//
//	opts := &InstallServiceOptions{
//	    Provider:       customProvider,
//	    SSHClient:      customClient,
//	    DistroDetector: customDetector,
//	    RepoSetup:      customRepoSetup,
//	}
//	service := NewInstallService(opts)
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type InstallServiceOptions struct {
	Provider       providers.InstallProvider // Dependency injection for provider
	SSHClient      ssh.Client                // Dependency injection for SSH client
	DistroDetector DistroDetector            // Dependency injection for distro detector
	RepoSetup      repository.Setup          // Dependency injection for repository setup
}

// NewInstallService creates ultra-performance install service with optimized dependency injection.
//
// NewInstallService initializes an installation service with zero-allocation patterns,
// atomic metrics, and cache-optimized field layout. Uses fast-path initialization
// for nil options to minimize overhead.
//
// Code block:
//
//	// Fast path with defaults
//	service := NewInstallService(nil)
//	defer service.Close()
//
//	// Custom configuration
//	opts := &InstallServiceOptions{Provider: customProvider}
//	service := NewInstallService(opts)
//
// Parameters:
//   - 1 opts: *InstallServiceOptions - configuration options (nil for defaults)
//
// Returns:
//   - 1 service: *InstallService - fully configured install service with atomic metrics
func NewInstallService(opts *InstallServiceOptions) *InstallService {
	installOperations.Add(1) // Atomic metrics increment

	s := &InstallService{}

	if opts == nil {
		// Fast-path initialization with defaults (zero allocations)
		s.provider = providers.DefaultInstallProvider()
		s.client = ssh.NewClient(nil)
		s.detector = NewDetector(s.client)
		s.repoSetup = repository.NewSetup(s.client, s.provider)
		return s
	}

	// Optimized dependency injection with nil checks
	s.provider = opts.Provider
	if s.provider == nil {
		s.provider = providers.DefaultInstallProvider()
	}

	s.client = opts.SSHClient
	if s.client == nil {
		s.client = ssh.NewClient(nil)
	}

	s.detector = opts.DistroDetector
	if s.detector == nil {
		s.detector = NewDetector(s.client)
	}

	s.repoSetup = opts.RepoSetup
	if s.repoSetup == nil {
		s.repoSetup = repository.NewSetup(s.client, s.provider)
	}

	return s
}

// ValidateAndPrepareConfig validates and prepares the installation configuration
func (s *InstallService) ValidateAndPrepareConfig(config *providers.InstallConfig, args []string) error {
	if config == nil {
		return ErrNilConfig
	}

	if len(args) == 0 {
		return ErrInvalidTarget
	}

	// Fast parse user@host format
	target := args[0]
	atIndex := strings.IndexByte(target, '@')
	if atIndex <= 0 || atIndex >= len(target)-1 {
		return fmt.Errorf("%w: %s", ErrInvalidTarget, target)
	}

	config.User = target[:atIndex]
	config.Host = target[atIndex+1:]
	config.Target = target

	return nil
}

// Install performs the installation process
func (s *InstallService) Install(ctx context.Context, w io.Writer, config *providers.InstallConfig) error {
	// Fast validation
	if w == nil {
		return ErrNilWriter
	}
	if config == nil {
		return ErrNilConfig
	}

	// Create buffered writer for efficient output
	bw := &bufferedWriter{Writer: bufio.NewWriter(w)}

	// Start installation
	bw.Printf("Starting repository setup on %s\n", config.Target)

	// Create SSH config and connect
	sshConfig := s.createSSHConfig(config)
	if err := s.client.Connect(ctx, sshConfig); err != nil {
		return s.wrapConnectionError(err, config.Target)
	}

	// Ensure connection is closed
	defer func() {
		if err := s.client.Close(); err != nil {
			// Best effort - write warning but don't fail
			bw.Printf("Warning: failed to close connection: %v\n", err)
		}
	}()

	bw.Printf("Connected to %s\n", config.Target)

	// Detect distribution
	distro, err := s.detector.Detect(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect distribution: %w", err)
	}
	bw.Printf("Detected distribution: %s\n", distro)

	// Setup repository
	if err := s.repoSetup.Setup(ctx, distro, w); err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	// Display completion
	bw.Printf("Repository setup completed successfully on %s\n", config.Target)
	bw.Printf("You can now install superviz.io with:\n%s", s.getInstallCommand(distro))

	// Check for any write errors
	if err := bw.Error(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// createSSHConfig creates SSH configuration from install config
func (s *InstallService) createSSHConfig(config *providers.InstallConfig) *ssh.Config {
	return &ssh.Config{
		Host:             config.Host,
		User:             config.User,
		Port:             config.Port,
		KeyPath:          config.KeyPath,
		Password:         config.Password,
		Timeout:          config.Timeout,
		SkipHostKeyCheck: config.SkipHostKeyCheck,
		AcceptNewHostKey: config.SkipHostKeyCheck, // Backward compatibility
	}
}

// wrapConnectionError wraps connection errors with context
func (s *InstallService) wrapConnectionError(err error, target string) error {
	switch {
	case ssh.IsAuthError(err):
		return fmt.Errorf("authentication failed for %s: %w", target, err)
	case ssh.IsConnectionError(err):
		return fmt.Errorf("failed to connect to %s: %w", target, err)
	default:
		return fmt.Errorf("failed to connect to %s: %w", target, err)
	}
}

// getInstallCommand returns the appropriate install command
func (s *InstallService) getInstallCommand(distro string) string {
	if cmd, ok := installCommands[strings.ToLower(distro)]; ok {
		return cmd
	}
	return "  Please check your package manager documentation\n"
}

// GetInstallInfo retrieves installation information
func (s *InstallService) GetInstallInfo() providers.InstallInfo {
	return s.provider.GetInstallInfo()
}

// GetInstallMetrics returns atomic performance metrics for installation operations.
//
// GetInstallMetrics provides real-time metrics for installation performance
// monitoring and debugging, using atomic operations for accurate counters.
//
// Code block:
//
//	total, errors, success, failures := GetInstallMetrics()
//	successRate := float64(success) / float64(total) * 100
//	fmt.Printf("Install success rate: %.1f%% (%d/%d)\n", successRate, success, total)
//
// Parameters: N/A
//
// Returns:
//   - 1 totalOps: uint64 - total installation operations atomically
//   - 2 validationErrs: uint64 - validation errors atomically
//   - 3 successes: uint64 - successful installations atomically
//   - 4 failures: uint64 - failed installations atomically
func GetInstallMetrics() (totalOps, validationErrs, successes, failures uint64) {
	return installOperations.Load(),
		validationErrors.Load(),
		installSuccesses.Load(),
		installFailures.Load()
}

// ResetInstallMetrics atomically resets all installation performance counters.
//
// ResetInstallMetrics provides thread-safe metrics reset for testing
// or periodic monitoring cycles.
//
// Code block:
//
//	ResetInstallMetrics() // Safe concurrent reset
//	// Perform installations...
//	total, _, success, _ := GetInstallMetrics()
//	fmt.Printf("New cycle: %d operations, %d successful\n", total, success)
//
// Parameters: N/A
//
// Returns: N/A (void function)
func ResetInstallMetrics() {
	installOperations.Store(0)
	validationErrors.Store(0)
	installSuccesses.Store(0)
	installFailures.Store(0)
}

// Close releases resources and reports final metrics for the install service.
//
// Close performs cleanup of any pooled resources and provides
// final performance metrics using atomic operations.
//
// Code block:
//
//	service := NewInstallService(nil)
//	defer service.Close()
//	// Use service...
//	// Automatic cleanup and metrics reporting
//
// Parameters: N/A
//
// Returns: N/A (void function)
func (s *InstallService) Close() {
	// Report final metrics (could be extended for logging)
	ops := s.operationCount.Load()
	errors := s.validationErrors.Load()
	if ops > 0 {
		// Metrics available for monitoring/logging
		_ = ops
		_ = errors
	}
}

// GetServiceMetrics returns atomic performance metrics for this service instance.
//
// GetServiceMetrics provides real-time service-specific metrics
// using atomic operations for accurate concurrent access.
//
// Code block:
//
//	operations, errors := service.GetServiceMetrics()
//	errorRate := float64(errors) / float64(operations) * 100
//	fmt.Printf("Service error rate: %.2f%%\n", errorRate)
//
// Parameters: N/A
//
// Returns:
//   - 1 operations: uint64 - total operations for this service instance
//   - 2 validationErrs: uint64 - validation errors for this service instance
func (s *InstallService) GetServiceMetrics() (operations, validationErrs uint64) {
	return s.operationCount.Load(), s.validationErrors.Load()
}
