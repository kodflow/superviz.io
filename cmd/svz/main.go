package main

import (
	"github.com/kodflow/superviz.io/internal/cli"
)

// main executes the CLI command and exits with status code 1 if an error occurs.
func main() {
	cli.GetCLICommand().Execute() //nolint:errcheck
}
