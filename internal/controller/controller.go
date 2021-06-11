package controller

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/nicjohnson145/godot/internal/bootstrap"
	"github.com/nicjohnson145/godot/internal/builder"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/nicjohnson145/godot/internal/util"
)

const (
	AllTarget = "ALL"
)

type Controller struct {
	homeDirGetter util.HomeDirGetter
	repo          repo.Repo
	config        *config.Config
	builder       *builder.Builder
	runner        bootstrap.Runner
}

func NewController(opts ControllerOpts) *Controller {
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

	var rn bootstrap.Runner
	if opts.Runner != nil {
		rn = opts.Runner
	} else {
		rn = bootstrap.NewRunner()
	}

	return &Controller{
		homeDirGetter: getter,
		config:        conf,
		repo:          r,
		builder:       b,
		runner:        rn,
	}
}

func (c *Controller) targetIsSet(t string) bool {
	return t != ""
}

func (c *Controller) getTarget(t string) string {
	if t == "" || t == config.CURRENT {
		t = c.config.Target
	}
	return t
}

func (c *Controller) git_pushAndPull(function func() error) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	if err := function(); err != nil {
		return err
	}

	return c.repo.Push()
}

func (c *Controller) git_pullOnly(function func() error) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	return function()
}

func (c *Controller) Sync(opts SyncOpts) error {
	f := func() error {
		return c.sync(opts)
	}

	return c.git_pullOnly(f)
}

func (c *Controller) sync(opts SyncOpts) error {
	var errs *multierror.Error

	if err := c.builder.Build(opts.Force); err != nil {
		errs = multierror.Append(errs, err)
	}

	if !opts.NoBootstrap {
		impls, err := c.config.GetRelevantBootstrapImpls(c.config.Target)
		if err != nil {
			if merr, ok := err.(*multierror.Error); ok {
				for _, e := range merr.Errors {
					errs = multierror.Append(errs, e)
				}
			} else {
				errs = multierror.Append(errs, err)
			}
			return errs.ErrorOrNil()
		}

		if err := c.runner.RunAll(impls); err != nil {
			if merr, ok := err.(*multierror.Error); ok {
				for _, e := range merr.Errors {
					errs = multierror.Append(errs, e)
				}
			} else {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs.ErrorOrNil()
}

func (c *Controller) Import(file string, as string) error {
	f := func() error {
		// import the file into the repo
		if err := c.builder.Import(file, as); err != nil {
			return err
		}

		// Add the file to the repos config
		_, err := c.config.AddFile(as, file)
		if err != nil {
			return err
		}

		// If everything has gone right up to this point, write the config to disk
		return c.config.Write()
	}
	return c.git_pushAndPull(f)
}

func (c *Controller) ShowFilesEntry(target string, w io.Writer) error {
	f := func() error {
		if c.targetIsSet(target) {
			return c.TargetShowFiles(target, w)
		} else {
			return c.ListAllFiles(w)
		}
	}

	return c.git_pullOnly(f)
}

func (c *Controller) ListAllFiles(w io.Writer) error {
	return c.config.ListAllFiles(w)
}

func (c *Controller) TargetShowFiles(target string, w io.Writer) error {
	target = c.getTarget(target)
	return c.config.ListTargetFiles(target, w)
}

func (c *Controller) TargetAddFile(target string, args []string) error {
	if target != AllTarget {
		return c.targetAddFileSingle(target, args)
	} else {
		return c.targetAddFileAll(target, args)
	}
}

func (c *Controller) targetAddFileSingle(target string, args []string) error {
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

func (c *Controller) targetAddFileAll(target string, args []string) error {
	f := func() error {
		options := c.config.GetAllTemplateNames()
		tmpl, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		var errs *multierror.Error
		for _, target := range c.config.GetAllTargets() {
			if err := c.config.AddTargetFile(target, tmpl); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		if errs == nil {
			return c.write()
		} else {
			return errs.ErrorOrNil()
		}
	}

	return c.git_pushAndPull(f)
}

func (c *Controller) TargetRemoveFile(target string, args []string) error {
	if target != AllTarget {
		return c.targetRemoveFileSingle(target, args)
	} else {
		return c.targetRemoveFileAll(target, args)
	}
}

func (c *Controller) targetRemoveFileSingle(target string, args []string) error {
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

func (c *Controller) targetRemoveFileAll(target string, args []string) error {
	f := func() error {
		options := c.config.GetAllTemplateNames()
		tmpl, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		var errs *multierror.Error
		for _, target := range c.config.GetAllTargets() {
			if err := c.config.RemoveTargetFile(target, tmpl); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		if errs == nil {
			return c.write()
		} else {
			return errs.ErrorOrNil()
		}
	}

	return c.git_pushAndPull(f)
}

func (c *Controller) EditFile(args []string, opts EditOpts) error {
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
			if err := c.sync(SyncOpts{NoBootstrap: true}); err != nil {
				return err
			}
		}

		return nil
	}

	return c.git_pushAndPull(f)
}

func (c *Controller) ShowBootstrapsEntry(target string, w io.Writer) error {
	f := func() error {
		if c.targetIsSet(target) {
			return c.ListAllBootstrapsForTarget(target, w)
		} else {
			return c.ListAllBootstraps(w)
		}
	}

	return c.git_pullOnly(f)
}

func (c *Controller) ListAllBootstraps(w io.Writer) error {
	f := func() error {
		return c.config.ListAllBootstraps(w)
	}
	return c.git_pullOnly(f)
}

func (c *Controller) ListAllBootstrapsForTarget(target string, w io.Writer) error {
	f := func() error {
		return c.config.ListBootstrapsForTarget(w, c.getTarget(target))
	}
	return c.git_pullOnly(f)
}

func (c *Controller) AddBootstrapItem(item, manager, pkg, location string) error {
	f := func() error {
		if !config.IsValidPackageManager(manager) {
			return fmt.Errorf("non-supported package manager of %q", manager)
		}

		c.config.AddBootstrapItem(item, manager, pkg, location)
		return c.write()
	}

	return c.git_pushAndPull(f)
}

func (c *Controller) AddTargetBootstrap(target string, args []string) error {
	if target != AllTarget {
		return c.addTargetBootstrapSingle(target, args)
	} else {
		return c.addTargetBootstrapAll(target, args)
	}
}

func (c *Controller) addTargetBootstrapSingle(target string, args []string) error {
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

func (c *Controller) addTargetBootstrapAll(target string, args []string) error {
	f := func() error {
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

		var errs *multierror.Error
		for _, target := range c.config.GetAllTargets() {
			if err := c.config.AddTargetBootstrap(target, bootstrap); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		if errs == nil {
			return c.write()
		} else {
			return errs.ErrorOrNil()
		}
	}

	return c.git_pushAndPull(f)
}

func (c *Controller) RemoveTargetBootstrap(target string, args []string) error {
	if target != AllTarget {
		return c.removeTargetBootstrapSingle(target, args)
	} else {
		return c.removeTargetBootstrapAll(target, args)
	}
}

func (c *Controller) removeTargetBootstrapSingle(target string, args []string) error {
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

func (c *Controller) removeTargetBootstrapAll(target string, args []string) error {
	f := func() error {
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

		var errs *multierror.Error
		for _, target := range c.config.GetAllTargets() {
			if err := c.config.RemoveTargetBootstrap(target, bootstrap); err != nil {
				errs = multierror.Append(errs, err)
			}
		}

		if errs == nil {
			return c.write()
		} else {
			return errs.ErrorOrNil()
		}
	}

	return c.git_pushAndPull(f)
}

func (c *Controller) getFile(args []string) (filePath string, outErr error) {
	allFiles := c.config.GetAllFiles()
	options := make([]string, 0, len(allFiles))
	for _, dest := range allFiles {
		options = append(options, dest)
	}

	return c.fuzzyOrArgs(args, options)
}

func (c *Controller) editFile(path string, opts EditOpts) error {
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

func (c *Controller) fuzzyOrArgs(args []string, options []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	idx, err := fuzzyfinder.Find(options, func(i int) string { return options[i] })
	if err != nil {
		return "", err
	}

	return options[idx], nil
}

func (c *Controller) write() error {
	return c.config.Write()
}
