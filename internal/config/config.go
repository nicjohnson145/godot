package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicjohnson145/godot/internal/file"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Config struct {
	Target       string      `json:"target"`
	DotfilesRoot string      `json:"dotfiles_root"`
	content      string      // The raw json content
	repoConfig   string      // Path to repo config, we'll need to rewrite it often
	Files        []file.File // This will be pulled from the config that lives inside the dotfiles repo
}

func NewConfig(getter util.HomeDirGetter) *Config {
	home, err := getter.GetHomeDir()
	if err != nil {
		panic("Cannot get home dir")
	}
	c := parseUserConfig(home)
	c.readRepoConfig()
	c.setRelevantFiles(home)
	return c
}

func parseUserConfig(home string) *Config {
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

func (c *Config) readRepoConfig() {
	configPath := filepath.Join(c.DotfilesRoot, "config.json")
	c.repoConfig = configPath
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
	c.content = content
}

func (c *Config) setRelevantFiles(home string) {
	if c.content == "" {
		// TODO: what do if no user config
		return
	}
	allVal := gjson.Get(c.content, "all_files")
	if !allVal.Exists() {
		panic(fmt.Sprintf("Malformed repo conf, missing all files key"))
	}

	allFiles := make(map[string]file.File)
	allVal.ForEach(func(key, value gjson.Result) bool {
		f := file.File{
			DestinationPath: value.String(),
			TemplatePath:    filepath.Join(c.DotfilesRoot, "templates", key.String()),
		}
		substituteTilde(&f, home)
		allFiles[key.String()] = f
		return true // keep iterating
	})

	names := gjson.Get(c.content, fmt.Sprintf("renders.%v", c.Target))
	if !names.Exists() {
		return
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

func (c *Config) Write() error {
	return ioutil.WriteFile(c.repoConfig, []byte(c.content), 0744)
}

func (c *Config) AddFile(template string, destination string) error {
	value, err := sjson.Set(c.content, fmt.Sprintf("all_files.%v", template), destination)
	if err != nil {
		err = fmt.Errorf("error adding file, %v", err)
		return err
	}
	c.content = value
	return nil
}

func substituteTilde(f *file.File, home string) {
	if strings.HasPrefix(f.DestinationPath, "~/") {
		f.DestinationPath = filepath.Join(home, f.DestinationPath[2:])
	}
}
