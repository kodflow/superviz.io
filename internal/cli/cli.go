package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root Cobra command for the Superviz CLI.
//
// Parameters:
//   - subcommands: list of subcommands to attach
//
// Returns:
//   - *cobra.Command: the fully configured root CLI command.
func NewRootCommand(subcommands ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "svz",
		Short:                 "Superviz - Declarative Process Supervisor",
		DisableAutoGenTag:     true,
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
	}

	// Hide default help command
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Add provided subcommands
	cmd.AddCommand(subcommands...)

	return cmd
}
