package lib

import (
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/go-fuzzyfinder"
)

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
		if !IsValidPackageManager(manager) {
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
