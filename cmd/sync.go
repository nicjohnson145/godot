package cmd

import (
	"github.com/nicjohnson145/godot/internal/builder"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var (
	force bool

	syncCmd = &cobra.Command{
		Use: "sync",
		Short: "Build and symlink dotfiles",
		Long: "Compile templates and symlink them to their final destinations",
		RunE: func(cmd *cobra.Command, args []string) error {
			b := builder.Builder{Getter: &util.OSHomeDir{}}
			err := b.Build()
			return err
		},
	}
)

func init() {
	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "If a target file already exists, delete it then symlink it")
}

