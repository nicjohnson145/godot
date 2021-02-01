package cmd

import (
	"os"

	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listFilesCmd)
	listCmd.AddCommand(listBootstrapsCmd)

	rootCmd.AddCommand(listCmd)
}

var (
	listCmd = &cobra.Command{
		Use:   "list [object]",
		Short: "List information",
		Args:  cobra.ExactArgs(1),
	}

	listFilesCmd = &cobra.Command{
		Use:   "files",
		Short: "List files managed by godot",
		Long:  "List files managed by godot, if -t/--target is given, only that targets files will be listed",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.ShowFilesEntry(target, os.Stdout)
		},
	}

	listBootstrapsCmd = &cobra.Command{
		Use:   "bootstraps",
		Short: "List bootstrap items managed by godot",
		Long:  "List bootstrap items managed by godot, if -t/--target is given, only that targets bootstrap items will be listed",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.ShowBootstrapsEntry(target, os.Stdout)
		},
	}
)