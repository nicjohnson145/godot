package cmd

import (
	"os"

	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	targetCmd.AddCommand(listCmd)
	rootCmd.AddCommand(targetCmd)
}

var (
	targetCmd = &cobra.Command{
		Use: "target",
		Short: "Interact with the contents of targets",
		Long: "Add/remove/show what files are assigned to each target",
		RunE: List,
	}

	listCmd = &cobra.Command{
		Use: "list",
		Short: "list all files maintained by godot",
		RunE: List,
	}
)

func List(cmd *cobra.Command, args []string) error {
	conf := config.NewConfig(&util.OSHomeDir{})
	conf.ListAllFiles(os.Stdout)
	return nil
}
