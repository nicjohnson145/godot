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

type Controller struct {
	homeDirGetter util.HomeDirGetter
	repo          repo.Repo
	Config        *config.Config
	builder       *builder.Builder
}

func (c *Controller) ensureSetup() {
	if c.homeDirGetter == nil {
		c.homeDirGetter = &util.OSHomeDir{}
	}

	if c.Config == nil {
		c.Config = config.NewConfig(c.homeDirGetter)
	}

	if c.repo == nil {
		c.repo = repo.NewShellGitRepo(c.Config.DotfilesRoot)
	}

	if c.builder == nil {
		c.builder = &builder.Builder{
			Getter: c.homeDirGetter,
			Config: c.Config,
		}
	}

}

func (c *Controller) Sync(opts SyncOpts) error {
	c.ensureSetup()

	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	return c.builder.Build(opts.Force)
}

func (c *Controller) Import(file string, as string, opts ImportOpts) error {
	c.ensureSetup()

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

func (c *Controller) write() error {
	return c.Config.Write()
}

func (c *Controller) ListAll(w io.Writer) {
	c.ensureSetup()
	c.Config.ListAllFiles(w)
}

func (c *Controller) TargetShow(target string, w io.Writer) {
	c.ensureSetup()
	c.Config.ListTargetFiles(target, w)
}

func (c *Controller) TargetAdd(target string, file string, opts AddOpts) error {
	c.ensureSetup()

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

func (c *Controller) TargetRemove(target string, file string, opts RemoveOpts) error {
	c.ensureSetup()

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

func (c *Controller) Edit(args []string, opts EditOpts) error {
	c.ensureSetup()

	if !opts.NoGit {
		err := c.repo.Pull()
		if err != nil {
			return err
		}
	}

	path, err := c.getFile(args)
	if err != nil {
		return err
	}

	err = c.editFile(path, opts)
	if err != nil {
		return err
	}

	if !opts.NoSync{
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

func (c *Controller) getFile(args []string) (string, error) {
	if len(args) == 0 {
		targetFiles := c.Config.GetTargetFiles()
		idx, err := fuzzyfinder.Find(
			targetFiles,
			func(i int) string { return targetFiles[i].DestinationPath },
		)
		if err != nil {
			return "", err
		}
		return targetFiles[idx].DestinationPath, nil
	} else {
		return args[0], nil
	}
}

func (c *Controller) editFile(path string, opts EditOpts) error {
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
