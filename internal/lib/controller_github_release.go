package lib

import (
	"fmt"
	"io"

	"github.com/ktr0731/go-fuzzyfinder"
)

func (c *Controller) ShowGithubRelease(target string, w io.Writer) error {
	f := func() error {
		if c.targetIsSet(target) {
			return c.showTargetGithubRelease(target, w)
		} else {
			return c.showAllGithubRelease(w)
		}
	}
	return c.git_pullOnly(f)
}

func (c *Controller) showTargetGithubRelease(target string, w io.Writer) error {
	target = c.getTarget(target)
	return c.config.TargetShowGithubRelease(target, w)
}

func (c *Controller) showAllGithubRelease(w io.Writer) error {
	return c.config.ShowAllGithubRelease(w)
}

func (c *Controller) TargetUseGithubRelease(target string, name string, location string, trackUpdates bool) error {
	f := func() error {
		target = c.getTarget(target)
		err := c.config.TargetUseGithubRelease(target, name, location, trackUpdates)
		if err != nil {
			return err
		}
		return c.write()
	}
	return c.git_pushAndPull(f)
}

func (c *Controller) TargetCeaseGithubRelease(target string, args []string) error {
	f := func() error {
		target = c.getTarget(target)
		options, err := c.config.GetAllGithubReleaseNamesForTarget(target)
		if err != nil {
			return err
		}

		ghr, err := c.fuzzyOrArgs(args, options)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				fmt.Println("Aborting...")
				return nil
			}
			return err
		}

		err = c.config.TargetCeaseGithubRelease(target, ghr)
		if err != nil {
			return err
		}
		return c.write()
	}
	return c.git_pushAndPull(f)
}
