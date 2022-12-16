package lib

import (
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func EnsureDotfilesRepo(conf UserConfig) error {
	return getDotfilesRepo(conf)
}

func getDotfilesRepo(conf UserConfig) error {
	// If it's already cloned, open it and pull latest
	cloned, err := isRepoCloned(conf.CloneLocation)
	if err != nil {
		return err
	}
	if cloned {
		repo, err := openGitRepo(conf.CloneLocation)
		if err != nil {
			return fmt.Errorf("error opening repo: %w", err)
		}

		w, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("error getting dotfiles worktree: %v", err)
		}

		err = w.Pull(&git.PullOptions{
			Auth: &http.BasicAuth{
				Username: "my-cool-token",
				Password: conf.GithubPAT,
			},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("error pulling dotfiles repo: %v", err)
		}

		return nil
	}

	// Otherwise, clone it
	_, err = cloneGitRepo(
		conf.DotfilesURL,
		conf.CloneLocation,
		authFromConfig(conf),
	)
	if err != nil {
		return fmt.Errorf("error cloning dotfiles repo: %w", err)
	}
	return nil
}

//nolint:ireturn
func authFromConfig(conf UserConfig) http.AuthMethod {
	return &http.BasicAuth{
		Username: "my-cool-token",
		Password: conf.GithubPAT,
	}
}

func isRepoCloned(location string) (bool, error) {
	exists, err := pathExists(path.Join(location, ".git"))
	if err != nil {
		return false, fmt.Errorf("error checking existence of dotfiles repo: %v", err)
	}
	return exists, nil
}

func cloneGitRepo(url string, location string, auth http.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(location, false, &git.CloneOptions{
		Auth: auth,
		URL:  url,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning %v: %v", url, err)
	}
	return repo, nil
}

func openGitRepo(location string) (*git.Repository, error) {
	repo, err := git.PlainOpen(location)
	if err != nil {
		return nil, fmt.Errorf("error opening repo: %v", err)
	}
	return repo, nil
}

func ClonePublicRepo(url string, location string) (*git.Repository, error) {
	return cloneGitRepo(url, location, nil)
}

func ClonePrivateRepo(url string, location string, conf UserConfig) (*git.Repository, error) {
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

func fetchPrivateRepo(repo *git.Repository, conf UserConfig) error {
	return fetchRepo(repo, authFromConfig(conf))
}

func fetchPublicRepo(repo *git.Repository) error {
	return fetchRepo(repo, nil)
}

func fetchRepo(repo *git.Repository, auth http.AuthMethod) error {
	err := repo.Fetch(&git.FetchOptions{
		Auth: auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error fetching new commits: %v", err)
	}
	return nil
}

func ensureCommitCheckedOut(repo *git.Repository, ref Ref) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}

	err = w.Checkout(getCheckoutOptions(ref))
	if err != nil {
		return fmt.Errorf("error checking out commit %v: %v", ref.String(), err)
	}
	return nil
}

func getCheckoutOptions(r Ref) *git.CheckoutOptions {
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
