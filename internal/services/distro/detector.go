// internal/services/distro/detector.go
package distro

import (
	"context"
	"fmt"

	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// Detector defines the interface for detecting Linux distributions.
type Detector interface {
	Detect(ctx context.Context) (string, error)
}

// detector implements distribution detection.
type detector struct {
	client ssh.Client
}

// NewDetector creates a new distribution detector.
func NewDetector(client ssh.Client) Detector {
	return &detector{
		client: client,
	}
}

// Detect detects the Linux distribution.
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
