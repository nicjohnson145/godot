package main

import (
	"fmt"

	"github.com/nicjohnson145/godot/internal/lib"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Version info set by goreleaser
var (
	version = "development"
	date    = "unknown"
)

func main() {
	if err := buildCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func buildCommand() *cobra.Command {
	syncOpts := lib.SyncOpts{}
	var debug bool
	var verbose bool

	rootCmd := &cobra.Command{
		Use:   "godot",
		Short: "A dotfiles manager",
		Long:  "A staticly linked dotfiles manager written in Go",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.WarnLevel)
			if verbose {
				log.SetLevel(log.InfoLevel)
			}
			if debug {
				log.SetLevel(log.DebugLevel)
			}
		},
	}
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync configuration",
		Long:  "Sync local filesystem with configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return lib.Sync(syncOpts)
		},
	}
	syncCmd.Flags().BoolVarP(&syncOpts.Quick, "quick", "q", false, "Run a quick sync, skipping some stages")
	syncCmd.Flags().StringSliceVarP(&syncOpts.Ignore, "ignore", "i", []string{}, "Ignore these configs")
	syncCmd.Flags().BoolVar(&syncOpts.NoVault, "no-vault", false, "Ignore vault lookup directives in templates")
	syncCmd.Flags().StringSliceVarP(&syncOpts.Executors, "executors", "e", []string{}, fmt.Sprintf("Limit run to only these executor types (valid values: %v)", lib.ExecutorTypeNames()))
	rootCmd.AddCommand(syncCmd)

	validateCmd := &cobra.Command{
		Use:   "validate <path-to-config>",
		Args:  cobra.ExactArgs(1),
		Short: "Validate configuration",
		Long:  "Validate a configuration file on disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			return lib.Validate(args[0])
		},
	}
	rootCmd.AddCommand(validateCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Args:  cobra.NoArgs,
		Long:  "Show version info and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Build Version: ", version)
			fmt.Println("Build Time: ", date)
		},
	}
	rootCmd.AddCommand(versionCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Self update",
		Args:  cobra.NoArgs,
		Long:  "Check for newer version and install if found",
		RunE: func(cmd *cobra.Command, args []string) error {
			return lib.SelfUpdate(version)
		},
	}
	rootCmd.AddCommand(updateCmd)

	return rootCmd
}
