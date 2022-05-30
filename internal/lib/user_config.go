package lib

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"runtime"
)

type VaultPatConfig struct {
	Path string `yaml:"path"`
	Key  string `yaml:"key"`
}

type VaultConfig struct {
	Address            string         `yaml:"address"`
	TokenPath          string         `yaml:"token-path"`
	GithubPatFromVault bool           `yaml:"pat-from-vault"`
	GithubPatConfig    VaultPatConfig `yaml:"github-pat-config"`
	Client             VaultClient
}

type UserConfig struct {
	BinaryDir      string      `yaml:"binary-dir"`
	GithubUser     string      `yaml:"github-user"`
	Target         string      `yaml:"target"`
	DotfilesURL    string      `yaml:"dotfiles-url"`
	CloneLocation  string      `yaml:"clone-location"`
	BuildLocation  string      `yaml:"build-location"`
	PackageManager string      `yaml:"package-manager"`
	VaultConfig    VaultConfig `yaml:"vault-config"`
	GithubPAT      string
	GithubAuth     string
	HomeDir        string
}

func NewConfig() UserConfig {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	return NewConfigFromPath(path.Join(dir, ".config", "godot", "config.yaml"))
}

func NewConfigFromPath(confPath string) UserConfig {
	data, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatalf("Error reading config path %v: %v", confPath, err)
	}

	var conf UserConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatalf("Error parsing user config: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	conf.HomeDir = home

	// Set the default binary directory
	if conf.BinaryDir == "" {
		conf.BinaryDir = path.Join(home, "bin")
	} else {
		conf.BinaryDir = replaceTilde(conf.BinaryDir, home)
	}

	if conf.GithubUser == "" {
		log.Warn("github-user not set, requests to the github API might be rate limited")
	}

	// Default the target to the hostname
	if conf.Target == "" {
		name, err := os.Hostname()
		if err != nil {
			log.Fatalf("Error getting hostname for default target: %v", err)
		}
		conf.Target = name
	}

	// Default the dotfiles url
	if conf.DotfilesURL == "" {
		if conf.GithubUser == "" {
			log.Fatal("both dotfiles-url and github-user are not set, dotfiles url cannot be inferred")
		}
		conf.DotfilesURL = fmt.Sprintf("https://github.com/%v/dotfiles", conf.GithubUser)
	}

	// Default the clone location
	if conf.CloneLocation == "" {
		conf.CloneLocation = path.Join(home, ".config", "godot", "dotfiles")
	}

	// Default the build location
	if conf.BuildLocation == "" {
		conf.BuildLocation = path.Join(home, ".config", "godot", "rendered")
	}

	// Default and validate the package manager
	if conf.PackageManager == "" {
		switch runtime.GOOS {
		case "linux":
			conf.PackageManager = PackageManagerApt
		case "darwin":
			conf.PackageManager = PackageManagerBrew
		}
	} else {
		if !isValidPackageManager(conf.PackageManager) {
			log.Fatalf("Unsupported packaged manager of %v\n", conf.PackageManager)
		}
	}

	// Initialize the vault client (if requested), since we may get the github pat from vault and
	// not the environment
	setVaultClient(&conf)

	// Now setup the github auth
	if conf.VaultConfig.GithubPatFromVault {
		if !conf.VaultConfig.Client.Initialized() {
			log.Fatalf("Configured to read github PAT from vault, but vault client not properly initialized")
		}
		conf.GithubPAT = conf.VaultConfig.Client.ReadKeyOrDie(
			conf.VaultConfig.GithubPatConfig.Path,
			conf.VaultConfig.GithubPatConfig.Key,
		)
	} else {
		pat, ok := os.LookupEnv("GITHUB_PAT")
		if ok {
			conf.GithubPAT = pat
		}
		if ok && conf.GithubUser != "" {
			conf.GithubAuth = BasicAuth(conf.GithubUser, pat)
		}

		if !ok {
			log.Warn("GITHUB_PAT not set, requests to the github API might be rate limited")
		}
	}

	return conf
}
