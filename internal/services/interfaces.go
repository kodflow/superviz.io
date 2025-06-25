// internal/services/interfaces.go - Service layer interfaces for superviz.io installation
package services

import (
	"context"
	"io"

	"github.com/kodflow/superviz.io/internal/providers"
)

// InstallServiceInterface defines the contract for installation services.
//
// InstallServiceInterface provides methods for validating configuration,
// performing installations, and retrieving installation information.
type InstallServiceInterface interface {
	// ValidateAndPrepareConfig validates and prepares the installation configuration.
	//
	// ValidateAndPrepareConfig ensures that the provided configuration is valid
	// and prepares it for the installation process, including any necessary
	// preprocessing or validation of command-line arguments.
	//
	// Parameters:
	//   - config: Installation configuration to validate and prepare
	//   - args: Command-line arguments for additional validation context
	//
	// Returns:
	//   - Error if validation fails or configuration is invalid
	ValidateAndPrepareConfig(config *providers.InstallConfig, args []string) error

	// Install performs the installation process.
	//
	// Install executes the complete installation workflow using the provided
	// configuration, writing progress and status information to the writer.
	//
	// Parameters:
	//   - ctx: context.Context for timeout and cancellation
	//   - writer: Output writer for installation progress and messages
	//   - config: Validated installation configuration
	//
	// Returns:
	//   - Error if installation fails at any stage
	Install(ctx context.Context, writer io.Writer, config *providers.InstallConfig) error

	// GetInstallInfo retrieves installation information.
	//
	// GetInstallInfo returns metadata about the installation process,
	// including version information, supported platforms, and other details.
	//
	// Returns:
	//   - InstallInfo containing installation metadata
	GetInstallInfo() providers.InstallInfo
}

// DistroDetector detects the Linux distribution on the target system.
//
// DistroDetector provides methods to identify the Linux distribution
// for proper package manager and repository configuration.
type DistroDetector interface {
	// Detect identifies the Linux distribution.
	//
	// Detect analyzes the system to determine the Linux distribution,
	// typically by examining files like /etc/os-release or distribution-specific
	// files and commands.
	//
	// Parameters:
	//   - ctx: context.Context for timeout and cancellation
	//
	// Returns:
	//   - String identifier for the detected distribution
	//   - Error if detection fails or distribution is unsupported
	Detect(ctx context.Context) (string, error)
}

// RepositorySetup handles repository setup operations for package management.
//
// RepositorySetup provides methods to configure package repositories
// on different Linux distributions to enable software installation.
type RepositorySetup interface {
	// Setup configures repositories for the specified distribution.
	//
	// Setup installs and configures the necessary package repositories
	// for the detected Linux distribution, writing progress information
	// to the provided writer.
	//
	// Parameters:
	//   - ctx: context.Context for timeout and cancellation
	//   - distro: Linux distribution identifier
	//   - writer: Output writer for setup progress and messages
	//
	// Returns:
	//   - Error if repository setup fails
	Setup(ctx context.Context, distro string, writer io.Writer) error
}
