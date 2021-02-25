package cmd

import (
	"strings"

	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/controller"
	"github.com/spf13/cobra"
)

func init() {
	addFile.Flags().StringVar(&as, "as", "", "Override the default template name (useful for resolving name conflicts)")

	addBootstrapCmd.Flags().StringVarP(&name, "name", "n", "", "The name of the package for this manager, for git this is the repo URL")
	addBootstrapCmd.MarkFlagRequired("name")
	addBootstrapCmd.Flags().StringVarP(&manager, "manager", "m", "", "The name of the package manager, one of "+strings.Join(config.ValidManagers, ", "))
	addBootstrapCmd.MarkFlagRequired("manager")
	addBootstrapCmd.Flags().StringVarP(&location, "location", "l", "", "When using the git manager, where the repo should be checked out to, defaults to home directory")

	addCmd.AddCommand(addFile)
	addCmd.AddCommand(addBootstrapCmd)

	rootCmd.AddCommand(addCmd)
}

var (
	add bool

	as string

	name     string
	location string
	manager  string

	addCmd = &cobra.Command{
		Use:   "add [object]",
		Short: "Add an object",
		Args:  cobra.ExactArgs(1),
	}

	addFile = &cobra.Command{
		Use:   "file file",
		Short: "Add a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.Import(args[0], as)
		},
	}

	addBootstrapCmd = &cobra.Command{
		Use:   "bootstrap bs",
		Short: "Add a bootstrap item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := controller.NewController(controller.ControllerOpts{NoGit: noGit})
			return c.AddBootstrapItem(args[0], manager, name, location)
		},
	}
)
