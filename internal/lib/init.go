package lib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/nicjohnson145/godot/internal/help"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func Init(opts InitOpts) error {
	home, err := opts.HomeDirGetter.GetHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".config", "godot", "config.json")
	if checkFileExists(configPath) {
		overwrite, err := promptConfirm("Configuration already exists, overwrite?")
		if err != nil {
			return err
		}
		if !overwrite {
			return nil
		}
	}

	username, err := promptNonEmpty("Github Username", "")
	if err != nil {
		return err
	}
	dotfiles_location, err := promptNonEmpty("Dotfiles Location", filepath.Join(home, "dotfiles"))
	if err != nil {
		return err
	}
	target , err := promptNonEmpty("Target Name", "")
	if err != nil {
		return err
	}
	additionalMgrs, err := promptPackageManagers()
	if err != nil {
		return err
	}

	config := usrConfig{
		Target: target,
		DotfilesRoot: dotfiles_location,
		PackageManagers: additionalMgrs,
		GithubUser: username,
	}
	err = dumpJson(config, configPath)
	if err != nil {
		return err
	}

	if checkFileExists(dotfiles_location) {
		overwrite, err := promptConfirm("Dotfiles location already exists, delete and re-clone?")
		if err != nil {
			return err
		}

		if overwrite {
			err := os.RemoveAll(dotfiles_location)
			if err != nil {
				return err
			}
			err = cloneRepo(dotfiles_location, username)
			if err != nil {
				return err
			}
		}
	} else {
		err := cloneRepo(dotfiles_location, username)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	panic(fmt.Sprintf("Unable to determine existance of %v", path))
}

func promptNonEmpty(label string, _default string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if input == "" {
				return errors.New("Must supply a value")
			}
			return nil
		},
		Default: _default,
	}
	result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		return "", errors.New("Cancelling...")
	}
	return result, err
}


func promptPackageManagers() ([]string, error) {
	prompt := promptui.Prompt{
		Label: "Additional package managers (enter for none)",
		Validate: func(input string) error {
			if input == "" {
				return nil
			}

			arr := strings.Split(input, ",")
			for _, item := range arr {
				if !help.StringInSlice(item, ValidManagers) {
					return errors.New(fmt.Sprintf(
						"Invalid package manager %v. Valid managers are %v",
						item,
						strings.Join(ValidManagers, ", "),
					))
				}
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		return []string{}, errors.New("Cancelling...")
	}

	deduped := []string{GIT}
	if result != "" {
		arr := strings.Split(result, ",")
		for _, item := range arr {
			if !help.StringInSlice(item, deduped) {
				deduped = append(deduped, item)
			}
		}
	}
	return deduped, nil
}

func promptConfirm(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if !help.StringInSlice(input, []string{"y", "N"}) {
				return errors.New("Must supply 'y' or 'N'")
			}
			return nil
		},
	}
	result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		return false, errors.New("Cancelling...")
	}
	return result == "y", err
}

func cloneRepo(dotfiles string, username string) error {
	url, err := promptNonEmpty(
		"Dotfiles github URL",
		fmt.Sprintf("https://github.com/%v/dotfiles", username),
	)
	if err != nil {
		return err
	}

	_, err = git.PlainClone(dotfiles, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: os.Getenv("GITHUB_PAT"),
		},
		URL: url,
	})

	return err
}
