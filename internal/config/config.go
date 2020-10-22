package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"github.com/nicjohnson145/godot/internal/util"
)

type PathPair struct {
	DestPath   string `json:"dest_path"`
	SourcePath string `json:"src_path"`
	Group      string `json:"group"`
}

func (this *PathPair) AddSourceRoot(root string) {
	this.SourcePath = filepath.Join(root, this.SourcePath)
}

func (this *PathPair) AddDestRoot(root string) {
	this.DestPath = filepath.Join(root, this.DestPath)
}

type GodotConfig struct {
	RepoPath     string     `json:"repo_path"`
	ManagedFiles []PathPair `json:"managed_files"`
	TargetName   string     `json:"target"`
}

func (this GodotConfig) String() string {
	s := ""
	s += "GodotConfig {\n"
	s += fmt.Sprintf("  RepoPath: '%v',\n", this.RepoPath)
	s += fmt.Sprintf("  TargetName: '%v',\n", this.TargetName)
	s += fmt.Sprintf("  ManagedFiles: %+v,\n", this.ManagedFiles)
	s += "}"
	return s
}

func defaultConfig() GodotConfig {
	usr, _ := user.Current()

	return GodotConfig{
		RepoPath:     filepath.Join(usr.HomeDir, "dotfiles"),
		ManagedFiles: []PathPair{},
		TargetName:   "",
	}
}

func readConfig() GodotConfig {
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".config", "godot.json")
	if !util.IsFile(path) {
		return GodotConfig{}
	}

	data, _ := ioutil.ReadFile(path)
	var userConf GodotConfig
	err := json.Unmarshal(data, &userConf)
	util.Check(err)
	return userConf
}

func GetGodotConfig() GodotConfig {
	basic := defaultConfig()
	usr := readConfig()

	if usr.RepoPath != "" {
		basic.RepoPath = usr.RepoPath
	}
	if len(usr.ManagedFiles) != 0 {
		basic.ManagedFiles = usr.ManagedFiles
	}
	if usr.TargetName == "" {
		host, _ := os.Hostname()
		basic.TargetName = host
	} else {
		basic.TargetName = usr.TargetName
	}
	return basic
}
