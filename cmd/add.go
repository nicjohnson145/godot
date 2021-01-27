package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	addFile.Flags().StringVar(&as, "as", "", "Override the default template name (useful for resolving name conflicts)")

	addCmd.AddCommand(addFile)
	rootCmd.AddCommand(addCmd)
}

var (
	as string

	addCmd = &cobra.Command{
		Use: "add [object]",
		Short: "Add an object",
		Args: cobra.ExactArgs(1),
	}

	addFile = &cobra.Command{
		Use: "file [file]",
		Short: "Add a file",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.Import(args[0], as, controller.ImportOpts{})
		},
	}
)
