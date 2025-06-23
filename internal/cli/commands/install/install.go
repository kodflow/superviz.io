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
	defaultService *services.InstallService
	defaultCmd     *cobra.Command
	once           sync.Once
)

// initDefaults initializes the default service and command once
func initDefaults() {
	defaultService = services.NewInstallService(nil)
	defaultCmd = createInstallCommand(defaultService)
}

// GetCommand returns the singleton Cobra command for installation.
func GetCommand() *cobra.Command {
	once.Do(initDefaults)
	return defaultCmd
}

// GetCommandWithService returns a Cobra command with a custom service.
// If service is nil, returns the default singleton command.
func GetCommandWithService(service *services.InstallService) *cobra.Command {
	if service == nil {
		return GetCommand()
	}
	return NewInstallCommand(service)
}

// NewInstallCommand creates a new install command with the given service.
func NewInstallCommand(service *services.InstallService) *cobra.Command {
	return createInstallCommand(service)
}

// createInstallCommand creates the cobra command with all flags and validation
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

	// Flags
	cmd.Flags().StringVarP(&opts.KeyPath, "ssh-key", "i", "", "Path to SSH private key file")
	cmd.Flags().IntVarP(&opts.Port, "ssh-port", "p", 22, "SSH port")
	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", 300*time.Second, "Connection timeout (e.g. 30s, 5m)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force installation even if components already exist")
	cmd.Flags().BoolVar(&opts.SkipHostKeyCheck, "skip-host-key-check", false, "Skip host key verification (development only)")

	return cmd
}
