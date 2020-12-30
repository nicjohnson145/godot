package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(manageCmd)
	manageCmd.Flags().StringVar(&as, "as", "", "Override the generated template name, useful for resolving collisions")
}

var (
	as string

	manageCmd = &cobra.Command{
		Use: "manage <file>",
		Short: "Add file to be managed by godot",
		Long: "Import/create a file to be managed by godot at the specified location",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
)
