package main

import (
	"os"

	"github.com/kodflow/superviz.io/internal/cli"
)

// main executes the CLI command and exits with status code 1 if an error occurs.
func main() {
	if err := cli.GetCLICommand().Execute(); err != nil {
		os.Exit(1)
	}
}
