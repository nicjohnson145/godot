package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/nicjohnson145/godot/internal/lib"
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
		Use: "sync",
		Short: "Sync configuration",
		Long: "Sync local filesystem with configuration",
		Run: func(cmd *cobra.Command, args []string) {
			lib.Sync()
		},
	}
	rootCmd.AddCommand(syncCmd)

	return rootCmd
}
