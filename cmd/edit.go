package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().BoolVarP(&noSync, "no-sync", "s", false, "Do not issue a sync after editing")
}

var (
	noSync bool

	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit a file managed by godot",
		Long: "Edit a file managed by godot, if no file is suppled a prompt will be displayed",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{})
			return c.Edit(args, controller.EditOpts{NoSync: noSync, NoGit: noGit})
		},
	}
)

