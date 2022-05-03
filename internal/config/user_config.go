package config

import (
	"fmt"
	"github.com/nicjohnson145/godot/internal/lib"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"path"
)

type UserConfig struct {
	BinaryDir     string `yaml:"binary-dir"`
	GithubUser    string `yaml:"github-user"`
	Target        string `yaml:"target"`
	DotfilesURL   string `yaml:"dotfiles-url"`
	CloneLocation string `yaml:"clone-location"`
	GithubPAT     string
	GithubAuth    string
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

	// Set the default binary directory
	if conf.BinaryDir == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Error getting home directory: %v", err)
		}

		conf.BinaryDir = path.Join(dir, "bin")
	}

	// Now setup the github auth
	pat, ok := os.LookupEnv("GITHUB_PAT")
	if ok {
		conf.GithubPAT = pat
	}
	if ok && conf.GithubUser != "" {
		conf.GithubAuth = lib.BasicAuth(conf.GithubUser, pat)
	}

	if !ok {
		log.Warn("GITHUB_PAT not set, requests to the github API might be rate limited")
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
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Error getting home directory: %v", err)
		}
		conf.CloneLocation = path.Join(dir, ".config", "godot", "dotfiles")
	}

	return conf
}
