package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type PathPair struct {
	DestPath   string `json:"dest_path"`
	SourcePath string `json:"src_path"`
	Group      string `json:"group"`
}

func (this *PathPair) addSourceRoot(root string) {
	this.SourcePath = filepath.Join(root, this.SourcePath)
}

func (this *PathPair) addDestRoot(root string) {
	this.DestPath = filepath.Join(root, this.DestPath)
}

type godotConfig struct {
	RepoPath     string     `json:"repo_path"`
	ManagedFiles []PathPair `json:"managed_files"`
	TargetName   string     `json:"target"`
}

func (this godotConfig) String() string {
	s := ""
	s += "GodotConfig {\n"
	s += fmt.Sprintf("  RepoPath: '%v',\n", this.RepoPath)
	s += fmt.Sprintf("  TargetName: '%v',\n", this.TargetName)
	s += fmt.Sprintf("  ManagedFiles: %+v,\n", this.ManagedFiles)
	s += "}"
	return s
}

func defaultConfig() godotConfig {
	usr, _ := user.Current()

	return godotConfig{
		RepoPath:     filepath.Join(usr.HomeDir, "dotfiles"),
		ManagedFiles: []PathPair{},
		TargetName:   "",
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
		fmt.Println("Error parsing config file")
		fmt.Println(err)
		os.Exit(1)
	}
	return userConf
}

func getGodotConfig() godotConfig {
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
