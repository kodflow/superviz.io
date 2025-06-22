// internal/services/install.go
package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/distro"
	"github.com/kodflow/superviz.io/internal/services/repository"
	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// InstallService handles installation operations.
type InstallService struct {
	provider        providers.InstallProvider
	client          ssh.Client
	distroDetector  distro.Detector
	repositorySetup repository.Setup
}

// NewInstallService creates a new install service with the given provider.
func NewInstallService(provider providers.InstallProvider) *InstallService {
	if provider == nil {
		provider = providers.DefaultInstallProvider()
	}

	sshClient := ssh.NewClient()

	return &InstallService{
		provider:        provider,
		client:          sshClient,
		distroDetector:  distro.NewDetector(sshClient),
		repositorySetup: repository.NewSetup(sshClient, provider),
	}
}

// ValidateAndPrepareConfig validates and prepares the installation configuration.
func (s *InstallService) ValidateAndPrepareConfig(config *providers.InstallConfig, args []string) error {
	if config == nil {
		return ErrNilConfig
	}

	if len(args) == 0 {
		return ErrInvalidTarget
	}

	// Parse user@host format
	target := args[0]
	if !strings.Contains(target, "@") {
		return fmt.Errorf("%w: %s", ErrInvalidTarget, target)
	}

	parts := strings.SplitN(target, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("%w: %s", ErrInvalidTarget, target)
	}

	config.User = parts[0]
	config.Host = parts[1]
	config.Target = target

	return nil
}

// Install performs the installation process.
func (s *InstallService) Install(ctx context.Context, writer io.Writer, config *providers.InstallConfig) error {
	if writer == nil {
		return ErrNilWriter
	}
	if config == nil {
		return ErrNilConfig
	}

	if _, err := fmt.Fprintf(writer, "Starting repository setup on %s\n", config.Target); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Create SSH connection config
	sshConfig := &ssh.Config{
		Host:             config.Host,
		User:             config.User,
		Port:             config.Port,
		KeyPath:          config.KeyPath,
		Timeout:          config.Timeout,
		SkipHostKeyCheck: config.SkipHostKeyCheck,
	}

	// Connect to remote host
	if err := s.client.Connect(ctx, sshConfig); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", config.Target, err)
	}
	defer func() {
		if err := s.client.Disconnect(); err != nil {
			log.Printf("Warning: failed to disconnect SSH connection: %v", err)
		}
	}()

	if _, err := fmt.Fprintf(writer, "Connected to %s\n", config.Target); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Detect distribution
	distributionName, err := s.distroDetector.Detect(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect distribution: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "Detected distribution: %s\n", distributionName); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Setup repository based on distribution
	if err := s.repositorySetup.Setup(ctx, distributionName, writer); err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "Repository setup completed successfully on %s\n", config.Target); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "You can now install superviz.io with:\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Get install command for this distribution
	installCmd := s.getInstallCommand(distributionName)
	if _, err := fmt.Fprint(writer, installCmd); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	return nil
}

// getInstallCommand returns the appropriate install command for the distribution.
func (s *InstallService) getInstallCommand(distro string) string {
	switch distro {
	case "ubuntu", "debian":
		return "  sudo apt update && sudo apt install superviz\n"
	case "alpine":
		return "  sudo apk update && sudo apk add superviz\n"
	case "centos", "rhel", "fedora":
		return "  sudo yum install superviz  # or dnf install superviz\n"
	case "arch":
		return "  sudo pacman -S superviz\n"
	default:
		return "  Please check your package manager documentation\n"
	}
}

// GetInstallInfo retrieves installation information through the provider.
func (s *InstallService) GetInstallInfo() providers.InstallInfo {
	return s.provider.GetInstallInfo()
}
