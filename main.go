package main

import (
	"os"
	"strings"

	"github.com/nicjohnson145/godot/internal/lib"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/spf13/cobra"
)

func main() {
	m := New(lib.ControllerOpts{})
	if err := m.rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type Main struct {
	dependencies lib.ControllerOpts

	// Flags
	force        bool
	noBootstrap  bool
	noGit        bool
	target       string
	noSync       bool
	name         string
	manager      string
	location     string
	as           string
	trackUpdates bool

	// Commands
	rootCmd *cobra.Command

	// Cease Cmds
	ceaseCmd              *cobra.Command
	ceaseFileCmd          *cobra.Command
	ceaseBootstrapCmd     *cobra.Command
	ceaseGithubReleaseCmd *cobra.Command

	// Sync Cmds
	syncCmd *cobra.Command

	// List Cmds
	listCmd               *cobra.Command
	listFilesCmd          *cobra.Command
	listBootstrapsCmd     *cobra.Command
	listGithubReleasesCmd *cobra.Command

	// Edit Cmds
	editCmd *cobra.Command

	// Use Cmds
	useCmd           *cobra.Command
	useFileCmd       *cobra.Command
	useBootstrapCmd  *cobra.Command
	useGithubRelease *cobra.Command

	// Add Cmds
	addCmd          *cobra.Command
	addFileCmd      *cobra.Command
	addBootstrapCmd *cobra.Command
}

func New(dependencies lib.ControllerOpts) Main {
	m := Main{
		dependencies: dependencies,
	}

	if m.dependencies.HomeDirGetter == nil {
		m.dependencies.HomeDirGetter = &util.OSHomeDir{}
	}
	if m.dependencies.Config == nil {
		m.dependencies.Config = lib.NewConfig(m.dependencies.HomeDirGetter)
	}
	if m.dependencies.Builder == nil {
		m.dependencies.Builder = &lib.Builder{
			Config: m.dependencies.Config,
		}
	}
	if m.dependencies.Runner == nil {
		m.dependencies.Runner = lib.NewRunner()
	}
	if m.dependencies.Repo == nil {
		// m.dependencies.Repo = lib.NewShellGitRepo(m.dependencies.Config.DotfilesRoot)
		m.dependencies.Repo = lib.PureGoRepo{
			Path: m.dependencies.Config.DotfilesRoot,
			User: m.dependencies.Config.GithubUser,
		}
	}

	// Root cmd
	m.rootCmd = &cobra.Command{
		Use:   "godot",
		Short: "A dotfiles manager",
		Long:  "A staticly linked dotfiles manager written in Go",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
	}

	// Root flags
	m.rootCmd.PersistentFlags().BoolVarP(
		&m.noGit,
		"no-git",
		"g",
		false,
		"Do not perform any git operations (this will require the git operations to be completed manually)",
	)
	m.rootCmd.PersistentFlags().StringVar(
		&m.target,
		"target",
		"",
		"Apply the command to the given target, not supplying a value (i.e --target vs --target=a), will result in the current target being used. "+
			"The special value of 'ALL' will apply the change to all available targets",
	)
	m.rootCmd.PersistentFlags().Lookup("target").NoOptDefVal = lib.CURRENT

	// Cease Cmds
	m.ceaseCmd = &cobra.Command{
		Use:   "cease",
		Short: "Remove an object to a target",
	}
	m.ceaseFileCmd = &cobra.Command{
		Use:   "file [file]",
		Short: "Remove an file from a target",
		Long: "Remove a file from a target. If a template name is not given, a selection prompt will " +
			"open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.TargetCeaseFile(m.target, args)
		},
	}
	m.ceaseBootstrapCmd = &cobra.Command{
		Use:   "bootstrap [bootstrap]",
		Short: "Remove an item to be bootstrapped from a target",
		Long: "Remove an item to be bootstrapped from a target. If a bootstrap name is not given, a " +
			"selection prompt wil open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.TargetCeaseBootstrap(m.target, args)
		},
	}
	m.ceaseGithubReleaseCmd = &cobra.Command{
		Use:     "github-release <ghr>",
		Aliases: []string{"ghr"},
		Short:   "Remove a github release from a target",
		Long: "Remove a github release from a target. If a release name is not given, a prompt " +
			"will open. If no target is given, the current target will be used.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.TargetCeaseGithubRelease(m.target, args)
		},
	}
	m.ceaseCmd.AddCommand(
		m.ceaseBootstrapCmd,
		m.ceaseFileCmd,
		m.ceaseGithubReleaseCmd,
	)

	// Sync Cmds
	m.syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Build and symlink dotfiles",
		Long:  "Compile templates, symlink them to their final destinations, install configured bootstraps",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.Sync(lib.SyncOpts{Force: m.force, NoBootstrap: m.noBootstrap})
		},
	}
	// Sync Flags
	m.syncCmd.Flags().BoolVarP(&m.force, "force", "f", false, "If a target file already exists, delete it then symlink it")
	m.syncCmd.Flags().BoolVarP(&m.noBootstrap, "no-bootstrap", "b", false, "Don't attempt to bootstrap this host, only symlink")

	// List Cmds
	m.listCmd = &cobra.Command{
		Use:   "list [object]",
		Short: "List information",
		Args:  cobra.ExactArgs(1),
	}
	m.listFilesCmd = &cobra.Command{
		Use:   "files",
		Short: "List files managed by godot",
		Long:  "List files managed by godot, if -t/--target is given, only that targets files will be listed",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.ShowFile(m.target, cmd.OutOrStdout())
		},
	}
	m.listBootstrapsCmd = &cobra.Command{
		Use:   "bootstraps",
		Short: "List bootstrap items managed by godot",
		Long:  "List bootstrap items managed by godot, if -t/--target is given, only that targets bootstrap items will be listed",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.ShowBootstrap(m.target, cmd.OutOrStdout())
		},
	}
	m.listGithubReleasesCmd = &cobra.Command{
		Use:     "github-release",
		Aliases: []string{"ghr"},
		Short:   "List github releases managed by godot",
		Long:    "List github releases managed by godot, if -t/--target is given, only that targets in-use releases items will be listed",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.ShowGithubRelease(m.target, cmd.OutOrStdout())
		},
	}
	m.listCmd.AddCommand(
		m.listFilesCmd,
		m.listBootstrapsCmd,
		m.listGithubReleasesCmd,
	)

	// Edit Cmd
	m.editCmd = &cobra.Command{
		Use:   "edit [file]",
		Short: "Edit a file managed by godot",
		Long:  "Edit a file managed by godot, if no file is suppled a prompt will be displayed",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.EditFile(args, lib.EditOpts{NoSync: m.noSync})
		},
	}
	m.editCmd.Flags().BoolVarP(&m.noSync, "no-sync", "s", false, "Do not issue a sync after editing")

	// Use Cmds
	m.useCmd = &cobra.Command{
		Use:   "use",
		Short: "Add an object to a target",
	}
	m.useFileCmd = &cobra.Command{
		Use:   "file [file]",
		Short: "Add an file to a target",
		Long: "Add a file to a target. If a template name is not given, a selection prompt will " +
			"open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.TargetUseFile(m.target, args)
		},
	}
	m.useBootstrapCmd = &cobra.Command{
		Use:   "bootstrap [bootstrap]",
		Short: "Add an item to be bootstrapped to a target",
		Long: "Add an item to be bootstrapped to a target. If a bootstrap name is not given, a " +
			"selection prompt wil open. If no target is given, the current target will be used",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.TargetUseBootstrap(m.target, args)
		},
	}
	m.useGithubRelease = &cobra.Command{
		Use:     "github-release <ghr>",
		Aliases: []string{"ghr"},
		Short:   "Add a github release to a target",
		Long:    "Add a github release to a target. If no target is given, the current target will be used.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.TargetUseGithubRelease(m.target, args[0], m.location, m.trackUpdates)
		},
	}
	//m.useGithubRelease.Flags().BoolVarP(
	//    &m.trackUpdates,
	//    "track-updates",
	//    "u",
	//    true,
	//    "If new updates are found, get them. (Default: true)",
	//)
	m.useGithubRelease.Flags().StringVarP(
		&m.location,
		"location",
		"l",
		"",
		"The location to place the resulting binary. (Default: ~/bin)",
	)
	m.useCmd.AddCommand(
		m.useFileCmd,
		m.useBootstrapCmd,
		m.useGithubRelease,
	)

	// Add Cmds
	m.addCmd = &cobra.Command{
		Use:   "add [object]",
		Short: "Add an object",
		Args:  cobra.ExactArgs(1),
	}
	m.addFileCmd = &cobra.Command{
		Use:   "file file",
		Short: "Add a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.AddFile(args[0], m.as)
		},
	}
	m.addBootstrapCmd = &cobra.Command{
		Use:   "bootstrap bs",
		Short: "Add a bootstrap item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := lib.NewController(m.controllerOpts())
			return c.AddBootstrap(args[0], m.manager, m.name, m.location)
		},
	}
	m.addCmd.AddCommand(
		m.addFileCmd,
		m.addBootstrapCmd,
	)
	m.addFileCmd.Flags().StringVar(
		&m.as,
		"as",
		"",
		"Override the default template name (useful for resolving name conflicts)",
	)
	m.addBootstrapCmd.Flags().StringVarP(
		&m.name,
		"name",
		"n",
		"",
		"The name of the package for this manager, for git this is the repo URL",
	)
	m.addBootstrapCmd.MarkFlagRequired("name")
	m.addBootstrapCmd.Flags().StringVarP(
		&m.manager,
		"manager",
		"m",
		"",
		"The name of the package manager, one of "+strings.Join(lib.ValidManagers, ", "),
	)
	m.addBootstrapCmd.MarkFlagRequired("manager")
	m.addBootstrapCmd.Flags().StringVarP(
		&m.location,
		"location",
		"l",
		"",
		"When using the git manager, where the repo should be checked out to, defaults to home directory",
	)

	m.rootCmd.AddCommand(
		m.ceaseCmd,
		m.syncCmd,
		m.listCmd,
		m.editCmd,
		m.useCmd,
		m.addCmd,
	)

	return m
}

func (m Main) controllerOpts() lib.ControllerOpts {
	m.dependencies.NoGit = m.noGit
	return m.dependencies
}
