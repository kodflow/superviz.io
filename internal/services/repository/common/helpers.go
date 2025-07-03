// internal/services/repository/common/helpers.go
package common

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kodflow/superviz.io/internal/infrastructure/transports/ssh"
)

// SudoHelper helps with sudo detection and command prefixing.
//
// SudoHelper provides intelligent sudo privilege detection and automatic
// command prefixing for repository setup operations across different distributions.
type SudoHelper struct {
	// client provides SSH connectivity for privilege testing
	client ssh.Client
}

// NewSudoHelper creates a new sudo helper.
//
// NewSudoHelper initializes a sudo helper with the provided SSH client
// for privilege detection and command modification operations.
//
// Example:
//
//	client := ssh.NewClient(config)
//	sudo := NewSudoHelper(client)
//	needed, err := sudo.IsNeeded(ctx)
//
// Parameters:
//   - client: ssh.Client for executing privilege detection commands
//
// Returns:
//   - helper: *SudoHelper configured sudo helper instance
func NewSudoHelper(client ssh.Client) *SudoHelper {
	return &SudoHelper{
		client: client,
	}
}

// IsNeeded checks if sudo is needed for system operations.
//
// IsNeeded performs intelligent privilege detection by testing write access
// to system directories across different distributions (APT, APK, YUM, Pacman).
// It returns false if the current user has sufficient privileges, or true if
// sudo is required and available.
//
// Example:
//
//	needed, err := sudo.IsNeeded(ctx)
//	if err != nil {
//		return fmt.Errorf("privilege detection failed: %w", err)
//	}
//	if needed {
//		fmt.Println("Sudo required for repository setup")
//	}
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//
// Returns:
//   - needed: bool true if sudo is required, false if current privileges are sufficient
//   - err: error if privilege detection fails or sudo is unavailable when needed
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
//
// AddPrefix intelligently adds the "sudo" prefix to commands that require
// root privileges while leaving commands that don't need elevated permissions unchanged.
// This ensures minimal privilege escalation while maintaining functionality.
//
// Example:
//
//	commands := []string{"apt update", "echo test"}
//	sudoCommands := sudo.AddPrefix(commands, true)
//	// Result: ["sudo apt update", "echo test"]
//
// Parameters:
//   - commands: []string original command list to modify
//   - needSudo: bool whether to apply sudo prefixes
//
// Returns:
//   - modified: []string commands with sudo prefixes applied where needed
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
//
// commandNeedsSudo analyzes commands to determine if they require root privileges
// by checking against known package manager commands and system path operations.
// This helps minimize unnecessary privilege escalation.
//
// Example:
//
//	needsRoot := sudo.commandNeedsSudo("apt update")  // returns true
//	needsRoot = sudo.commandNeedsSudo("echo test")    // returns false
//
// Parameters:
//   - cmd: string command to analyze for privilege requirements
//
// Returns:
//   - needed: bool true if command requires root privileges, false otherwise
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
		"/etc/", "/usr/", "/usr/share/keyrings/", "/etc/yum.repos.d/",
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
//
// CommandExecutor provides sequential command execution with progress reporting
// and comprehensive error handling for repository setup operations.
type CommandExecutor struct {
	// client provides SSH connectivity for command execution
	client ssh.Client
}

// NewCommandExecutor creates a new command executor.
//
// NewCommandExecutor initializes a command executor with the provided SSH client
// for sequential command execution with progress reporting.
//
// Example:
//
//	client := ssh.NewClient(config)
//	executor := NewCommandExecutor(client)
//	err := executor.Execute(ctx, commands, writer)
//
// Parameters:
//   - client: ssh.Client for executing commands on remote systems
//
// Returns:
//   - executor: *CommandExecutor configured command executor instance
func NewCommandExecutor(client ssh.Client) *CommandExecutor {
	return &CommandExecutor{
		client: client,
	}
}

// Execute executes a list of commands in sequence.
//
// Execute runs commands sequentially with progress reporting and stops on the first
// error. Each command execution is numbered and reported to the writer for
// user feedback during repository setup operations.
//
// Example:
//
//	commands := []string{"apt update", "apt install -y curl"}
//	err := executor.Execute(ctx, commands, os.Stdout)
//	// Output: [1/2] apt update
//	//         [2/2] apt install -y curl
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - commands: []string list of commands to execute sequentially
//   - writer: io.Writer for progress output and command reporting
//
// Returns:
//   - err: error if any command fails or output writing fails
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
