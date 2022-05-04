package lib

import (
	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
)

type GitRepo struct {
	URL      string `yaml:"url"`
	Location string `yaml:"location"`
	Private  bool   `yaml:"private"`
	Commit   string `yaml:"commit"`
}

func (g GitRepo) Execute(conf UserConfig) {
	log.Infof("Ensuring %v cloned", g.URL)

	var repo *git.Repository
	if !isRepoCloned(g.Location) {
		if g.Private {
			repo = ClonePrivateRepo(g.URL, g.Location, conf)
		} else {
			repo = ClonePublicRepo(g.URL, g.Location)
		}
	} else {
		repo = openGitRepo(g.Location)
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
