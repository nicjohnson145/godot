package lib

import (
	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
)

var _ Executor = (*GitRepo)(nil)

type GitRepo struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Location string `yaml:"location"`
	Private  bool   `yaml:"private"`
	Ref      Ref    `yaml:"ref"`
}

type Ref struct {
	Commit string `yaml:"commit"`
	Tag    string `yaml:"tag"`
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

func (g *GitRepo) Type() ExecutorType {
	return ExecutorTypeGitRepos
}

func (g *GitRepo) Execute(conf UserConfig, opts SyncOpts, _ TargetConfig) {
	log.Infof("Ensuring %v cloned", g.URL)

	location := replaceTilde(g.Location, conf.HomeDir)

	var repo *git.Repository
	if !isRepoCloned(location) {
		if g.Private {
			repo = ClonePrivateRepo(g.URL, location, conf)
		} else {
			repo = ClonePublicRepo(g.URL, location)
		}
	} else {
		repo = openGitRepo(location)
	}

	if g.Private {
		fetchPrivateRepo(repo, conf)
	} else {
		fetchPublicRepo(repo)
	}

	if !g.Ref.IsZero() {
		log.Infof("Ensuring %v at commit %v", g.URL, g.Ref.String())
		ensureCommitCheckedOut(repo, g.Ref)
	}
}
