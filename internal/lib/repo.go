package lib

import (
	"errors"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Repo interface {
	Pull() error
	Push() error
}

var _ Repo = (*NoopRepo)(nil)
var _ Repo = (*PureGoRepo)(nil)

type NoopRepo struct{}

func (n NoopRepo) Push() error { return nil }
func (n NoopRepo) Pull() error { return nil }

type PureGoRepo struct {
	Path string
	User string
}

func (p PureGoRepo) ensureClean() error {
	// Open up the repo
	r, err := git.PlainOpen(p.Path)
	if err != nil {
		return err
	}

	// Get a reference to the work tree
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	if len(status) != 0 {
		return errors.New("Godot requires a clean workdir, stash/reset any manual changes")
	}

	return nil
}

func (p PureGoRepo) Push() error {
	p.ensureClean()

	// Open up the repo
	r, err := git.PlainOpen(p.Path)
	if err != nil {
		return err
	}

	// Get a reference to the work tree
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	for path := range status {
		_, err = w.Add(path)
		if err != nil {
			return err
		}
	}

	_, err = w.Commit("[godot]: update configuration", &git.CommitOptions{})
	if err != nil {
		return err
	}

	return r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: p.User,
			Password: os.Getenv("GITHUB_PAT"),
		},
	})
}

func (p PureGoRepo) Pull() error {
	p.ensureClean()

	// Open up the repo
	r, err := git.PlainOpen(p.Path)
	if err != nil {
		return err
	}

	// Get a reference to the work tree
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Execute the pull
	err = w.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: p.User,
			Password: os.Getenv("GITHUB_PAT"),
		},
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}
