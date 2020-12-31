package cmd

import (
	"github.com/spf13/cobra"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/nicjohnson145/godot/internal/config"
)

func init() {
	rootCmd.AddCommand(manageCmd)
	manageCmd.Flags().StringVar(&as, "as", "", "Override the generated template name, useful for resolving collisions")
}

var (
	as string

	manageCmd = &cobra.Command{
		Use: "manage <file>",
		Short: "Add file to be managed by godot",
		Long: "Import/create a file to be managed by godot at the specified location",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := config.NewConfig(&util.OSHomeDir{})
			if as != "" {
				return conf.AddFile(as, args[0])
			} else {
				return conf.ManageFile(args[0])
			}
		},
	}
)
