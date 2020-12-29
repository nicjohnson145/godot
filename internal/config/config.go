package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nicjohnson145/godot/internal/file"
	"github.com/tidwall/gjson"
)

type Config struct {
	Target       string `json:"target"`
	DotfilesRoot string `json:"dotfiles_root"`
	// This will be pulled from the config that lives inside the dotfiles repo
	Files []file.File
}

func NewConfig(getter file.HomeDirGetter) *Config {
	c := parseUserConfig(getter)
	c.setRelevantFiles()
	return c
}

func parseUserConfig(getter file.HomeDirGetter) *Config {
	home, err := getter.GetHomeDir()
	if err != nil {
		panic("Cannot get home dir")
	}
	bytes, err := ioutil.ReadFile(filepath.Join(home, ".config", "godot", "config.json"))
	if err != nil {
		panic(fmt.Errorf("Error reading build target, %v", err))
	}
	contents := string(bytes)

	if !gjson.Valid(contents) {
		panic("invalid json")
	}

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
		Target:       target.String(),
		DotfilesRoot: dotfilesRoot,
	}
}

func (c *Config) setRelevantFiles() {
	configPath := filepath.Join(c.DotfilesRoot, "config.json")
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// TODO: touch the file?
			return
		} else {
			panic(fmt.Errorf("error reading repo conf, %v", err))
		}
	}

	content := string(bytes)

	allVal := gjson.Get(content, "all_files")
	if !allVal.Exists() {
		panic(fmt.Sprintf("Malformed repo conf, missing all files key"))
	}

	allFiles := make(map[string]file.File)
	allVal.ForEach(func(key, value gjson.Result) bool {
		allFiles[key.String()] = file.File{
			DestinationPath: value.String(),
			TemplatePath: filepath.Join(c.DotfilesRoot, "templates", key.String()),
		}
		return true // keep iterating
	})

	names := gjson.Get(content, fmt.Sprintf("renders.%v", c.Target))
	if !names.Exists() {
		panic(fmt.Sprintf("No file list found for target %q", c.Target))
	}

	var files []file.File
	names.ForEach(func(key, value gjson.Result) bool {
		file, ok := allFiles[value.String()]
		if !ok {
			panic(fmt.Sprintf("Invalid file key of %q for target %q", value.String(), c.Target))
		}
		files = append(files, file)
		return true // keep iterating
	})

	c.Files = files
}
