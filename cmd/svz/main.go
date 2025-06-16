package main

import (
	"github.com/kodflow/superviz.io/internal/cli"
)

// main executes the CLI command for the application.
func main() {
	cli.GetCLICommand().Execute() //nolint:errcheck
}
