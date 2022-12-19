package lib

import (
	"fmt"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var _ Executor = (*GitRepo)(nil)

type GitRepo struct {
	Name        string `yaml:"-"`
	URL         string `yaml:"url" mapstructure:"url"`
	Location    string `yaml:"location" mapstructure:"location"`
	Private     bool   `yaml:"private" mapstructure:"private"`
	TrackLatest bool   `yaml:"track-latest" mapstructure:"track-latest"`
	Ref         Ref    `yaml:"ref" mapstructure:"ref"`
}

type Ref struct {
	Commit string `yaml:"commit" mapstructure:"commit"`
	Tag    string `yaml:"tag" mapstructure:"tag"`
}

func (r *Ref) IsZero() bool {
	return r.Commit == "" && r.Tag == ""
}

func (r *Ref) String() string {
	if r.Commit != "" {
		return r.Commit
	} else {
		return r.Tag
	}
}

func (g *GitRepo) GetName() string {
	return g.Name
}

func (g *GitRepo) SetName(n string) {
	g.Name = n
}

func (g *GitRepo) Type() ExecutorType {
	return ExecutorTypeGitRepo
}

func (g *GitRepo) Validate() error {
	var errs *multierror.Error

	if g.URL == "" {
		errs = multierror.Append(errs, fmt.Errorf("url is required"))
	}

	if g.Location == "" {
		errs = multierror.Append(errs, fmt.Errorf("location is required"))
	}

	if g.TrackLatest && !g.Ref.IsZero() {
		errs = multierror.Append(errs, fmt.Errorf("cannot specify a ref & track-latest"))
	}

	return errs.ErrorOrNil()
}

func (g *GitRepo) location(conf UserConfig) string {
	return replaceTilde(g.Location, conf.HomeDir)
}

func (g *GitRepo) Execute(conf UserConfig, _ SyncOpts, _ GodotConfig) error {
	log.Infof("Ensuring %v cloned", g.URL)

	var repo *git.Repository
	var err error

	// Check if it's already cloned
	cloned, err := g.isRepoCloned(conf)
	if err != nil {
		return err
	}

	// Either clone it or open it
	if !cloned {
		repo, err = g.cloneRepo(conf)
	} else {
		repo, err = g.openRepo(conf)
	}
	if err != nil {
		return fmt.Errorf("error ensuring repo cloned: %w", err)
	}

	// Either pull the latest commits, or ensure that that requested commit is checked out
	if g.TrackLatest {
		if err := g.pullRepo(repo, conf); err != nil {
			return fmt.Errorf("error pulling latest: %w", err)
		}
		return nil
	} else {
		if !g.Ref.IsZero() {
			// Fetch any new commits
			if err := g.fetchRepo(repo, conf); err != nil {
				return fmt.Errorf("error fetching new commits: %w", err)
			}
			log.Infof("Ensuring %v at commit %v", g.URL, g.Ref.String())
			if err := g.ensureCommitCheckedOut(repo, g.Ref); err != nil {
				return fmt.Errorf("error ensuring commit checked out: %w", err)
			}
		}
	}

	return nil
}

func (g *GitRepo) isRepoCloned(conf UserConfig) (bool, error) {
	exists, err := pathExists(path.Join(g.location(conf), ".git"))
	if err != nil {
		return false, fmt.Errorf("error checking existence of dotfiles repo: %v", err)
	}
	return exists, nil
}

func (g *GitRepo) authFromConfig(conf UserConfig) *http.BasicAuth {
	if g.Private {
		return &http.BasicAuth{
			Username: "my-cool-token",
			Password: conf.GithubPAT,
		}
	}
	return nil
}

func (g *GitRepo) cloneRepo(conf UserConfig) (*git.Repository, error) {
	repo, err := git.PlainClone(
		g.location(conf),
		false,
		&git.CloneOptions{
			Auth: g.authFromConfig(conf),
			URL: g.URL,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error cloning %v: %v", g.URL, err)
	}
	return repo, nil
}

func (g *GitRepo) openRepo(conf UserConfig) (*git.Repository, error) {
	repo, err := git.PlainOpen(g.location(conf))
	if err != nil {
		return nil, fmt.Errorf("error opening repo: %v", err)
	}
	return repo, nil
}

func (g *GitRepo) fetchRepo(repo *git.Repository, conf UserConfig) error {
	err := repo.Fetch(&git.FetchOptions{
		Auth: g.authFromConfig(conf),
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error fetching new commits: %v", err)
	}
	return nil
}

func (g *GitRepo) pullRepo(repo *git.Repository, conf UserConfig) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}

	err = w.Pull(&git.PullOptions{
		Auth: g.authFromConfig(conf),
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error pulling repo: %v", err)
	}

	return nil
}

func (g *GitRepo) ensureCommitCheckedOut(repo *git.Repository, ref Ref) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}

	err = w.Checkout(g.getCheckoutOptions(ref))
	if err != nil {
		return fmt.Errorf("error checking out commit %v: %v", ref.String(), err)
	}
	return nil
}

func (g *GitRepo) getCheckoutOptions(r Ref) *git.CheckoutOptions {
	if r.Commit != "" {
		return &git.CheckoutOptions{
			Hash: plumbing.NewHash(r.Commit),
		}
	} else {
		return &git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/tags/" + r.Tag),
		}
	}
}
