package controller

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/nicjohnson145/godot/internal/builder"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/nicjohnson145/godot/internal/util"
)

type Controller interface {
	Sync(SyncOpts) error
	Import(string, string, ImportOpts) error
	ListAllFiles(io.Writer) error
	TargetShowFiles(string, io.Writer) error
	TargetAddFile(string, []string) error
	TargetRemoveFile(string, []string) error
	EditFile([]string, EditOpts) error
}

type controller struct {
	homeDirGetter util.HomeDirGetter
	repo          repo.Repo
	config        *config.Config
	builder       *builder.Builder
}

func NewController(opts ControllerOpts) *controller {
	var getter util.HomeDirGetter
	if opts.HomeDirGetter != nil {
		getter = opts.HomeDirGetter
	} else {
		getter = &util.OSHomeDir{}
	}

	var conf *config.Config
	if opts.Config != nil {
		conf = opts.Config
	} else {
		conf = config.NewConfig(getter)
	}

	var r repo.Repo
	if opts.Repo != nil {
		r = opts.Repo
	} else {
		if opts.NoGit {
			r = repo.NoopRepo{}
		} else {
			r = repo.NewShellGitRepo(conf.DotfilesRoot)
		}
	}

	var b *builder.Builder
	if opts.Builder != nil {
		b = opts.Builder
	} else {
		b = &builder.Builder{
			Getter: getter,
			Config: conf,
		}
	}

	return &controller{
		homeDirGetter: getter,
		config:        conf,
		repo:          r,
		builder:       b,
	}
}

func (c *controller) targetIsSet(t string) bool {
	return t == ""
}

func (c *controller) getTarget(t string) string {
	if t == "" || t == config.CURRENT {
		t = c.config.Target
	}
	return t
}

func (c *controller) git_pushAndPull(function func() error) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	if err := function(); err != nil {
		return err
	}

	return c.repo.Push()
}

func (c *controller) git_pullOnly(function func() error) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	return function()
}

func (c *controller) Sync(opts SyncOpts) error {
	f := func() error {
		return c.builder.Build(opts.Force)
	}

	return c.git_pullOnly(f)
}

func (c *controller) Import(file string, as string, opts ImportOpts) error {
	f := func() error {
		// import the file into the repo
		if err := c.builder.Import(file, as); err != nil {
			return err
		}

		// Add the file to the repos config
		name, err := c.config.AddFile(as, file)
		if err != nil {
			return err
		}

		// Potentially add the file to the current target
		if !opts.NoAdd {
			if err := c.config.AddTargetFile(c.config.Target, name); err != nil {
				return err
			}
		}

		// If everything has gone right up to this point, write the config to disk
		return c.config.Write()
	}
	return c.git_pushAndPull(f)
}

func (c *controller) ShowFilesEntry(target string, w io.Writer) error {
	f := func() error {
		if c.targetIsSet(target) {
			return c.TargetShowFiles(target, os.Stdout)
		} else {
			return c.ListAllFiles(os.Stdout)
		}
	}

	return c.git_pullOnly(f)
}

func (c *controller) ListAllFiles(w io.Writer) error {
	return c.config.ListAllFiles(w)
}

func (c *controller) TargetShowFiles(target string, w io.Writer) error {
	target = c.getTarget(target)
	return c.config.ListTargetFiles(target, w)
}

func (c *controller) TargetAddFile(target string, args []string) error {
	f := func() error {
		target = c.getTarget(target)

		options := c.config.GetAllTemplateNames()
		tmpl, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		if err := c.config.AddTargetFile(target, tmpl); err != nil {
			return err
		}

		return c.write()
	}

	return c.git_pushAndPull(f)
}

func (c *controller) TargetRemoveFile(target string, args []string) error {
	f := func() error {
		target = c.getTarget(target)

		options := c.config.GetAllTemplateNames()
		tmpl, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		if err := c.config.RemoveTargetFile(target, tmpl); err != nil {
			return err
		}

		return c.write()
	}

	return c.git_pushAndPull(f)
}

func (c *controller) EditFile(args []string, opts EditOpts) error {
	f := func() error {
		path, err := c.getFile(args)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		if err := c.editFile(util.ReplacePrefix(path, "~/", c.config.Home), opts); err != nil {
			return err
		}

		if !opts.NoSync {
			if err := c.Sync(SyncOpts{}); err != nil {
				return err
			}
		}

		return nil
	}

	return c.git_pushAndPull(f)
}

func (c *controller) ShowBootstrapsEntry(target string, w io.Writer) error {
	f := func() error {
		if c.targetIsSet(target) {
			return c.ListAllBootstrapsForTarget(target, os.Stdout)
		} else {
			return c.ListAllBootstraps(os.Stdout)
		}
	}

	return c.git_pullOnly(f)
}

func (c *controller) ListAllBootstraps(w io.Writer) error {
	f := func() error {
		return c.config.ListAllBootstraps(w)
	}
	return c.git_pullOnly(f)
}

func (c *controller) ListAllBootstrapsForTarget(target string, w io.Writer) error {
	f := func() error {
		return c.config.ListBootstrapsForTarget(w, target)
	}
	return c.git_pullOnly(f)
}

func (c *controller) AddBootstrapItem(item, manager, pkg, location string) error {
	f := func() error {
		if !config.IsValidPackageManager(manager) {
			return fmt.Errorf("non-supported package manager of %q", manager)
		}

		c.config.AddBootstrapItem(item, manager, pkg, location)
		return c.write()
	}

	return c.git_pushAndPull(f)
}

func (c *controller) AddTargetBootstrap(target string, args []string) error {
	f := func() error {
		target = c.getTarget(target)

		all := c.config.GetAllBootstraps()
		if len(all) == 0 {
			return fmt.Errorf("No bootstraps defined")
		}

		options := make([]string, 0, len(all))
		for key := range all {
			options = append(options, key)
		}
		bootstrap, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		if err := c.config.AddTargetBootstrap(target, bootstrap); err != nil {
			return err
		}

		return c.write()
	}

	return c.git_pushAndPull(f)
}

func (c *controller) RemoveTargetBootstrap(target string, args []string) error {
	f := func() error {
		target = c.getTarget(target)
		options := c.config.GetBootstrapTargetsForTarget(target)
		bootstrap, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		if err := c.config.RemoveTargetBootstrap(target, bootstrap); err != nil {
			return err
		}

		return c.write()
	}

	return c.git_pushAndPull(f)
}

func (c *controller) getFile(args []string) (filePath string, outErr error) {
	allFiles := c.config.GetAllFiles()
	options := make([]string, 0, len(allFiles))
	for _, dest := range allFiles {
		options = append(options, dest)
	}

	return c.fuzzyOrArgs(args, options)
}

func (c *controller) editFile(path string, opts EditOpts) error {
	// TODO: opts not necessary
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("'edit' requires $EDITOR to be set")
	}

	file, err := c.config.GetTemplateFromFullPath(path)
	if err != nil {
		return err
	}

	proc := exec.Command(editor, file)
	proc.Stdout = os.Stdout
	proc.Stdin = os.Stdin
	return proc.Run()
}

func (c *controller) fuzzyOrArgs(args []string, options []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	idx, err := fuzzyfinder.Find(options, func(i int) string { return options[i] })
	if err != nil {
		return "", err
	}

	return options[idx], nil
}

func (c *controller) write() error {
	return c.config.Write()
}
