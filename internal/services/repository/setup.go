// internal/services/repository/setup.go
package repository

import (
	"context"
	"fmt"
	"io"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/repository/alpine"
	"github.com/kodflow/superviz.io/internal/services/repository/arch"
	"github.com/kodflow/superviz.io/internal/services/repository/debian"
	"github.com/kodflow/superviz.io/internal/services/repository/rhel"
)

// Setup defines the interface for repository setup operations.
//
// Setup provides methods to configure package repositories on different Linux distributions
// enabling installation of superviz.io packages through native package managers.
type Setup interface {
	// Setup configures the package repository for the specified distribution.
	//
	// Setup analyzes the target distribution and configures the appropriate
	// package repository, including GPG keys, repository sources, and package lists.
	//
	// Example:
	//
	//	err := setup.Setup(ctx, "ubuntu", os.Stdout)
	//	if err != nil {
	//		return fmt.Errorf("repository setup failed: %w", err)
	//	}
	//
	// Parameters:
	//   - ctx: context.Context for timeout and cancellation
	//   - distro: string identifier of the Linux distribution (ubuntu, debian, alpine, etc.)
	//   - writer: io.Writer for output messages and progress information
	//
	// Returns:
	//   - error: Error if repository setup fails or distribution is unsupported
	Setup(ctx context.Context, distro string, writer io.Writer) error
}

// setup implements repository setup for different distributions.
//
// setup coordinates repository configuration across different Linux distributions
// by delegating to distribution-specific handlers while maintaining a unified interface.
type setup struct {
	// client provides SSH connectivity for executing setup commands on remote systems
	client ssh.Client
	// provider supplies installation configuration and repository information
	provider providers.InstallProvider
}

// NewSetup creates a new repository setup instance.
//
// NewSetup initializes a setup coordinator that uses the provided SSH client
// and install provider to configure repositories on remote systems.
//
// Example:
//
//	sshClient := ssh.NewClient()
//	provider := providers.NewInstallProvider()
//	setup := NewSetup(sshClient, provider)
//
// Parameters:
//   - client: ssh.Client for executing commands on remote systems
//   - provider: providers.InstallProvider for installation configuration
//
// Returns:
//   - setup: Setup instance ready for repository configuration
func NewSetup(client ssh.Client, provider providers.InstallProvider) Setup {
	return &setup{
		client:   client,
		provider: provider,
	}
}

// Setup sets up the repository for the specified distribution.
//
// Setup analyzes the distribution identifier and delegates to the appropriate
// distribution-specific handler for repository configuration. Supported distributions
// include Debian-based (ubuntu, debian), Alpine, RHEL-based (centos, rhel, fedora), and Arch.
//
// Example:
//
//	err := s.Setup(ctx, "ubuntu", os.Stdout)
//	if err != nil {
//		return fmt.Errorf("failed to setup Ubuntu repository: %w", err)
//	}
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - distro: string identifier for the target distribution
//   - writer: io.Writer for progress messages and command output
//
// Returns:
//   - error: Error if setup fails or distribution is unsupported
func (s *setup) Setup(ctx context.Context, distro string, writer io.Writer) error {
	switch distro {
	case "ubuntu", "debian":
		handler := debian.NewHandler(s.client)
		return handler.Setup(ctx, writer)
	case "alpine":
		handler := alpine.NewHandler(s.client)
		return handler.Setup(ctx, writer)
	case "centos", "rhel", "fedora":
		handler := rhel.NewHandler(s.client)
		return handler.Setup(ctx, writer)
	case "arch":
		handler := arch.NewHandler(s.client, s.provider)
		return handler.Setup(ctx, writer)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}
