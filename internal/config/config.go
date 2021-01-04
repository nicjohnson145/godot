package config

import (
	"encoding/json"
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
)

type configContents struct {
	AllFiles map[string]string   `json:"all_files"`
	Renders  map[string][]string `json:"renders"`
}

type Config struct {
	Target       string         `json:"target"`
	DotfilesRoot string         `json:"dotfiles_root"`
	content      configContents // The raw json content
	repoConfig   string         // Path to repo config, we'll need to rewrite it often
	Home         string         // Users home directory
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

	var content configContents
	if err := json.Unmarshal(bytes, &content); err != nil {
		panic(err)
	}

	c.content = content
}

func (c *Config) getAllFiles() map[string]file.File {
	allFiles := make(map[string]file.File)

	for key, path := range c.content.AllFiles {
		allFiles[key] = file.File{
			DestinationPath: util.ReplacePrefix(path, "~/", c.Home),
			TemplatePath:    filepath.Join(c.DotfilesRoot, "templates", key),
		}
	}
	return allFiles
}

func (c *Config) Write() error {
	bytes, err := json.Marshal(c.content)
	if err != nil {
		return err
	}
	prettyContents := pretty.PrettyOptions(bytes, &pretty.Options{Indent: "    "})
	return ioutil.WriteFile(c.repoConfig, prettyContents, 0644)
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
	c.content.AllFiles[template] = newDest
	return template, nil
}

func (c *Config) AddToTarget(target string, name string) error {
	if !c.IsValidFile(name) {
		return errors.New(fmt.Sprintf("unknown file name of %q", name))
	}
	c.content.Renders[target] = append(c.content.Renders[target], name)
	return nil
}

func (c *Config) RemoveFromTarget(target string, name string) error {
	files, ok := c.content.Renders[target]
	if !ok {
		return errors.New(fmt.Sprintf("unknown target %q", target))
	}
	var newFiles []string
	for _, fl := range files {
		if fl == name {
			continue
		}
		newFiles = append(newFiles, fl)
	}
	c.content.Renders[target] = newFiles
	return nil
}

func (c *Config) IsValidFile(name string) bool {
	_, ok := c.content.AllFiles[name]
	return ok
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
	for _, name := range c.content.Renders[target] {
		files[name] = allFiles[name]
	}
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
	fmt.Fprintln(w, "Target: " + target)
	c.writeFileMap(w, files)
}
