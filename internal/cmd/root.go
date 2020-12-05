package cmd

import (
    "github.com/spf13/cobra"
)

var (
    rootCmd = &cobra.Command{
        Use: "godot",
        Short: "A dotfiles manager",
        Long: "Godot is a dotfiles manager, supporting templating, and per-host configurations",
    }
)

func Execute() error {
    return rootCmd.Execute()
}
