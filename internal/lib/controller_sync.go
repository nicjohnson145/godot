package lib

import (
	"github.com/hashicorp/go-multierror"
)

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
		var allImpls []Item

		bsImpls, err := c.config.GetRelevantBootstrapImpls(c.config.Target)
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
		allImpls = append(allImpls, bsImpls...)

		ghrImpls, err := c.config.GetGithubReleaseImplForTarget(c.config.Target)
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
		allImpls = append(allImpls, ghrImpls...)

		if err := c.runner.RunAll(allImpls); err != nil {
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
