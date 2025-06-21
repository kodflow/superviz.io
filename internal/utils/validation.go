package utils

import (
	"fmt"

	"github.com/spf13/cobra"
)

func RequireOneTarget(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("you must specify the target as user@host")
	}
	return nil
}
