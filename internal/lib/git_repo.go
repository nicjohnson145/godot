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
	Commit   string `yaml:"commit"`
}

func (g *GitRepo) GetName() string {
	return g.Name
}

func (g *GitRepo) Execute(conf UserConfig, opts SyncOpts) {
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

	if g.Commit != "" {
		log.Infof("Ensuring %v at commit %v", g.URL, g.Commit)
		if g.Private {
			ensurePrivateRepoCommitCheckedOut(repo, g.Commit, conf)
		} else {
			ensurePublicRepoCommitCheckedOut(repo, g.Commit)
		}
	}
}
