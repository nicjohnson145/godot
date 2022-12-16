package lib

import (
	"fmt"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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

type vaultFunc func(*UserConfig) error

type ConfigOverrides struct {
	IgnoreVault bool
}

func homeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}
	return dir, nil
}

func NewConfig() (UserConfig, error) {
	home, err := homeDir()
	if err != nil {
		return UserConfig{}, err
	}
	return NewConfigFromPath(
		path.Join(home, ".config", "godot", "config.yaml"),
		setVaultClient,
		ConfigOverrides{},
	)
}

func NewOverrideableConfig(overrides ConfigOverrides) (UserConfig, error) {
	home, err := homeDir()
	if err != nil {
		return UserConfig{}, err
	}

	return NewConfigFromPath(
		path.Join(home, ".config", "godot", "config.yaml"),
		setVaultClient,
		overrides,
	)
}

//nolint:gocognit,gocyclo
func NewConfigFromPath(confPath string, setClient vaultFunc, overrides ConfigOverrides) (UserConfig, error) {
	data, err := os.ReadFile(confPath)
	if err != nil {
		return UserConfig{}, fmt.Errorf("error reading config path %v: %v", confPath, err)
	}

	var conf UserConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return UserConfig{}, fmt.Errorf("error parsing user config: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return UserConfig{}, fmt.Errorf("error getting home directory: %v", err)
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
			return UserConfig{}, fmt.Errorf("error getting hostname for default target: %v", err)
		}
		conf.Target = name
	}

	// Default the dotfiles url
	if conf.DotfilesURL == "" {
		if conf.GithubUser == "" {
			return UserConfig{}, fmt.Errorf("both dotfiles-url and github-user are not set, dotfiles url cannot be inferred")
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
			return UserConfig{}, fmt.Errorf("unsupported packaged manager of %v\n", conf.PackageManager)
		}
	}

	// Initialize the vault client (if requested), since we may get the github pat from vault and
	// not the environment
	if !overrides.IgnoreVault {
		if err := setClient(&conf); err != nil {
			return UserConfig{}, fmt.Errorf("error setting vault client: %w", err)
		}
	}

	// Now setup the github auth
	if conf.VaultConfig.GithubPatFromVault && !overrides.IgnoreVault {
		if !conf.VaultConfig.Client.Initialized() {
			return UserConfig{}, fmt.Errorf("configured to read github PAT from vault, but vault client not properly initialized")
		}
		pat, err := conf.VaultConfig.Client.ReadKey(
			conf.VaultConfig.GithubPatConfig.Path,
			conf.VaultConfig.GithubPatConfig.Key,
		)
		if err != nil {
			return UserConfig{}, fmt.Errorf("error getting PAT from vault: %w", err)
		}
		conf.GithubPAT = pat
	} else {
		pat, ok := os.LookupEnv("GITHUB_PAT")
		if ok {
			conf.GithubPAT = pat
		}
		if !ok {
			log.Warn("GITHUB_PAT not set, requests to the github API might be rate limited")
		}
	}

	if conf.GithubPAT != "" && conf.GithubUser != "" {
		conf.GithubAuth = BasicAuth(conf.GithubUser, conf.GithubPAT)
	}

	return conf, nil
}
