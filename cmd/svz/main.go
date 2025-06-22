package main

import (
	"fmt"
	"os"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/kodflow/superviz.io/internal/cli/commands/install"
	"github.com/kodflow/superviz.io/internal/cli/commands/version"
	"github.com/spf13/cobra"
)

// run executes the given Cobra command with args and returns the exit code.
func run(cmd *cobra.Command, args []string) int {
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

// main is the CLI entry point.
func main() {
	rootCmd := cli.NewRootCommand(
		version.GetCommand(),
		install.GetCommand(),
	)

	os.Exit(run(rootCmd, os.Args[1:]))
}
