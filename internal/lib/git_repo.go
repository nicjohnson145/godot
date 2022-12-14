package lib

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
)

var _ Executor = (*GitRepo)(nil)

type GitRepo struct {
	Name     string `yaml:"-"`
	URL      string `yaml:"url" mapstructure:"url"`
	Location string `yaml:"location" mapstructure:"location"`
	Private  bool   `yaml:"private" mapstructure:"private"`
	Ref      Ref    `yaml:"ref" mapstructure:"ref"`
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

func (g *GitRepo) Execute(conf UserConfig, opts SyncOpts, _ GodotConfig) error {
	log.Infof("Ensuring %v cloned", g.URL)

	location := replaceTilde(g.Location, conf.HomeDir)

	var repo *git.Repository
	var err error

	cloned, err := isRepoCloned(location)
	if err != nil {
		return err
	}

	if !cloned {
		if g.Private {
			repo, err = ClonePrivateRepo(g.URL, location, conf)
		} else {
			repo, err = ClonePublicRepo(g.URL, location)
		}
	} else {
		repo, err = openGitRepo(location)
	}
	if err != nil {
		return fmt.Errorf("error ensuring repo cloned: %w", err)
	}

	if g.Private {
		if err := fetchPrivateRepo(repo, conf); err != nil {
			return fmt.Errorf("error fetching private repo: %w", err)
		}
	} else {
		if err := fetchPublicRepo(repo); err != nil {
			return fmt.Errorf("error fetching public repo: %w", err)
		}
	}

	if !g.Ref.IsZero() {
		log.Infof("Ensuring %v at commit %v", g.URL, g.Ref.String())
		if err := ensureCommitCheckedOut(repo, g.Ref); err != nil {
			return fmt.Errorf("error ensuring commit checked out: %w", err)
		}
	}

	return nil
}
