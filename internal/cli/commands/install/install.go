// Package install provides CLI command functionality for superviz.io installation
package install

import (
	"sync"
	"time"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services"
	"github.com/kodflow/superviz.io/internal/utils"
	"github.com/spf13/cobra"
)

var (
	// defaultService holds the singleton install service instance
	defaultService *services.InstallService
	// defaultCmd holds the singleton install command instance
	defaultCmd *cobra.Command
	// once ensures the default instances are initialized only once
	once sync.Once
)

// initDefaults initializes the default service and command instances once.
//
// initDefaults creates the singleton instances of the install service and
// command, ensuring they are created only once for the lifetime of the application.
func initDefaults() {
	defaultService = services.NewInstallService(nil)
	defaultCmd = createInstallCommand(defaultService)
}

// GetCommand returns the singleton Cobra command for installation operations.
//
// GetCommand provides access to the default install command instance, initializing
// it if necessary using sync.Once for thread safety.
//
// Returns:
//   - Cobra command instance configured for superviz.io installation
func GetCommand() *cobra.Command {
	once.Do(initDefaults)
	return defaultCmd
}

// GetCommandWithService returns a Cobra command with a custom install service.
//
// GetCommandWithService allows injection of a custom install service while
// falling back to the singleton command if service is nil.
//
// Parameters:
//   - service: Custom install service instance (nil for default)
//
// Returns:
//   - Cobra command instance with the specified or default service
func GetCommandWithService(service *services.InstallService) *cobra.Command {
	if service == nil {
		return GetCommand()
	}
	return NewInstallCommand(service)
}

// NewInstallCommand creates a new install command with the given service.
//
// NewInstallCommand constructs a fresh install command instance with the
// provided service, bypassing the singleton pattern for testing or special cases.
//
// Parameters:
//   - service: Install service instance to use for the command
//
// Returns:
//   - New Cobra command instance configured with the provided service
func NewInstallCommand(service *services.InstallService) *cobra.Command {
	return createInstallCommand(service)
}

// createInstallCommand creates the cobra command with all flags and validation.
//
// createInstallCommand constructs a fully configured Cobra command for installation
// operations, including argument validation, flags, and command execution logic.
//
// Parameters:
//   - service: Install service instance to handle the installation logic
//
// Returns:
//   - Configured Cobra command ready for execution
func createInstallCommand(service *services.InstallService) *cobra.Command {
	opts := &providers.InstallConfig{
		Port:    22,
		Timeout: 300 * time.Second,
	}

	cmd := &cobra.Command{
		Use:   "install user@host [flags]",
		Short: "Setup superviz.io repository on remote system",
		Long:  "Setup superviz.io package repository on the remote system so you can install superviz.io using the system package manager (apt, apk, yum, etc.).",
		Args:  utils.RequireOneTarget,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return service.ValidateAndPrepareConfig(opts, args)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return service.Install(cmd.Context(), cmd.OutOrStdout(), opts)
		},
	}

	// Configure command flags for SSH connection and installation options
	cmd.Flags().StringVarP(&opts.KeyPath, "ssh-key", "i", "", "Path to SSH private key file")
	cmd.Flags().IntVarP(&opts.Port, "ssh-port", "p", 22, "SSH port")
	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", 300*time.Second, "Connection timeout (e.g. 30s, 5m)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force installation even if components already exist")
	cmd.Flags().BoolVar(&opts.SkipHostKeyCheck, "skip-host-key-check", false, "Skip host key verification (development only)")

	return cmd
}
