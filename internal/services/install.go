package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// InstallService handles installation operations.
type InstallService struct {
	provider providers.InstallProvider
	client   ssh.Client
}

// NewInstallService creates a new install service with the given provider.
func NewInstallService(provider providers.InstallProvider) *InstallService {
	if provider == nil {
		provider = providers.DefaultInstallProvider()
	}
	return &InstallService{
		provider: provider,
		client:   ssh.NewClient(),
	}
}

// ValidateAndPrepareConfig validates and prepares the installation configuration.
func (s *InstallService) ValidateAndPrepareConfig(config *providers.InstallConfig, args []string, errWriter io.Writer) error {
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
		Host:    config.Host,
		User:    config.User,
		Port:    config.Port,
		KeyPath: config.KeyPath,
		Timeout: config.Timeout,
	}

	// Connect to remote host
	if err := s.client.Connect(ctx, sshConfig); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", config.Target, err)
	}
	defer func() {
		if err := s.client.Disconnect(); err != nil {
			// Log the error but don't fail the operation
			log.Printf("Warning: failed to disconnect SSH connection: %v", err)
		}
	}()

	if _, err := fmt.Fprintf(writer, "Connected to %s\n", config.Target); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Detect distribution
	distro, err := s.detectDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect distribution: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "Detected distribution: %s\n", distro); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// Setup repository based on distribution
	if err := s.setupRepository(ctx, distro, config, writer); err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "Repository setup completed successfully on %s\n", config.Target); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	if _, err := fmt.Fprintf(writer, "You can now install superviz.io with:\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	var installCmd string
	switch distro {
	case "ubuntu", "debian":
		installCmd = "  sudo apt update && sudo apt install superviz\n"
	case "alpine":
		installCmd = "  sudo apk update && sudo apk add superviz\n"
	case "centos", "rhel", "fedora":
		installCmd = "  sudo yum install superviz  # or dnf install superviz\n"
	case "arch":
		installCmd = "  sudo pacman -S superviz\n"
	default:
		installCmd = "  Please check your package manager documentation\n"
	}

	if _, err := fmt.Fprint(writer, installCmd); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	return nil
}

// detectDistribution detects the Linux distribution.
func (s *InstallService) detectDistribution(ctx context.Context) (string, error) {
	// Try to detect using /etc/os-release (most modern distributions)
	if err := s.client.Execute(ctx, "test -f /etc/os-release"); err == nil {
		// Check for specific distributions in order of preference
		distroChecks := map[string]string{
			"ubuntu": "grep -q 'ID=ubuntu' /etc/os-release",
			"debian": "grep -q 'ID=debian' /etc/os-release",
			"alpine": "grep -q 'ID=alpine' /etc/os-release",
			"centos": "grep -q 'ID.*centos' /etc/os-release",
			"rhel":   "grep -q 'ID.*rhel' /etc/os-release",
			"fedora": "grep -q 'ID=fedora' /etc/os-release",
			"arch":   "grep -q 'ID=arch' /etc/os-release",
		}

		for distro, cmd := range distroChecks {
			if err := s.client.Execute(ctx, cmd); err == nil {
				return distro, nil
			}
		}
	}

	// Fallback: check for package managers
	if err := s.client.Execute(ctx, "command -v apt >/dev/null 2>&1"); err == nil {
		return "debian", nil // Generic Debian-based
	}
	if err := s.client.Execute(ctx, "command -v apk >/dev/null 2>&1"); err == nil {
		return "alpine", nil
	}
	if err := s.client.Execute(ctx, "command -v yum >/dev/null 2>&1"); err == nil {
		return "centos", nil // Generic RHEL-based
	}
	if err := s.client.Execute(ctx, "command -v pacman >/dev/null 2>&1"); err == nil {
		return "arch", nil
	}

	return "unknown", fmt.Errorf("unable to detect distribution")
}

// setupRepository sets up the superviz.io repository for the detected distribution.
func (s *InstallService) setupRepository(ctx context.Context, distro string, config *providers.InstallConfig, writer io.Writer) error {
	switch distro {
	case "ubuntu", "debian":
		return s.setupDebianRepository(ctx, writer)
	case "alpine":
		return s.setupAlpineRepository(ctx, writer)
	case "centos", "rhel", "fedora":
		return s.setupRHELRepository(ctx, writer)
	case "arch":
		return s.setupArchRepository(ctx, writer)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}

// setupDebianRepository sets up the repository for Debian/Ubuntu systems.
func (s *InstallService) setupDebianRepository(ctx context.Context, writer io.Writer) error {
	commands := []string{
		// Install required packages
		"apt update",
		"apt install -y curl gnupg lsb-release",

		// Add GPG key
		"curl -fsSL https://repo.superviz.io/gpg | gpg --dearmor -o /usr/share/keyrings/superviz.gpg",

		// Add repository
		`echo "deb [signed-by=/usr/share/keyrings/superviz.gpg] https://repo.superviz.io/apt $(lsb_release -cs) main" > /etc/apt/sources.list.d/superviz.list`,

		// Update package list
		"apt update",
	}

	if _, err := fmt.Fprintf(writer, "Setting up APT repository...\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}
	return s.executeCommands(ctx, commands, writer)
}

// setupAlpineRepository sets up the repository for Alpine systems.
func (s *InstallService) setupAlpineRepository(ctx context.Context, writer io.Writer) error {
	commands := []string{
		// Add repository
		"echo 'https://repo.superviz.io/alpine/v$(cat /etc/alpine-release | cut -d'.' -f1-2)/main' >> /etc/apk/repositories",

		// Add public key
		"wget -O /etc/apk/keys/superviz.rsa.pub https://repo.superviz.io/alpine/superviz.rsa.pub",

		// Update package index
		"apk update",
	}

	if _, err := fmt.Fprintf(writer, "Setting up APK repository...\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}
	return s.executeCommands(ctx, commands, writer)
}

// setupRHELRepository sets up the repository for RHEL/CentOS/Fedora systems.
func (s *InstallService) setupRHELRepository(ctx context.Context, writer io.Writer) error {
	repoContent := `[superviz]
name=Superviz.io Repository
baseurl=https://repo.superviz.io/rpm/
enabled=1
gpgcheck=1
gpgkey=https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz`

	commands := []string{
		// Create repository file
		fmt.Sprintf("cat > /etc/yum.repos.d/superviz.repo << 'EOF'\n%s\nEOF", repoContent),

		// Import GPG key
		"rpm --import https://repo.superviz.io/rpm/RPM-GPG-KEY-superviz",

		// Update package cache
		"yum clean all || dnf clean all",
	}

	if _, err := fmt.Fprintf(writer, "Setting up YUM/DNF repository...\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}
	return s.executeCommands(ctx, commands, writer)
}

// setupArchRepository sets up the repository for Arch systems.
func (s *InstallService) setupArchRepository(ctx context.Context, writer io.Writer) error {
	commands := []string{
		// Add repository to pacman.conf
		"echo '' >> /etc/pacman.conf",
		"echo '[superviz]' >> /etc/pacman.conf",
		"echo 'Server = https://repo.superviz.io/arch/$arch' >> /etc/pacman.conf",

		// Import key
		"pacman-key --recv-keys SUPERVIZ_KEY_ID",
		"pacman-key --lsign-key SUPERVIZ_KEY_ID",

		// Update package database
		"pacman -Sy",
	}

	if _, err := fmt.Fprintf(writer, "Setting up Pacman repository...\n"); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}
	return s.executeCommands(ctx, commands, writer)
}

// executeCommands executes a list of commands in sequence.
func (s *InstallService) executeCommands(ctx context.Context, commands []string, writer io.Writer) error {
	for i, cmd := range commands {
		if _, err := fmt.Fprintf(writer, "  [%d/%d] %s\n", i+1, len(commands), cmd); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
		if err := s.client.Execute(ctx, cmd); err != nil {
			return fmt.Errorf("command failed: %s: %w", cmd, err)
		}
	}
	return nil
}

// GetInstallInfo retrieves installation information through the provider.
func (s *InstallService) GetInstallInfo() providers.InstallInfo {
	return s.provider.GetInstallInfo()
}
