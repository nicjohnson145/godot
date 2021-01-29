package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	useCmd.AddCommand(useFileCmd)
	rootCmd.AddCommand(useCmd)
}

var (
	useCmd = &cobra.Command{
		Use: "use",
		Short: "Add an object to a target",
	}

	useFileCmd = &cobra.Command{
		Use: "file [file]",
		Short: "Add an file to a target",
		Long: "Add a file to a target. If a template name is not given, a selection prompt will " +
			  "open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.TargetAdd(target, args)
		},
	}
)
