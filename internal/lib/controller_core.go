package lib

import (
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
	var getter util.HomeDirGetter
	if opts.HomeDirGetter != nil {
		getter = opts.HomeDirGetter
	} else {
		getter = &util.OSHomeDir{}
	}

	var conf *Config
	if opts.Config != nil {
		conf = opts.Config
	} else {
		conf = NewConfig(getter)
	}

	var r Repo
	if opts.Repo != nil {
		r = opts.Repo
	} else {
		if opts.NoGit {
			r = NoopRepo{}
		} else {
			r = NewShellGitRepo(conf.DotfilesRoot)
		}
	}

	var b *Builder
	if opts.Builder != nil {
		b = opts.Builder
	} else {
		b = &Builder{
			Getter: getter,
			Config: conf,
		}
	}

	var rn ItemRunner
	if opts.Runner != nil {
		rn = opts.Runner
	} else {
		rn = NewRunner()
	}

	return &Controller{
		homeDirGetter: getter,
		config:        conf,
		repo:          r,
		builder:       b,
		runner:        rn,
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
