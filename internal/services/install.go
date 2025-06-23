// internal/services/install.go
package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services/repository"
)

// Pre-compiled install commands for performance
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

// bufferedWriter wraps a writer with buffering and error tracking
type bufferedWriter struct {
	*bufio.Writer
	err error
}

// Write implements io.Writer with error tracking
func (bw *bufferedWriter) Write(p []byte) (n int, err error) {
	if bw.err != nil {
		return 0, bw.err
	}
	n, err = bw.Writer.Write(p)
	if err != nil {
		bw.err = err
	}
	return n, err
}

// Printf writes formatted output
func (bw *bufferedWriter) Printf(format string, args ...interface{}) {
	if bw.err != nil {
		return
	}
	_, bw.err = fmt.Fprintf(bw.Writer, format, args...)
}

// Error returns any accumulated error
func (bw *bufferedWriter) Error() error {
	if bw.err != nil {
		return bw.err
	}
	return bw.Flush()
}

// InstallService handles installation operations
type InstallService struct {
	provider  providers.InstallProvider
	client    ssh.Client
	detector  DistroDetector
	repoSetup repository.Setup
}

// InstallServiceOptions contains options for creating an InstallService
type InstallServiceOptions struct {
	Provider       providers.InstallProvider
	SSHClient      ssh.Client
	DistroDetector DistroDetector
	RepoSetup      repository.Setup
}

// NewInstallService creates a new install service with the given options
func NewInstallService(opts *InstallServiceOptions) *InstallService {
	s := &InstallService{}

	if opts == nil {
		// Fast initialization with defaults
		s.provider = providers.DefaultInstallProvider()
		s.client = ssh.NewClient(nil)
		s.detector = NewDetector(s.client)
		s.repoSetup = repository.NewSetup(s.client, s.provider)
		return s
	}

	// Use provided or default implementations
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
