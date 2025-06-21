package version

import (
	"sync"

	"github.com/kodflow/superviz.io/internal/services"
	"github.com/spf13/cobra"
)

var (
	defaultService *services.VersionService
	defaultCmd     *cobra.Command
	once           sync.Once
)

// initDefaults initializes the default service and command once
func initDefaults() {
	defaultService = services.NewVersionService(nil)
	defaultCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return defaultService.DisplayVersion(cmd.OutOrStdout())
		},
	}
}

// GetCommand returns the singleton Cobra command for displaying version information.
func GetCommand() *cobra.Command {
	once.Do(initDefaults)
	return defaultCmd
}

// GetCommandWithService returns a Cobra command with a custom service.
// If service is nil, returns the default singleton command.
func GetCommandWithService(service *services.VersionService) *cobra.Command {
	if service == nil {
		return GetCommand()
	}

	// Create new command for custom services (bypass singleton)
	return NewVersionCommand(service)
}

// NewVersionCommand creates a new version command with the given service.
// Exposed for testing or special cases where you need a fresh instance.
func NewVersionCommand(service *services.VersionService) *cobra.Command {
	// Don't use default service here - use the provided one directly
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.DisplayVersion(cmd.OutOrStdout())
		},
	}
}
