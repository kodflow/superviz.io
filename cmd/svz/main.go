package main

import (
	"os"

	"github.com/kodflow/superviz.io/internal/cli"
	"github.com/spf13/cobra"
)

// run exécute la commande passée et renvoie le code de sortie.
func run(cmd *cobra.Command, args []string) int {
	cmd.SetArgs(args) // facile à surcharger en test
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(cli.GetCLICommand(), os.Args[1:]))
}
