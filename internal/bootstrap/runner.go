package bootstrap

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-multierror"
	"github.com/nicjohnson145/godot/internal/config"
)

type runner struct{}

func NewRunner() Runner {
	return runner{}
}

func (r runner) ToItems(impls []config.BootstrapImpl) []Item {
	items := make([]Item, 0, len(impls))

	for _, impl := range impls {
		items = append(items, r.ToItem(impl))
	}

	return items
}

func (r runner) ToItem(impl config.BootstrapImpl) Item {
	switch impl.Name {
	case config.APT:
		return NewAptItem(impl.Item.Name)
	case config.BREW:
		return NewBrewItem(impl.Item.Name)
	case config.GIT:
		return NewRepoItem(impl.Item.Name, impl.Item.Location)
	default:
		panic(fmt.Sprintf("Unknown bootstrap impl of %v", impl))
	}
}

func (r runner) RunAll(impls []config.BootstrapImpl) error {
	items := r.ToItems(impls)
	var errs *multierror.Error
	for _, item := range items {
		if err := r.runSingle(item); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

func (r runner) RunSingle(impl config.BootstrapImpl) error {
	item := r.ToItem(impl)
	return r.runSingle(item)
}

func (r runner) runSingle(item Item) error {
	// Check if the item is already installed, and bail early
	installed, err := item.Check()
	if err != nil {
		return err
	}
	if installed {
		return nil
	}

	// Otherwise, install it
	return item.Install()
}

func runCmd(bin string, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func getReturnCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode(), nil
	} else {
		return -1, err
	}
}
