// Package version provides CLI command functionality for displaying version information
package version

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/kodflow/superviz.io/internal/services"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the different output formats supported by the version command
type OutputFormat string

const (
	// FormatDefault represents the default human-readable output format
	FormatDefault OutputFormat = "default"
	// FormatJSON represents JSON output format
	FormatJSON OutputFormat = "json"
	// FormatYAML represents YAML output format
	FormatYAML OutputFormat = "yaml"
	// FormatShort represents a short version-only output format
	FormatShort OutputFormat = "short"
)

// VersionFlags contains command-line flags for the version command
type VersionFlags struct {
	// Format specifies the output format (default, json, yaml, short)
	Format string
}

var (
	// defaultService holds the singleton version service instance
	defaultService *services.VersionService
	// defaultCmd holds the singleton version command instance
	defaultCmd *cobra.Command
	// flags holds the command flags for the singleton command
	flags *VersionFlags
	// once ensures the default instances are initialized only once
	once sync.Once
)

// initDefaults initializes the default service and command instances once.
//
// initDefaults creates the singleton instances of the version service and
// command with enhanced flags support, ensuring they are created only once
// for the lifetime of the application.
//
// Example:
//
//	initDefaults() // Called automatically via sync.Once
//
// Parameters:
//   - None
//
// Returns:
//   - None (initializes global variables)
func initDefaults() {
	defaultService = services.NewVersionService(nil)
	flags = &VersionFlags{}

	defaultCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information with flexible output formats",
		Long: `Print version information including build details, commit hash, and runtime information.

Supports multiple output formats:
- default: Human-readable format (default)
- json: JSON structured output
- yaml: YAML structured output  
- short: Version string only

Examples:
  svz version                    # Default format
  svz version --format=json      # JSON format
  svz version --format=short     # Version only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersionCommand(cmd, defaultService, flags)
		},
	}

	// Add command flags
	defaultCmd.Flags().StringVarP(&flags.Format, "format", "f", "default",
		"Output format (default|json|yaml|short)")
}

// GetCommand returns the singleton Cobra command for displaying version information.
//
// GetCommand provides access to the default version command instance, initializing
// it if necessary using sync.Once for thread safety.
//
// Example:
//
//	versionCmd := GetCommand()
//	rootCmd.AddCommand(versionCmd)
//
// Parameters:
//   - None
//
// Returns:
//   - cmd: *cobra.Command instance configured for version display
func GetCommand() *cobra.Command {
	once.Do(initDefaults)
	return defaultCmd
}

// GetCommandWithService returns a Cobra command with a custom version service.
//
// GetCommandWithService allows injection of a custom version service while
// falling back to the singleton command if service is nil.
//
// Example:
//
//	customProvider := &MyVersionProvider{}
//	service := services.NewVersionService(customProvider)
//	cmd := GetCommandWithService(service)
//
// Parameters:
//   - service: *services.VersionService custom version service instance (nil for default)
//
// Returns:
//   - cmd: *cobra.Command instance with the specified or default service
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
// provided service and full flag support, bypassing the singleton pattern
// for testing or special cases.
//
// Example:
//
//	mockProvider := &MockVersionProvider{}
//	service := services.NewVersionService(mockProvider)
//	cmd := NewVersionCommand(service)
//
// Parameters:
//   - service: *services.VersionService version service instance to use for the command
//
// Returns:
//   - cmd: *cobra.Command new Cobra command instance configured with the provided service
func NewVersionCommand(service *services.VersionService) *cobra.Command {
	cmdFlags := &VersionFlags{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information with flexible output formats",
		Long: `Print version information including build details, commit hash, and runtime information.

Supports multiple output formats:
- default: Human-readable format (default)
- json: JSON structured output
- yaml: YAML structured output  
- short: Version string only

Examples:
  svz version                    # Default format
  svz version --format=json      # JSON format
  svz version --format=short     # Version only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersionCommand(cmd, service, cmdFlags)
		},
	}

	// Add command flags
	cmd.Flags().StringVarP(&cmdFlags.Format, "format", "f", "default",
		"Output format (default|json|yaml|short)")

	return cmd
}

// runVersionCommand executes the version command logic with zero-allocation optimizations
//
// runVersionCommand handles the core version display logic with support for
// multiple output formats.
//
// Code block:
//
//	flags := &VersionFlags{Format: "json"}
//	err := runVersionCommand(cmd, service, flags)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - 1 cmd: *cobra.Command - cobra command instance for output
//   - 2 service: *services.VersionService - version service for data retrieval
//   - 3 flags: *VersionFlags - command flags configuration
//
// Returns:
//   - 1 error - non-nil if validation fails or output generation errors
func runVersionCommand(cmd *cobra.Command, service *services.VersionService, flags *VersionFlags) error {
	// Validate output format
	format := OutputFormat(strings.ToLower(flags.Format))
	switch format {
	case FormatDefault, FormatJSON, FormatYAML, FormatShort:
		// Valid formats
	default:
		return fmt.Errorf("invalid output format: %s (valid: default, json, yaml, short)", flags.Format)
	}

	// Get version information
	info := service.GetVersionInfo()

	// Generate output based on format
	switch format {
	case FormatJSON:
		return outputJSON(cmd, info)
	case FormatYAML:
		return outputYAML(cmd, info)
	case FormatShort:
		return outputShort(cmd, info)
	default: // FormatDefault
		return outputDefault(cmd, service)
	}
}

// outputDefault writes version information in human-readable format
func outputDefault(cmd *cobra.Command, service *services.VersionService) error {
	return service.DisplayVersion(cmd.OutOrStdout())
}

// outputJSON writes version information in JSON format
func outputJSON(cmd *cobra.Command, info providers.VersionInfo) error {
	output := struct {
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		BuiltAt   string `json:"built_at"`
		BuiltBy   string `json:"built_by"`
		GoVersion string `json:"go_version"`
		OSArch    string `json:"os_arch"`
	}{
		Version:   info.Version,
		Commit:    info.Commit,
		BuiltAt:   info.BuiltAt,
		BuiltBy:   info.BuiltBy,
		GoVersion: info.GoVersion,
		OSArch:    info.OSArch,
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// outputYAML writes version information in YAML format
func outputYAML(cmd *cobra.Command, info providers.VersionInfo) error {
	output := struct {
		Version   string `yaml:"version"`
		Commit    string `yaml:"commit"`
		BuiltAt   string `yaml:"built_at"`
		BuiltBy   string `yaml:"built_by"`
		GoVersion string `yaml:"go_version"`
		OSArch    string `yaml:"os_arch"`
	}{
		Version:   info.Version,
		Commit:    info.Commit,
		BuiltAt:   info.BuiltAt,
		BuiltBy:   info.BuiltBy,
		GoVersion: info.GoVersion,
		OSArch:    info.OSArch,
	}

	encoder := yaml.NewEncoder(cmd.OutOrStdout())
	defer func() {
		if err := encoder.Close(); err != nil {
			// Log the error but don't fail the command
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to close encoder: %v\n", err)
		}
	}()
	return encoder.Encode(output)
}

// outputShort writes only the version string
func outputShort(cmd *cobra.Command, info providers.VersionInfo) error {
	_, err := cmd.OutOrStdout().Write([]byte(info.Version + "\n"))
	return err
}
