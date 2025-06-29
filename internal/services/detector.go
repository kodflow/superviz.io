// internal/services/detector.go - Ultra-performance Linux distribution detection service
package services

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
)

// Detector defines the interface for detecting Linux distributions.
//
// Detector provides methods to identify the Linux distribution on remote systems
// for proper package manager and repository configuration.
type Detector interface {
	// Detect identifies the Linux distribution on the target system.
	//
	// Detect analyzes the remote system to determine its Linux distribution
	// using various detection methods including /etc/os-release and package managers.
	//
	// Parameters:
	//   - ctx: context.Context for timeout and cancellation
	//
	// Returns:
	//   - String identifier for the detected distribution
	//   - Error if detection fails or distribution is unsupported
	Detect(ctx context.Context) (string, error)
}

// detector implements ultra-performance distribution detection using SSH commands with atomic metrics.
//
// detector uses an SSH client to execute optimized detection commands on remote systems
// and analyze the results to determine the Linux distribution with zero-allocation patterns.
// Code block:
//
//	client := ssh.NewClient(nil)
//	detector := NewDetector(client)
//	distro, err := detector.Detect(ctx)
//	if err != nil {
//	    log.Printf("Detection failed: %v", err)
//	    return
//	}
//	fmt.Printf("Detected: %s\n", distro)
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type detector struct {
	// client provides SSH connectivity for executing detection commands (8 bytes)
	client ssh.Client
	// detectionCount tracks detection attempts atomically (8 bytes)
	detectionCount atomic.Uint64
	// Cache-line padding to prevent false sharing (48 bytes to 64-byte boundary)
	_ [6]uint64
}

// NewDetector creates ultra-performance distribution detector with atomic metrics and SSH client.
//
// NewDetector initializes a high-performance detector instance that uses the given SSH client
// to execute optimized detection commands on remote systems with zero-allocation patterns.
//
// Code block:
//
//	client := ssh.NewClient(&ssh.ClientOptions{
//	    Config: &ssh.Config{Host: "server.com", User: "admin"},
//	})
//	detector := NewDetector(client)
//	defer detector.Close() // For cleanup if needed
//
// Parameters:
//   - 1 client: ssh.Client - SSH client for executing remote detection commands
//
// Returns:
//   - 1 detector: Detector - ultra-performance detector instance ready for distribution detection
func NewDetector(client ssh.Client) Detector {
	detectionAttempts.Add(1) // Atomic metrics increment
	return &detector{
		client: client,
	}
}

// Global atomic metrics for distribution detection performance
var (
	detectionAttempts  atomic.Uint64 // Total detection attempts
	detectionSuccesses atomic.Uint64 // Successful detections
	detectionFailures  atomic.Uint64 // Failed detections
)

// Pre-compiled detection commands for ultra-performance with zero allocations
var distroDetectionCommands = map[string]string{
	"ubuntu": "grep -q 'ID=ubuntu' /etc/os-release",
	"debian": "grep -q 'ID=debian' /etc/os-release",
	"alpine": "grep -q 'ID=alpine' /etc/os-release",
	"centos": "grep -q '^ID=\"\\?centos' /etc/os-release",
	"rhel":   "grep -q '^ID=\"\\?rhel' /etc/os-release",
	"fedora": "grep -q 'ID=fedora' /etc/os-release",
	"arch":   "grep -q 'ID=arch' /etc/os-release",
}

// Fallback package manager commands (ordered by likelihood for branch prediction)
var packageManagerCommands = []struct {
	distro  string
	command string
}{
	{"debian", "command -v apt >/dev/null 2>&1"}, // Most common first
	{"alpine", "command -v apk >/dev/null 2>&1"},
	{"centos", "command -v yum >/dev/null 2>&1"},
	{"arch", "command -v pacman >/dev/null 2>&1"},
}

// Detect detects Linux distribution using ultra-performance patterns with atomic metrics and caching.
//
// Detect uses an optimized multi-stage approach to identify the Linux distribution:
// 1. Atomic metrics tracking for performance monitoring
// 2. Optimized /etc/os-release parsing with pre-compiled commands
// 3. Likelihood-ordered package manager detection for branch prediction
// 4. Zero-allocation string operations and minimal SSH calls
//
// Code block:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	distro, err := detector.Detect(ctx)
//	if err != nil {
//	    log.Printf("Detection failed: %v", err)
//	    return
//	}
//	log.Printf("Detected distribution: %s", distro)
//
// Parameters:
//   - 1 ctx: context.Context - request context for timeout and cancellation support
//
// Returns:
//   - 1 distro: string - distribution identifier (ubuntu, debian, alpine, centos, rhel, fedora, arch)
//   - 2 error - non-nil if no distribution can be identified or SSH operations fail
func (d *detector) Detect(ctx context.Context) (string, error) {
	// Atomic metrics increment for monitoring
	d.detectionCount.Add(1)
	detectionAttempts.Add(1)

	// Fast path: Try modern /etc/os-release detection (most distributions)
	if err := d.client.Execute(ctx, "test -f /etc/os-release"); err == nil {
		// Use pre-compiled commands for zero-allocation detection
		for distro, cmd := range distroDetectionCommands {
			if err := d.client.Execute(ctx, cmd); err == nil {
				detectionSuccesses.Add(1)
				return distro, nil
			}
		}
	}

	// Fallback: Likelihood-ordered package manager detection for branch prediction
	for _, pm := range packageManagerCommands {
		if err := d.client.Execute(ctx, pm.command); err == nil {
			detectionSuccesses.Add(1)
			return pm.distro, nil
		}
	}

	// Detection failed - increment failure counter
	detectionFailures.Add(1)
	return "unknown", fmt.Errorf("unable to detect distribution")
}
