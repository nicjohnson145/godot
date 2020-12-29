package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/nicjohnson145/godot/internal/file"
	"github.com/tidwall/gjson"
)

type Config struct {
	Target       string `json:"target"`
	DotfilesRoot string `json:"dotfiles_root"`
	//// This will be pulled from the config that lives inside the dotfiles repo
	//Files        []file.File
}

func NewConfig(getter file.HomeDirGetter) *Config {
	home, err := getter.GetHomeDir()
	if err != nil {
		panic("Cannot get home dir")
	}
	bytes, err := ioutil.ReadFile(filepath.Join(home, ".config", "godot", "config.json"))
	if err != nil {
		panic(fmt.Errorf("Error reading build target, %v", err))
	}
	contents := string(bytes)

	target := gjson.Get(contents, "target")
	if !target.Exists() {
		panic("missing 'target' key in config")
	}

	root := gjson.Get(contents, "dotfiles_root")
	var dotfilesRoot string
	if !root.Exists() {
		dotfilesRoot = filepath.Join(home, "dotfiles")
	} else {
		dotfilesRoot = root.String()
	}

	return &Config{
		Target: target.String(),
		DotfilesRoot: dotfilesRoot,
	}
}

