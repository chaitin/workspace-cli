package chaitin

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "chaitin",
		Short: "Demo product command",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Uncomputable, infinite possibilities")
			return nil
		},
	}
}
