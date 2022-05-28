package main

import (
	"github.com/nicjohnson145/godot/internal/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func buildCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "godot",
		Short: "A dotfiles manager",
		Long:  "A staticly linked dotfiles manager written in Go",
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync configuration",
		Long:  "Sync local filesystem with configuration",
		Run: func(cmd *cobra.Command, args []string) {
			lib.Sync()
		},
	}
	rootCmd.AddCommand(syncCmd)

	validateCmd := &cobra.Command{
		Use:   "validate <path-to-config>",
		Args: cobra.ExactArgs(1),
		Short: "Validate configuration",
		Long:  "Validate a configuration file on disk",
		Run: func(cmd *cobra.Command, args []string) {
			lib.Validate(args[0])
		},
	}
	rootCmd.AddCommand(validateCmd)

	return rootCmd
}
