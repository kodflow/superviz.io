// Package version provides CLI command functionality for displaying version information
package version

import (
	"sync"

	"github.com/kodflow/superviz.io/internal/services"
	"github.com/spf13/cobra"
)

var (
	// defaultService holds the singleton version service instance
	defaultService *services.VersionService
	// defaultCmd holds the singleton version command instance
	defaultCmd *cobra.Command
	// once ensures the default instances are initialized only once
	once sync.Once
)

// initDefaults initializes the default service and command instances once.
//
// initDefaults creates the singleton instances of the version service and
// command, ensuring they are created only once for the lifetime of the application.
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
//
// GetCommand provides access to the default version command instance, initializing
// it if necessary using sync.Once for thread safety.
//
// Returns:
//   - Cobra command instance configured for version display
func GetCommand() *cobra.Command {
	once.Do(initDefaults)
	return defaultCmd
}

// GetCommandWithService returns a Cobra command with a custom version service.
//
// GetCommandWithService allows injection of a custom version service while
// falling back to the singleton command if service is nil.
//
// Parameters:
//   - service: Custom version service instance (nil for default)
//
// Returns:
//   - Cobra command instance with the specified or default service
func GetCommandWithService(service *services.VersionService) *cobra.Command {
	if service == nil {
		return GetCommand()
	}

	// Create new command for custom services (bypass singleton)
	return NewVersionCommand(service)
}

// NewVersionCommand creates a new version command with the given service.
//
// NewVersionCommand constructs a fresh version command instance with the
// provided service, bypassing the singleton pattern for testing or special cases.
//
// Parameters:
//   - service: Version service instance to use for the command
//
// Returns:
//   - New Cobra command instance configured with the provided service
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
