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
		if c.handleMultiError(errs, err) {
			return errs
		}
		allImpls = append(allImpls, bsImpls...)

		ghrImpls, err := c.config.GetGithubReleaseImplForTarget(c.config.Target)
		if c.handleMultiError(errs, err) {
			return errs
		}
		allImpls = append(allImpls, ghrImpls...)

		err = c.runner.RunAll(allImpls)
		if c.handleMultiError(errs, err) {
			return errs
		}
	}

	return errs.ErrorOrNil()
}
