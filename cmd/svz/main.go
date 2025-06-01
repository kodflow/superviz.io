package main

import (
	"os"

	"github.com/kodflow/superviz.io/internal/cli"
)

func main() {
	if err := cli.GetCLICommand().Execute(); err != nil {
		os.Exit(1)
	}
}
