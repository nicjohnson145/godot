package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	ceaseCmd.AddCommand(ceaseFileCmd)
	ceaseCmd.AddCommand(ceaseBootstrapCmd)
	rootCmd.AddCommand(ceaseCmd)
}

var (
	ceaseCmd = &cobra.Command{
		Use:   "cease",
		Short: "Remove an object to a target",
	}

	ceaseFileCmd = &cobra.Command{
		Use:   "file [file]",
		Short: "Remove an file from a target",
		Long: "Remove a file from a target. If a template name is not given, a selection prompt will " +
			"open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.TargetRemoveFile(target, args)
		},
	}

	ceaseBootstrapCmd = &cobra.Command{
		Use: "bootstrap [bootstrap]",
		Short: "Remove an item to be bootstrapped from a target",
		Long: "Remove an item to be bootstrapped from a target. If a bootstrap name is not given, a " +
			  "selection prompt wil open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.RemoveTargetBootstrap(target, args)
		},
	}
)
