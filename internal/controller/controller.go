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
	ListAll(io.Writer)
	TargetShow(string, io.Writer)
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

func (c *controller) getTarget(t string) string {
	if t == "" {
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
	var name string
	if as != "" {
		name, err = c.config.AddFile(as, file)
	} else {
		name, err = c.config.ManageFile(file)
	}
	if err != nil {
		return err
	}

	// Potentially add the file to the current target
	if !opts.NoAdd {
		err = c.config.AddToTarget(c.config.Target, name)
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

func (c *controller) ListAll(w io.Writer) {
	c.config.ListAllFiles(w)
}

func (c *controller) TargetShow(target string, w io.Writer) {
	target = c.getTarget(target)
	c.config.ListTargetFiles(target, w)
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

	err = c.config.AddToTarget(target, tmpl)
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

	options := c.config.GetTemplatesNamesForTarget(target)
	tmpl, err := c.fuzzyOrArgs(args, options)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Aborting...")
			return nil
		}
		return err
	}

	err = c.config.RemoveFromTarget(target, tmpl)
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

	err = c.editFile(path, opts)
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
	targetFiles := c.config.GetTargetFiles()
	options := make([]string, 0, len(targetFiles))
	for _, fl := range targetFiles {
		options = append(options, fl.DestinationPath)
	}

	return c.fuzzyOrArgs(args, options)
}

func (c *controller) editFile(path string, opts EditOpts) error {
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
