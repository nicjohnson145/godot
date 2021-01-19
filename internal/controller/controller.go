package controller

import (
	"errors"
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

var userCancelError = errors.New("user canceled")

type Controller interface {
	Sync(SyncOpts) error
	Import(string, string, ImportOpts) error
	ListAll(io.Writer)
	TargetShow(string, io.Writer)
	TargetAdd(string, string, AddOpts) error
	TargetRemove(string, string, RemoveOpts) error
	Edit([]string, EditOpts) error
}

type controller struct {
	homeDirGetter util.HomeDirGetter
	repo          repo.Repo
	Config        *config.Config
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
		r = repo.NewShellGitRepo(conf.DotfilesRoot)
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
		Config:        conf,
		repo:          r,
		builder:       b,
	}
}

func (c *controller) Sync(opts SyncOpts) error {
	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	return c.builder.Build(opts.Force)
}

func (c *controller) Import(file string, as string, opts ImportOpts) error {
	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	// import the file into the repo
	err := c.builder.Import(file, as)
	if err != nil {
		return err
	}

	// Add the file to the repos config
	var name string
	if as != "" {
		name, err = c.Config.AddFile(as, file)
	} else {
		name, err = c.Config.ManageFile(file)
	}
	if err != nil {
		return err
	}

	// Potentially add the file to the current target
	if !opts.NoAdd {
		err = c.Config.AddToTarget(c.Config.Target, name)
	}

	// If everything has gone right up to this point, write the config to disk
	if err == nil {
		err = c.Config.Write()
	}

	if !opts.NoGit {
		err = c.repo.Push()
		if err != nil {
			return err
		}
	}

	return err
}

func (c *controller) write() error {
	return c.Config.Write()
}

func (c *controller) ListAll(w io.Writer) {
	c.Config.ListAllFiles(w)
}

func (c *controller) TargetShow(target string, w io.Writer) {
	c.Config.ListTargetFiles(target, w)
}

func (c *controller) TargetAdd(target string, file string, opts AddOpts) error {
	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	err := c.Config.AddToTarget(target, file)
	if err != nil {
		return err
	}

	err = c.write()
	if err != nil {
		return err
	}

	if !opts.NoGit {
		err := c.repo.Push()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) TargetRemove(target string, file string, opts RemoveOpts) error {
	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	err := c.Config.RemoveFromTarget(target, file)
	if err != nil {
		return err
	}

	err = c.write()
	if err != nil {
		return err
	}

	if !opts.NoGit {
		err := c.repo.Push()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) Edit(args []string, opts EditOpts) error {
	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	path, err := c.getFile(args)
	if err != nil {
		if err == userCancelError {
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
		err = c.Sync(SyncOpts{NoGit: true})
		if err != nil {
			return err
		}
	}

	if !opts.NoGit {
		err = c.repo.Push()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) getFile(args []string) (filePath string, outErr error) {
	if len(args) == 0 {
		targetFiles := c.Config.GetTargetFiles()
		defer func() {
			if err := recover(); err != nil {
				// Ctrl-C ing inside fuzzy finder is a panic, catch that panic and abort the edit
				// operation
				filePath = ""
				outErr = userCancelError
			}
		}()

		idx, err := fuzzyfinder.Find(
			targetFiles,
			func(i int) string {
				if i == -1 {
					return ""
				}
				return targetFiles[i].DestinationPath
			},
		)
		if err != nil {
			return "", err
		}
		return targetFiles[idx].DestinationPath, nil
	} else {
		return args[0], nil
	}
}

func (c *controller) editFile(path string, opts EditOpts) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("'edit' requires $EDITOR to be set")
	}

	file, err := c.Config.GetTemplateFromFullPath(path)
	if err != nil {
		return err
	}

	proc := exec.Command(editor, file)
	proc.Stdout = os.Stdout
	proc.Stdin = os.Stdin
	return proc.Run()
}
