package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func RequireOneTarget(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must specify the target as user@host")
	}
	if !strings.Contains(args[0], "@") {
		return fmt.Errorf("target must be in format user@host")
	}
	return nil
}
