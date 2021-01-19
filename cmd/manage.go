package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(manageCmd)
	manageCmd.Flags().StringVar(&as, "as", "", "Override the generated template name, useful for resolving collisions")
	manageCmd.Flags().BoolVar(&add, "add", false, "Add the imported file to the current target")
}

var (
	as  string
	add bool

	manageCmd = &cobra.Command{
		Use:   "manage <file>",
		Short: "Add file to be managed by godot",
		Long:  "Import/create a file to be managed by godot at the specified location",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{})
			return c.Import(args[0], as, controller.ImportOpts{NoGit: noGit, NoAdd: !add})
		},
	}
)
