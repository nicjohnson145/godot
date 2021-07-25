package lib

import (
	"github.com/hashicorp/go-multierror"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/nicjohnson145/godot/internal/util"
)

type ItemRunner interface {
	RunSingle(Item) error
	RunAll([]Item) error
}

type Controller struct {
	homeDirGetter util.HomeDirGetter
	repo          Repo
	config        *Config
	builder       *Builder
	runner        ItemRunner
}

type ControllerOpts struct {
	HomeDirGetter util.HomeDirGetter
	Repo          Repo
	Config        *Config
	Builder       *Builder
	Runner        ItemRunner
	NoGit         bool
}

func NewController(opts ControllerOpts) *Controller {
	return &Controller{
		homeDirGetter: opts.HomeDirGetter,
		config:        opts.Config,
		repo:          opts.Repo,
		builder:       opts.Builder,
		runner:        opts.Runner,
	}
}

func (c *Controller) targetIsSet(t string) bool {
	return t != ""
}

func (c *Controller) getTarget(t string) string {
	if t == "" || t == CURRENT {
		t = c.config.Target
	}
	return t
}

func (c *Controller) git_pushAndPull(function func() error) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	if err := function(); err != nil {
		return err
	}

	return c.repo.Push()
}

func (c *Controller) git_pullOnly(function func() error) error {
	if err := c.repo.Pull(); err != nil {
		return err
	}

	return function()
}

func (c *Controller) fuzzyOrArgs(args []string, options []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	idx, err := fuzzyfinder.Find(options, func(i int) string { return options[i] })
	if err != nil {
		return "", err
	}

	return options[idx], nil
}

func (c *Controller) write() error {
	return c.config.Write()
}

func (c *Controller) handleMultiError(errs *multierror.Error, err error) bool {
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, e := range merr.Errors {
				errs = multierror.Append(errs, e)
			}
		} else {
			errs = multierror.Append(errs, err)
		}
	}
	return err != nil
}
