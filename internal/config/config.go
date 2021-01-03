package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/nicjohnson145/godot/internal/file"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

type Config struct {
	Target       string `json:"target"`
	DotfilesRoot string `json:"dotfiles_root"`
	content      string // The raw json content
	repoConfig   string // Path to repo config, we'll need to rewrite it often
	Home         string // Users home directory
}

func NewConfig(getter util.HomeDirGetter) *Config {
	home, err := getter.GetHomeDir()
	if err != nil {
		panic("Cannot get home dir")
	}
	c := &Config{
		Home: home,
	}
	c.parseUserConfig()
	c.readRepoConfig()
	return c
}

func (c *Config) parseUserConfig() {
	bytes, err := ioutil.ReadFile(filepath.Join(c.Home, ".config", "godot", "config.json"))
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
		dotfilesRoot = filepath.Join(c.Home, "dotfiles")
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
			DestinationPath: util.ReplacePrefix(value.String(), "~/", c.Home),
			TemplatePath:    filepath.Join(c.DotfilesRoot, "templates", key.String()),
		}
		allFiles[key.String()] = f
		return true // keep iterating
	})

	return allFiles
}

func (c *Config) Write() error {
	pretty := pretty.PrettyOptions([]byte(c.content), &pretty.Options{Indent: "    "})
	return ioutil.WriteFile(c.repoConfig, pretty, 0644)
}

func (c *Config) ManageFile(destination string) (string, error) {
	templateName := util.ToTemplateName(destination)
	if c.IsValidFile(templateName) {
		return "", errors.New(fmt.Sprintf("template name %q already exists", templateName))
	}
	return c.AddFile(templateName, destination)
}

func (c *Config) AddFile(template string, destination string) (string, error) {
	if c.IsValidFile(template) {
		return "", errors.New(fmt.Sprintf("template name %q already exists", template))
	}
	newDest := util.ReplacePrefix(destination, c.Home, "~")
	value, err := sjson.Set(c.content, fmt.Sprintf("all_files.%v", template), newDest)
	if err != nil {
		err = fmt.Errorf("error adding file, %v", err)
		return "", err
	}
	c.content = value
	return template, nil
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
	for _, value := range c.getFilesByTarget(c.Target) {
		files = append(files, value)
	}
	return files
}

func (c *Config) getFilesByTarget(target string) map[string]file.File {
	files := make(map[string]file.File)

	allFiles := c.getAllFiles()
	names := gjson.Get(c.content, fmt.Sprintf("renders.%v", target))
	if !names.Exists() {
		return files
	}

	names.ForEach(func(key, value gjson.Result) bool {
		file, ok := allFiles[value.String()]
		if !ok {
			panic(fmt.Sprintf("Invalid file key of %q for target %q", value.String(), target))
		}
		files[value.String()] = file
		return true // keep iterating
	})

	return files
}

func (c *Config) ListAllFiles(w io.Writer) {
	allFiles := c.getAllFiles()
	c.writeFileMap(w, allFiles)
}

func (c *Config) writeFileMap(w io.Writer, files map[string]file.File) {
	keys := make([]string, 0)
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	tw := tabwriter.NewWriter(w, 0, 1, 0, ' ', tabwriter.AlignRight)
	for _, key := range keys {
		fl := files[key]
		subbedDest := util.ReplacePrefix(fl.DestinationPath, c.Home, "~")
		_, err := fmt.Fprintln(tw, fmt.Sprintf("%v\t => %v", key, subbedDest))
		if err != nil {
			panic(fmt.Sprintf("error listing files, %v", err))
		}
	}
	err := tw.Flush()
	if err != nil {
		panic(err)
	}
}

func (c *Config) ListTargetFiles(target string, w io.Writer) {
	files := c.getFilesByTarget(target)
	c.writeFileMap(w, files)
}
