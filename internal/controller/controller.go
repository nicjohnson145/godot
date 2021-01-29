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
	TargetShow(string, io.Writer) error
	TargetAdd(string, []string) error
	TargetRemove(string, []string) error
	Edit([]string, EditOpts) error
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

func (c *controller) Sync(opts SyncOpts) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}
	return c.builder.Build(opts.Force)
}

func (c *controller) Import(file string, as string, opts ImportOpts) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	// import the file into the repo
	err := c.builder.Import(file, as)
	if err != nil {
		return err
	}

	// Add the file to the repos config
	name, err := c.config.AddFile(as, file)
	if err != nil {
		return err
	}

	// Potentially add the file to the current target
	if !opts.NoAdd {
		err = c.config.AddTargetFile(c.config.Target, name)
	}

	// If everything has gone right up to this point, write the config to disk
	if err == nil {
		err = c.config.Write()
	}

	if err := c.repo.Push(); err != nil {
		return err
	}

	return err
}

func (c *controller) ShowEntry(target string, w io.Writer) error {
	if c.targetIsSet(target) {
		return c.TargetShow(target, os.Stdout)
	} else {
		return c.ListAllFiles(os.Stdout)
	}
}

func (c *controller) ListAllFiles(w io.Writer) error {
	return c.config.ListAllFiles(w)
}

func (c *controller) TargetShow(target string, w io.Writer) error {
	target = c.getTarget(target)
	return c.config.ListTargetFiles(target, w)
}

func (c *controller) TargetAdd(target string, args []string) error {
	target = c.getTarget(target)

	if err := c.repo.Pull(); err != nil {
		return err
	}

	options := c.config.GetAllTemplateNames()
	tmpl, err := c.fuzzyOrArgs(args, options)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Aborting...")
			return nil
		}
		return err
	}

	err = c.config.AddTargetFile(target, tmpl)
	if err != nil {
		return err
	}

	err = c.write()
	if err != nil {
		return err
	}
if err := c.repo.Push(); err != nil {
		return err
	}
	return nil
}

func (c *controller) TargetRemove(target string, args []string) error {
	target = c.getTarget(target)

	if err := c.repo.Pull(); err != nil {
		return err
	}

	options := c.config.GetAllTemplateNamesForTarget(target)
	tmpl, err := c.fuzzyOrArgs(args, options)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Aborting...")
			return nil
		}
		return err
	}

	err = c.config.RemoveTargetFile(target, tmpl)
	if err != nil {
		return err
	}

	err = c.write()
	if err != nil {
		return err
	}

	if err := c.repo.Push(); err != nil {
		return err
	}
	return nil
}

func (c *controller) Edit(args []string, opts EditOpts) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	path, err := c.getFile(args)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Aborting...")
			return nil
		}
		return err
	}

	err = c.editFile(util.ReplacePrefix(path, "~/", c.config.Home), opts)
	if err != nil {
		return err
	}

	if !opts.NoSync {
		err = c.Sync(SyncOpts{})
		if err != nil {
			return err
		}
	}

	if err := c.repo.Push(); err != nil {
		return err
	}

	return nil
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
