package bootstrap

import (
	"github.com/hashicorp/go-multierror"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r Runner) RunAll(items []Item) error {
	var errs *multierror.Error
	for _, item := range items {
		if err := r.runSingleItem(item); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

func (r Runner) RunSingle(item Item) error {
	return r.runSingleItem(item)
}

func (r Runner) runSingleItem(item Item) error {
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
