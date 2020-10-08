package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type godotConfig struct {
	RepoPath string
}

func (this godotConfig) String() string {
	s := ""
	s += "GodotConfig {\n"
	s += fmt.Sprintf("  RepoPath: '%v',\n", this.RepoPath)
	s += "}"
	return s
}

func defaultConfig() godotConfig {
	usr, _ := user.Current()

	return godotConfig{
		RepoPath: filepath.Join(usr.HomeDir, "dotfiles"),
	}
}

func readConfig() godotConfig {
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".config", "godot.json")
	if !isFile(path) {
		return godotConfig{}
	}

	data, _ := ioutil.ReadFile(path)
	var userConf godotConfig
	err := json.Unmarshal(data, &userConf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return userConf
}

func GetGodotConfig() godotConfig {
	basic := defaultConfig()
	usr := readConfig()
	if usr.RepoPath != "" {
		basic.RepoPath = usr.RepoPath
	}
	return basic
}
