package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const CURRENT = "<CURRENT>"

func init() {
	rootCmd.PersistentFlags().BoolVarP(
		&noGit,
		"no-git",
		"g",
		false,
		"Do not perform any git operations (this will require the git operations to be completed manually)",
	)
	rootCmd.PersistentFlags().StringVarP(
		&target,
		"target",
		"t",
		"",
		"Apply the command to the given target, not supplying a value (i.e --target vs --target=a), will result in the current target being used",
	)
	rootCmd.PersistentFlags().Lookup("target").NoOptDefVal = CURRENT
}

var (
	noGit bool
	target string

	rootCmd = &cobra.Command{
		Use:   "godot",
		Short: "A dotfiles manager",
		Long:  "A staticly linked dotfiles manager written in Go",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
