// Package main provides the entry point for the superviz.io CLI application
package main

import (
	"fmt"
	"os"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/kodflow/superviz.io/internal/cli/commands/install"
	"github.com/kodflow/superviz.io/internal/cli/commands/version"
	"github.com/spf13/cobra"
)

// run executes the given Cobra command with arguments and returns the exit code.
//
// run provides a controlled execution environment for the CLI command,
// handling errors gracefully and returning appropriate exit codes.
//
// Parameters:
//   - cmd: Cobra command to execute
//   - args: Command-line arguments to pass to the command
//
// Returns:
//   - Exit code (0 for success, 1 for error)
func run(cmd *cobra.Command, args []string) int {
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

// main is the CLI entry point for the superviz.io application.
//
// main initializes the root command with all available subcommands
// and executes the CLI with the provided command-line arguments.
func main() {
	rootCmd := cli.NewRootCommand(
		version.GetCommand(),
		install.GetCommand(),
	)

	os.Exit(run(rootCmd, os.Args[1:]))
}
