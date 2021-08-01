package lib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type repoItem struct {
	Name     string
	Location string
}

func NewRepoItem(name string, location string) repoItem {
	return repoItem{
		Name:     name,
		Location: location,
	}
}

func (i repoItem) Check() (bool, error) {
	// Does the base location exist?
	baseExists, err := i.dirExists(i.Location)
	if err != nil {
		return false, err
	}
	if !baseExists {
		return false, nil
	}

	// The base directory exists, now does it have a .git in it?
	gitExists, err := i.dirExists(filepath.Join(i.Location, ".git"))
	if err != nil {
		return false, err
	}

	if gitExists {
		return true, nil
	} else {
		return false, fmt.Errorf("%q exists but is not a git checkout", i.Location)
	}
}

func (i repoItem) dirExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (i repoItem) Install() error {
	_, err := git.PlainClone(i.Location, false, &git.CloneOptions{
		URL: i.Name,
	})
	return err
}
