package main

import (
	"os"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/spf13/cobra"
)

// run executes the provided Cobra command with the given arguments.
//
// This function sets the arguments on the command and calls Execute().
// It returns a process exit code based on success or failure.
//
// Parameters:
//
//   - cmd: The Cobra command to execute.
//   - args: The list of CLI arguments to pass to the command.
//
// Returns:
//
//   - int: Exit code (0 on success, 1 on failure).
func run(cmd *cobra.Command, args []string) int {
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}

// main is the entry point of the Superviz CLI application.
//
// It delegates execution to the run() function using the root CLI command
// and the arguments passed via os.Args. The process exits with the code
// returned by run().
func main() {
	os.Exit(run(cli.GetCLICommand(), os.Args[1:]))
}
