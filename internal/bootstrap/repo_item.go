package bootstrap

import (
	"fmt"
	"github.com/nicjohnson145/godot/internal/help"
	"os"
	"path/filepath"
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
	_, _, err := help.RunCmd("git", "clone", i.Name, i.Location)
	return err
}
