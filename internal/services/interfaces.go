// internal/services/interfaces.go
package services

import (
	"context"
	"io"

	"github.com/kodflow/superviz.io/internal/providers"
)

// InstallServiceInterface defines the contract for installation services
type InstallServiceInterface interface {
	// ValidateAndPrepareConfig validates and prepares the installation configuration
	ValidateAndPrepareConfig(config *providers.InstallConfig, args []string) error

	// Install performs the installation process
	Install(ctx context.Context, writer io.Writer, config *providers.InstallConfig) error

	// GetInstallInfo retrieves installation information
	GetInstallInfo() providers.InstallInfo
}

// DistroDetector detects the Linux distribution
type DistroDetector interface {
	Detect(ctx context.Context) (string, error)
}

// RepositorySetup handles repository setup operations
type RepositorySetup interface {
	Setup(ctx context.Context, distro string, writer io.Writer) error
}
