// internal/services/detector.go - Linux distribution detection service
package services

import (
	"context"
	"fmt"

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
	//   - ctx: Context for timeout and cancellation
	//
	// Returns:
	//   - String identifier for the detected distribution
	//   - Error if detection fails or distribution is unsupported
	Detect(ctx context.Context) (string, error)
}

// detector implements distribution detection using SSH commands.
//
// detector uses an SSH client to execute detection commands on remote systems
// and analyze the results to determine the Linux distribution.
type detector struct {
	// client provides SSH connectivity for executing detection commands
	client ssh.Client
}

// NewDetector creates a new distribution detector with the provided SSH client.
//
// NewDetector initializes a detector instance that uses the given SSH client
// to execute distribution detection commands on remote systems.
//
// Parameters:
//   - client: SSH client for executing remote commands
//
// Returns:
//   - Detector instance ready for distribution detection
func NewDetector(client ssh.Client) Detector {
	return &detector{
		client: client,
	}
}

// Detect detects the Linux distribution on the connected remote system.
//
// Detect uses a multi-stage approach to identify the Linux distribution:
// 1. First, it attempts to read /etc/os-release for modern distributions
// 2. Then it checks for specific distribution identifiers
// 3. Finally, it falls back to detecting package managers
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//
// Returns:
//   - String identifier for the detected distribution (ubuntu, debian, alpine, centos, rhel, fedora, arch)
//   - Error if no distribution can be identified
func (d *detector) Detect(ctx context.Context) (string, error) {
	// Try to detect using /etc/os-release (most modern distributions)
	if err := d.client.Execute(ctx, "test -f /etc/os-release"); err == nil {
		// Check for specific distributions in order of preference
		distroChecks := map[string]string{
			"ubuntu": "grep -q 'ID=ubuntu' /etc/os-release",
			"debian": "grep -q 'ID=debian' /etc/os-release",
			"alpine": "grep -q 'ID=alpine' /etc/os-release",
			"centos": "grep -q '^ID=\"\\?centos' /etc/os-release",
			"rhel":   "grep -q '^ID=\"\\?rhel' /etc/os-release",
			"fedora": "grep -q 'ID=fedora' /etc/os-release",
			"arch":   "grep -q 'ID=arch' /etc/os-release",
		}

		for distro, cmd := range distroChecks {
			if err := d.client.Execute(ctx, cmd); err == nil {
				return distro, nil
			}
		}
	}

	// Fallback: check for package managers
	if err := d.client.Execute(ctx, "command -v apt >/dev/null 2>&1"); err == nil {
		return "debian", nil // Generic Debian-based
	}
	if err := d.client.Execute(ctx, "command -v apk >/dev/null 2>&1"); err == nil {
		return "alpine", nil
	}
	if err := d.client.Execute(ctx, "command -v yum >/dev/null 2>&1"); err == nil {
		return "centos", nil // Generic RHEL-based
	}
	if err := d.client.Execute(ctx, "command -v pacman >/dev/null 2>&1"); err == nil {
		return "arch", nil
	}

	return "unknown", fmt.Errorf("unable to detect distribution")
}
