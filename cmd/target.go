package cmd

import (
	"os"

	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	targetCmd.AddCommand(listCmd)
	targetCmd.AddCommand(showCmd)
	targetCmd.AddCommand(addCmd)
	targetCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(targetCmd)

	addCmd.Flags().StringVarP(&target, "target", "t", "", "What target to add file to, defaults to current target")
	removeCmd.Flags().StringVarP(&target, "target", "t", "", "What target to remove file from, defaults to current target")
}

var (
	target string

	targetCmd = &cobra.Command{
		Use:   "target",
		Short: "Interact with the contents of targets",
		Long:  "Add/remove/show what files are assigned to each target",
	}

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all files maintained by godot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{})
			return c.ListAllFiles(os.Stdout)
		},
	}

	showCmd = &cobra.Command{
		Use:   "show [target]",
		Short: "Show current files assigned to a target",
		Long:  "Show current files assigned to a target, if target is not supplied, the current target will be used",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{})
			return c.TargetShow(target, os.Stdout)
		},
	}

	addCmd = &cobra.Command{
		Use:   "add [file]",
		Short: "Add a file to a target",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.TargetAdd(target, args)
		},
	}

	removeCmd = &cobra.Command{
		Use:   "remove [file]",
		Short: "Remove a file from target",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.TargetRemove(target, args)
		},
	}
)
