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

// InstallServiceOptions contains options for creating an InstallService.
type InstallServiceOptions struct {
	Provider       providers.InstallProvider
	SSHClient      ssh.Client
	DistroDetector distro.Detector
	RepoSetup      repository.Setup
}

// NewInstallService creates a new install service with the given options.
func NewInstallService(opts *InstallServiceOptions) *InstallService {
	if opts == nil {
		opts = &InstallServiceOptions{}
	}

	// Use default implementations if not provided
	if opts.Provider == nil {
		opts.Provider = providers.DefaultInstallProvider()
	}

	if opts.SSHClient == nil {
		opts.SSHClient = ssh.NewClient(nil)
	}

	if opts.DistroDetector == nil {
		opts.DistroDetector = distro.NewDetector(opts.SSHClient)
	}

	if opts.RepoSetup == nil {
		opts.RepoSetup = repository.NewSetup(opts.SSHClient, opts.Provider)
	}

	return &InstallService{
		provider:        opts.Provider,
		client:          opts.SSHClient,
		distroDetector:  opts.DistroDetector,
		repositorySetup: opts.RepoSetup,
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
	// Validate inputs
	if err := s.validateInputs(writer, config); err != nil {
		return err
	}

	// Start installation process
	if err := s.writeOutput(writer, "Starting repository setup on %s\n", config.Target); err != nil {
		return err
	}

	// Create SSH configuration
	sshConfig := s.createSSHConfig(config)

	// Connect to remote host
	if err := s.connectToHost(ctx, writer, sshConfig, config.Target); err != nil {
		return err
	}
	defer s.disconnect()

	// Detect distribution
	distributionName, err := s.detectDistribution(ctx, writer)
	if err != nil {
		return err
	}

	// Setup repository
	if err := s.setupRepository(ctx, writer, distributionName); err != nil {
		return err
	}

	// Display completion message and install command
	return s.displayCompletionInfo(writer, config.Target, distributionName)
}

// validateInputs validates the input parameters.
func (s *InstallService) validateInputs(writer io.Writer, config *providers.InstallConfig) error {
	if writer == nil {
		return ErrNilWriter
	}
	if config == nil {
		return ErrNilConfig
	}
	return nil
}

// createSSHConfig creates SSH configuration from install config.
func (s *InstallService) createSSHConfig(config *providers.InstallConfig) *ssh.Config {
	// Create base configuration
	sshConfig := ssh.DefaultConfig()

	// Apply install config values
	sshConfig.Host = config.Host
	sshConfig.User = config.User
	sshConfig.Port = config.Port
	sshConfig.KeyPath = config.KeyPath
	sshConfig.Timeout = config.Timeout
	sshConfig.SkipHostKeyCheck = config.SkipHostKeyCheck

	// Set AcceptNewHostKey based on SkipHostKeyCheck for backward compatibility
	if config.SkipHostKeyCheck {
		sshConfig.AcceptNewHostKey = true
	}

	return sshConfig
}

// connectToHost establishes SSH connection to the target host.
func (s *InstallService) connectToHost(ctx context.Context, writer io.Writer, sshConfig *ssh.Config, target string) error {
	if err := s.client.Connect(ctx, sshConfig); err != nil {
		// Check if it's an authentication error
		if ssh.IsAuthError(err) {
			return fmt.Errorf("authentication failed for %s: %w", target, err)
		}
		// Check if it's a connection error
		if ssh.IsConnectionError(err) {
			return fmt.Errorf("failed to connect to %s: %w", target, err)
		}
		// Generic error
		return fmt.Errorf("failed to connect to %s: %w", target, err)
	}

	return s.writeOutput(writer, "Connected to %s\n", target)
}

// detectDistribution detects the remote host's distribution.
func (s *InstallService) detectDistribution(ctx context.Context, writer io.Writer) (string, error) {
	distributionName, err := s.distroDetector.Detect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to detect distribution: %w", err)
	}

	if err := s.writeOutput(writer, "Detected distribution: %s\n", distributionName); err != nil {
		return "", err
	}

	return distributionName, nil
}

// setupRepository sets up the repository on the remote host.
func (s *InstallService) setupRepository(ctx context.Context, writer io.Writer, distributionName string) error {
	if err := s.repositorySetup.Setup(ctx, distributionName, writer); err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}
	return nil
}

// displayCompletionInfo displays completion message and install instructions.
func (s *InstallService) displayCompletionInfo(writer io.Writer, target, distributionName string) error {
	if err := s.writeOutput(writer, "Repository setup completed successfully on %s\n", target); err != nil {
		return err
	}

	if err := s.writeOutput(writer, "You can now install superviz.io with:\n"); err != nil {
		return err
	}

	// Get and display install command
	installCmd := s.getInstallCommand(distributionName)
	if _, err := fmt.Fprint(writer, installCmd); err != nil {
		return fmt.Errorf("failed to write install command: %w", err)
	}

	return nil
}

// disconnect closes the SSH connection with error logging.
func (s *InstallService) disconnect() {
	if err := s.client.Close(); err != nil {
		log.Printf("Warning: failed to disconnect SSH connection: %v", err)
	}
}

// writeOutput writes formatted output to the writer.
func (s *InstallService) writeOutput(writer io.Writer, format string, args ...interface{}) error {
	if _, err := fmt.Fprintf(writer, format, args...); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}
	return nil
}

// getInstallCommand returns the appropriate install command for the distribution.
func (s *InstallService) getInstallCommand(distro string) string {
	commands := map[string]string{
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

	if cmd, exists := commands[strings.ToLower(distro)]; exists {
		return cmd
	}

	return "  Please check your package manager documentation\n"
}

// GetInstallInfo retrieves installation information through the provider.
func (s *InstallService) GetInstallInfo() providers.InstallInfo {
	return s.provider.GetInstallInfo()
}
