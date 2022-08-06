package lib

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

func EnsureDotfilesRepo(conf UserConfig) {
	getDotfilesRepo(conf)
}

func getDotfilesRepo(conf UserConfig) *git.Repository {
	// If it's already cloned, open it and pull latest
	if isRepoCloned(conf.CloneLocation) {
		repo := openGitRepo(conf.CloneLocation)

		w, err := repo.Worktree()
		if err != nil {
			log.Fatalf("Error getting dotfiles worktree: %v", err)
		}

		err = w.Pull(&git.PullOptions{
			Auth: &http.BasicAuth{
				Username: "my-cool-token",
				Password: conf.GithubPAT,
			},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			log.Fatalf("Error pulling dotfiles repo: %v", err)
		}

		return repo
	}

	// Otherwise, clone it
	return cloneGitRepo(
		conf.DotfilesURL,
		conf.CloneLocation,
		authFromConfig(conf),
	)
}

func authFromConfig(conf UserConfig) http.AuthMethod {
	return &http.BasicAuth{
		Username: "my-cool-token",
		Password: conf.GithubPAT,
	}
}

func isRepoCloned(location string) bool {
	exists, err := pathExists(path.Join(location, ".git"))
	if err != nil {
		log.Fatalf("Error checking existence of dotfiles repo: %v", err)
	}
	return exists
}

func cloneGitRepo(url string, location string, auth http.AuthMethod) *git.Repository {
	repo, err := git.PlainClone(location, false, &git.CloneOptions{
		Auth: auth,
		URL:  url,
	})
	if err != nil {
		log.Fatalf("Error cloning %v: %v", url, err)
	}
	return repo
}

func openGitRepo(location string) *git.Repository {
	repo, err := git.PlainOpen(location)
	if err != nil {
		log.Fatalf("Error opening repo: %v", err)
	}
	return repo
}

func ClonePublicRepo(url string, location string) *git.Repository {
	return cloneGitRepo(url, location, nil)
}

func ClonePrivateRepo(url string, location string, conf UserConfig) *git.Repository {
	return cloneGitRepo(
		url,
		location,
		authFromConfig(conf),
	)
}

func pathExists(loc string) (bool, error) {
	if _, err := os.Stat(loc); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func fetchPrivateRepo(repo *git.Repository, conf UserConfig) {
	fetchRepo(repo, authFromConfig(conf))
}

func fetchPublicRepo(repo *git.Repository) {
	fetchRepo(repo, nil)
}

func fetchRepo(repo *git.Repository, auth http.AuthMethod) {
	err := repo.Fetch(&git.FetchOptions{
		Auth: auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Fatalf("Error fetching new commits: %v", err)
	}
}

func ensureCommitCheckedOut(repo *git.Repository, ref Ref) {
	w, err := repo.Worktree()
	if err != nil {
		log.Fatalf("Error getting worktree: %v", err)
	}

	err = w.Checkout(getCheckoutOptions(ref))
	if err != nil {
		log.Fatalf("Error checking out commit %v: %v", ref.String(), err)
	}
}

func getCheckoutOptions(r Ref)  *git.CheckoutOptions {
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
