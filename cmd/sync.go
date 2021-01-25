package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "If a target file already exists, delete it then symlink it")
}

var (
	force bool

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Build and symlink dotfiles",
		Long:  "Compile templates and symlink them to their final destinations",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.Sync(controller.SyncOpts{Force: force})
		},
	}
)
