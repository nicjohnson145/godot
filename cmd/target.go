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
	targetCmd.AddCommand(addCmd)
	rootCmd.AddCommand(targetCmd)

	addCmd.Flags().StringVarP(&target, "target", "t", "", "What target to add file to, defaults to current target")
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
			conf := config.NewConfig(&util.OSHomeDir{})
			conf.ListAllFiles(os.Stdout)
			return nil
		},
	}

	showCmd = &cobra.Command{
		Use:   "show <target?>",
		Short: "Show current files assigned to a target",
		Long:  "Show current files assigned to a target, if target is not supplied, the current target will be used",
		Args:  cobra.MaximumNArgs(1),
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

	addCmd = &cobra.Command{
		Use:   "add <file>",
		Short: "Add a file to a target",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := config.NewConfig(&util.OSHomeDir{})
			if target == "" {
				target = conf.Target
			}
			err := conf.AddToTarget(target, args[0])
			if err == nil {
				err = conf.Write()
			}
			return err
		},
	}
)
