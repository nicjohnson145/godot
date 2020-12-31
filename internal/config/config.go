package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

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
	home         string      // Users home directory
}

func NewConfig(getter util.HomeDirGetter) *Config {
	home, err := getter.GetHomeDir()
	if err != nil {
		panic("Cannot get home dir")
	}
	c := &Config{
		home: home,
	}
	c.parseUserConfig()
	c.readRepoConfig()
	return c
}

func (c *Config) parseUserConfig() {
	bytes, err := ioutil.ReadFile(filepath.Join(c.home, ".config", "godot", "config.json"))
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
	c.Target = target.String()

	root := gjson.Get(contents, "dotfiles_root")
	var dotfilesRoot string
	if !root.Exists() {
		dotfilesRoot = filepath.Join(c.home, "dotfiles")
	} else {
		dotfilesRoot = root.String()
	}
	c.DotfilesRoot = dotfilesRoot
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

func (c *Config) getAllFiles() map[string]file.File {
	allFiles := make(map[string]file.File)

	allVal := gjson.Get(c.content, "all_files")
	if !allVal.Exists() {
		return allFiles
	}

	allVal.ForEach(func(key, value gjson.Result) bool {
		f := file.File{
			DestinationPath: util.ReplacePrefix(value.String(), "~/", c.home),
			TemplatePath:    filepath.Join(c.DotfilesRoot, "templates", key.String()),
		}
		allFiles[key.String()] = f
		return true // keep iterating
	})

	return allFiles
}

func (c *Config) Write() error {
	return ioutil.WriteFile(c.repoConfig, []byte(c.content), 0744)
}

func (c *Config) ManageFile(destination string) error {
	templateName := util.ToTemplateName(destination)
	if c.IsValidFile(templateName) {
		return errors.New(fmt.Sprintf("template name %q already exists", templateName))
	}
	newDest := util.ReplacePrefix(destination, c.home, "~")
	return c.AddFile(templateName, newDest)
}

func (c *Config) AddFile(template string, destination string) error {
	if c.IsValidFile(template) {
		return errors.New(fmt.Sprintf("template name %q already exists", template))
	}
	value, err := sjson.Set(c.content, fmt.Sprintf("all_files.%v", template), destination)
	if err != nil {
		err = fmt.Errorf("error adding file, %v", err)
		return err
	}
	c.content = value
	return nil
}

func (c *Config) AddToTarget(target string, name string) error {
	if !c.IsValidFile(name) {
		return errors.New(fmt.Sprintf("unknown file name of %q", name))
	}
	value, err := sjson.Set(c.content, fmt.Sprintf("renders.%v.-1", target), name)
	if err != nil {
		err = fmt.Errorf("error adding %q to %q, %v", name, target, err)
		return err
	}
	c.content = value
	return nil
}

func (c *Config) IsValidFile(name string) bool {
	value := gjson.Get(c.content, fmt.Sprintf("all_files.%v", name))
	return value.Exists()
}

func (c *Config) GetTargetFiles() []file.File {
	var files []file.File

	allFiles := c.getAllFiles()
	names := gjson.Get(c.content, fmt.Sprintf("renders.%v", c.Target))
	if !names.Exists() {
		return files
	}

	names.ForEach(func(key, value gjson.Result) bool {
		file, ok := allFiles[value.String()]
		if !ok {
			panic(fmt.Sprintf("Invalid file key of %q for target %q", value.String(), c.Target))
		}
		files = append(files, file)
		return true // keep iterating
	})

	return files
}

func (c *Config) ListAllFiles(w io.Writer) {
	
}
