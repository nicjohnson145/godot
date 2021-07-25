package lib

import (
	"errors"
)

const (
	APT             = "apt"
	BREW            = "brew"
	GIT             = "git"
	CURRENT         = "<CURRENT>"
	TarGz           = "tar_gz"
	BinaryOnly      = "binary_only"
	DefaultLocation = "~/bin"
	AllTarget       = "ALL"
)

var ValidManagers = []string{APT, BREW, GIT}

var NotFoundError = errors.New("not found")

type Config struct {
	Target          string     // Name of the current target
	DotfilesRoot    string     // Root of the dotfiles repo
	TemplateRoot    string     // Template directory inside of dotfiles repo
	content         RepoConfig // The raw json content
	repoConfig      string     // Path to repo config, we'll need to rewrite it often
	Home            string     // Users home directory
	PackageManagers []string   // Available package managers, in order of preference
	githubUser      string     // User to authenticate with when downloading github releases through API
}

type StringMap map[string]string

type BootstrapItem struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
}

type Host struct {
	Files         []string             `json:"files"`
	Bootstraps    []string             `json:"bootstraps"`
	GithubRelease []GithubReleaseUsage `json:"github_releases"`
}

type GithubReleaseUsage struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	TrackUpdates bool   `json:"track_updates"`
}

type GithubReleaseConfiguration struct {
	name     string            `json:"-"`
	RepoName string            `json:"repo_name"`
	Patterns map[string]string `json:"patterns"`
	Download Download          `json:"download"`
}

type Download struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

type HomeDirGetter interface {
	GetHomeDir() (string, error)
}

type usrConfig struct {
	Target          string   `json:"target"`
	DotfilesRoot    string   `json:"dotfiles_root"`
	PackageManagers []string `json:"package_managers"`
	GithubUser      string   `json:"github_user"`
}

type Item interface {
	Check() (bool, error)
	Install() error
}

type SyncOpts struct {
	Force       bool
	NoBootstrap bool
}

type EditOpts struct {
	NoSync bool
}
