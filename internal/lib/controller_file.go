package lib

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/nicjohnson145/godot/internal/util"
)

func (c *Controller) AddFile(file string, as string) error {
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

func (c *Controller) ShowFile(target string, w io.Writer) error {
	f := func() error {
		if c.targetIsSet(target) {
			return c.showTargetFiles(target, w)
		} else {
			return c.showAllFiles(w)
		}
	}

	return c.git_pullOnly(f)
}

func (c *Controller) showAllFiles(w io.Writer) error {
	return c.config.ListAllFiles(w)
}

func (c *Controller) showTargetFiles(target string, w io.Writer) error {
	target = c.getTarget(target)
	return c.config.ListTargetFiles(target, w)
}

func (c *Controller) TargetUseFile(target string, args []string) error {
	if target != AllTarget {
		return c.targetUseFileSingle(target, args)
	} else {
		return c.targetUseFileAll(target, args)
	}
}

func (c *Controller) targetUseFileSingle(target string, args []string) error {
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

func (c *Controller) targetUseFileAll(target string, args []string) error {
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

func (c *Controller) TargetCeaseFile(target string, args []string) error {
	if target != AllTarget {
		return c.targetCeaseFileSingle(target, args)
	} else {
		return c.targetCeaseFileAll(target, args)
	}
}

func (c *Controller) targetCeaseFileSingle(target string, args []string) error {
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

func (c *Controller) targetCeaseFileAll(target string, args []string) error {
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
				if !errors.Is(err, NotFoundError) {
					errs = multierror.Append(errs, err)
				}
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
