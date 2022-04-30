package main

import (
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

	return rootCmd
}
