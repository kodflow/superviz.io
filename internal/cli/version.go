package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Ces variables seront remplacées via -ldflags au build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Superviz version: %s\n", version)
		fmt.Printf("Git commit:        %s\n", commit)
		fmt.Printf("Built at:          %s\n", date)
		fmt.Printf("Built by:          %s\n", builtBy)
	},
}
