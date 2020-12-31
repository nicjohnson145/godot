package cmd

import (
	"os"

	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	targetCmd.AddCommand(listCmd)
	targetCmd.AddCommand(showCmd)
	rootCmd.AddCommand(targetCmd)
}

var (
	targetCmd = &cobra.Command{
		Use: "target",
		Short: "Interact with the contents of targets",
		Long: "Add/remove/show what files are assigned to each target",
	}

	listCmd = &cobra.Command{
		Use: "list",
		Short: "List all files maintained by godot",
		RunE: List,
	}

	showCmd = &cobra.Command{
		Use: "show <target?>",
		Short: "Show current files assigned to a target",
		Long: "Show current files assigned to a target, if target is not supplied, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := config.NewConfig(&util.OSHomeDir{})
			if len(args) == 0 {
				conf.ListTargetFiles(conf.Target, os.Stdout)
			} else {
				conf.ListTargetFiles(args[0], os.Stdout)
			}
			return nil
		},
	}
)

func List(cmd *cobra.Command, args []string) error {
	conf := config.NewConfig(&util.OSHomeDir{})
	conf.ListAllFiles(os.Stdout)
	return nil
}
