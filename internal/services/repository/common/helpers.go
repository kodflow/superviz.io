// internal/services/repository/common/helpers.go
package common

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kodflow/superviz.io/internal/transports/ssh"
)

// SudoHelper helps with sudo detection and command prefixing.
type SudoHelper struct {
	client ssh.Client
}

// NewSudoHelper creates a new sudo helper.
func NewSudoHelper(client ssh.Client) *SudoHelper {
	return &SudoHelper{
		client: client,
	}
}

// IsNeeded checks if sudo is needed for system operations.
func (s *SudoHelper) IsNeeded(ctx context.Context) (bool, error) {
	// Check if we can write to system directories without sudo
	testCommands := []string{
		"test -w /etc/apt/sources.list.d/", // Debian/Ubuntu
		"test -w /etc/apk/repositories",    // Alpine
		"test -w /etc/yum.repos.d/",        // RHEL/CentOS/Fedora
		"test -w /etc/pacman.conf",         // Arch
	}

	// If ANY system directory is writable, we don't need sudo
	for _, cmd := range testCommands {
		if err := s.client.Execute(ctx, cmd); err == nil {
			return false, nil // No sudo needed
		}
	}

	// Check if sudo is available
	if err := s.client.Execute(ctx, "command -v sudo >/dev/null 2>&1"); err != nil {
		return false, fmt.Errorf("root privileges required but sudo is not available")
	}

	return true, nil // Sudo needed and available
}

// AddPrefix adds sudo prefix to commands if needed.
func (s *SudoHelper) AddPrefix(commands []string, needSudo bool) []string {
	if !needSudo {
		return commands
	}

	sudoCommands := make([]string, len(commands))
	for i, cmd := range commands {
		// Don't add sudo to commands that don't need it
		if s.commandNeedsSudo(cmd) {
			sudoCommands[i] = "sudo " + cmd
		} else {
			sudoCommands[i] = cmd
		}
	}
	return sudoCommands
}

// commandNeedsSudo determines if a command requires root privileges.
func (s *SudoHelper) commandNeedsSudo(cmd string) bool {
	// Commands that typically need root access
	rootCommands := []string{
		// APT (Debian/Ubuntu)
		"apt update",
		"apt install",
		"apt-get update",
		"apt-get install",

		// APK (Alpine)
		"apk update",
		"apk add",
		"apk upgrade",

		// YUM / DNF (RHEL, CentOS, Fedora)
		"yum",
		"yum install",
		"yum update",
		"dnf",
		"dnf install",
		"dnf update",

		// RPM
		"rpm --import",
		"rpm -i",

		// Pacman (Arch)
		"pacman",
		"pacman -S",
		"pacman -Sy",
		"pacman-key",

		// Zypper (openSUSE)
		"zypper install",
		"zypper update",

		// Portage (Gentoo)
		"emerge",
	}

	// Commands that write to system directories
	systemPaths := []string{
		"/etc/", "/usr/share/keyrings/", "/etc/yum.repos.d/",
		"/etc/apk/", "/etc/pacman.conf",
	}

	// Check if command starts with a root-requiring command
	for _, rootCmd := range rootCommands {
		if strings.HasPrefix(cmd, rootCmd) {
			return true
		}
	}

	// Check if command writes to system directories
	for _, path := range systemPaths {
		if strings.Contains(cmd, path) {
			return true
		}
	}

	return false
}

// CommandExecutor executes commands with proper error handling.
type CommandExecutor struct {
	client ssh.Client
}

// NewCommandExecutor creates a new command executor.
func NewCommandExecutor(client ssh.Client) *CommandExecutor {
	return &CommandExecutor{
		client: client,
	}
}

// Execute executes a list of commands in sequence.
func (c *CommandExecutor) Execute(ctx context.Context, commands []string, writer io.Writer) error {
	for i, cmd := range commands {
		if _, err := fmt.Fprintf(writer, "  [%d/%d] %s\n", i+1, len(commands), cmd); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
		if err := c.client.Execute(ctx, cmd); err != nil {
			return fmt.Errorf("command failed: %s: %w", cmd, err)
		}
	}
	return nil
}
