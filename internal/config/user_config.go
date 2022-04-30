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
	BinaryDir  string `yaml:"binary-dir"`
	GithubUser string `yaml:"github-user"`
	GithubAuth string
}

func NewConfig() UserConfig {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting home directory: %v", err))
	}

	return NewConfigFromPath(path.Join(dir, ".config", "godot", "config.yaml"))
}

func NewConfigFromPath(confPath string) UserConfig {
	data, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading config path %v: %v", confPath, err))
	}

	var conf UserConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		log.Fatal(fmt.Sprintf("Error parsing user config: %v", err))
	}

	// Set the default binary directory
	if conf.BinaryDir == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(fmt.Sprintf("Error getting home directory: %v", err))
		}

		conf.BinaryDir = path.Join(dir, "bin")
	}

	// Now setup the github auth
	pat, ok := os.LookupEnv("GITHUB_PAT")
	if ok && conf.GithubUser != "" {
		conf.GithubAuth = lib.BasicAuth(conf.GithubUser, pat)
	}

	if !ok {
		log.Warn("GITHUB_PAT not set, requests to the github API might be rate limited")
	}

	if conf.GithubUser == "" {
		log.Warn("github-user not set, requests to the github API might be rate limited")
	}

	return conf
}
