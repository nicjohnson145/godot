package cmd

import (
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "If a target file already exists, delete it then symlink it")
	syncCmd.Flags().BoolVarP(&noBootstrap, "no-bootstrap", "-b", false, "Don't attempt to bootstrap this host, only symlink")
}

var (
	force bool
	noBootstrap bool

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Build and symlink dotfiles",
		Long:  "Compile templates, symlink them to their final destinations, install configured bootstraps",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.Sync(controller.SyncOpts{Force: force, NoBootstrap: noBootstrap})
		},
	}
)
