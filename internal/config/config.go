package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"text/tabwriter"

	"github.com/nicjohnson145/godot/internal/file"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/tidwall/pretty"
)

const (
	APT = "apt"
	BREW = "brew"
	GIT = "git"
)

var validManagers = []string{APT, BREW}

func IsValidPackageManager(candidate string) bool {
	for _, val := range validManagers {
		if candidate == val {
			return true
		}
	}
	return false
}

/*
{
	"all_files": {
		"dot_zshrc": "~/.zshrc"
	},
	"renders": {
		"host1": ["dot_zshrc"]
	},
	"all_bootstraps": {
		"ripgrep": {
			"brew": "ripgrep",
			"apt": "ripgrep"
		}
	},
	"bootstraps": {
		"host1": ["ripgrep"]
	}
}
*/

type repoConfig struct {
	AllFiles      map[string]string            `json:"all_files"`
	Renders       map[string][]string          `json:"renders"`
	AllBootstraps map[string]map[string]string `json:"all_bootstraps"`
	Bootstraps    map[string][]string          `json:"bootstraps"`
}

type Bootstrap struct {
	Method string
	Name   string
}

func (c *repoConfig) makeMaps() {
	if c.AllFiles == nil {
		c.AllFiles = make(map[string]string)
	}

	if c.Renders == nil {
		c.Renders = make(map[string][]string)
	}

	if c.AllBootstraps == nil {
		c.AllBootstraps = make(map[string]map[string]string)
	}

	if c.Bootstraps == nil {
		c.Bootstraps = make(map[string][]string)
	}
}

type usrConfig struct {
	Target         string `json:"target"`
	DotfilesRoot   string `json:"dotfiles_root"`
	PackageManager string `json:"package_manager"`
}

type Config struct {
	Target         string     // Name of the current target
	DotfilesRoot   string     // Root of the dotfiles repo
	content        repoConfig // The raw json content
	repoConfig     string     // Path to repo config, we'll need to rewrite it often
	Home           string     // Users home directory
	PackageManager string     // Configured package manager for bootstrapping
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
	c.content.makeMaps()
	return c
}

func (c *Config) parseUserConfig() {
	checkPanic := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	hostname, err := os.Hostname()
	checkPanic(err)

	c.Target = hostname
	c.DotfilesRoot = filepath.Join(c.Home, "dotfiles")

	switch runtime.GOOS {
	case "darwin":
		c.PackageManager = BREW
	case "linux":
		c.PackageManager = APT
	}

	conf := filepath.Join(c.Home, ".config", "godot", "config.json")
	if _, err := os.Stat(conf); os.IsNotExist(err) {
		// Missing config file
		return
	} else if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadFile(conf)
	checkPanic(err)

	var config usrConfig
	err = json.Unmarshal(bytes, &config)
	checkPanic(err)

	if config.Target != "" {
		c.Target = config.Target
	}

	if config.DotfilesRoot != "" {
		c.DotfilesRoot = config.DotfilesRoot
	}

	if config.PackageManager != "" {
		c.PackageManager = config.PackageManager
	}
}

func (c *Config) readRepoConfig() {
	configPath := filepath.Join(c.DotfilesRoot, "config.json")
	c.repoConfig = configPath
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.content = repoConfig{}
			return
		} else {
			panic(fmt.Errorf("error reading repo conf, %v", err))
		}
	}

	var content repoConfig
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

func (c *Config) GetTemplatesNamesForTarget(target string) []string {
	return c.content.Renders[target]
}

func (c *Config) GetAllTemplateNames() []string {
	names := make([]string, 0, len(c.content.AllFiles))
	for name := range c.content.AllFiles {
		names = append(names, name)
	}
	return names
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

func (c *Config) GetTemplateFromFullPath(path string) (string, error) {
	for _, fl := range c.getAllFiles() {
		if fl.DestinationPath == path {
			return fl.TemplatePath, nil
		}
	}
	return "", fmt.Errorf("Path %q is not managed by godot", path)
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
	fmt.Fprintln(w, "Target: "+target)
	c.writeFileMap(w, files)
}

func (c *Config) GetAllBootstraps() map[string][]Bootstrap {
	ret := make(map[string][]Bootstrap)
	for prog, installMap := range c.content.AllBootstraps {
		arr := make([]Bootstrap, 0, len(installMap))
		for method, name := range installMap {
			arr = append(arr, Bootstrap{Method: method, Name: name})
		}
		ret[prog] = arr
	}
	return ret
}

func (c *Config) GetBootstrapsForTarget(target string) map[string][]Bootstrap {
	if target == "" {
		target = c.Target
	}

	ret := make(map[string][]Bootstrap)
	keys, ok := c.content.Bootstraps[target]
	if !ok {
		return ret
	}

	all := c.GetAllBootstraps()

	for _, key := range keys {
		ret[key] = all[key]
	}

	return ret
}

func (c *Config) AddBootstrapItem(item string, manager string, pkg string) error {
	if !IsValidPackageManager(manager) {
		return fmt.Errorf("Invalid package manager of %q", manager)
	}

	m, ok := c.content.AllBootstraps[item]
	// Key didn't exist
	if !ok {
		m = make(map[string]string)
	}
	m[manager] = pkg
	c.content.AllBootstraps[item] = m
	return nil
}

func (c *Config) isValidBootstrap(name string) bool {
	_, ok := c.content.AllBootstraps[name]
	return ok
}


func (c *Config) AddTargetBootstrap(target string, name string) error {
	if target == "" {
		target = c.Target
	}

	if !c.isValidBootstrap(name) {
		return fmt.Errorf("Unknown bootstrap item of %q", name)
	}

	current, _ := c.content.Bootstraps[target]
	current = append(current, name)
	c.content.Bootstraps[target] = current
	return nil
}

func (c *Config) RemoveTargetBootstrap(target string, name string) error {
	if target == "" {
		target = c.Target
	}

	current, ok := c.content.Bootstraps[target]
	if !ok {
		return fmt.Errorf("Unknown target of %q", target)
	}

	new := make([]string, 0, len(current))
	for _, item := range current {
		if item == name {
			continue
		}

		new = append(new, item)
	}

	if len(new) == 0 {
		delete(c.content.Bootstraps, target)
	} else {
		c.content.Bootstraps[target] = new
	}

	return nil
}
