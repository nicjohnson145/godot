package cmd

import (
	// "github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var (
	listCmd = &cobra.Command{
		Use: "list [object]",
		Short: "List information",
		Args: cobra.ExactArgs(1),
	}

	listFilesCmd = &cobra.Command{
		Use: "files",
	}
)
