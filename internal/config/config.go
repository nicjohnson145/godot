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

	"github.com/nicjohnson145/godot/internal/util"
	"github.com/tidwall/pretty"
)

const (
	APT     = "apt"
	BREW    = "brew"
	GIT     = "git"
	CURRENT = "<CURRENT>"
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

type StringMap map[string]string

type repoConfig struct {
	Files      StringMap            `json:"files"`
	Renders    map[string][]string  `json:"renders"`
	Bootstraps map[string]Bootstrap `json:"bootstraps"`
	Hosts      map[string]Host      `json:"hosts"`
}

type BootstrapItem struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
}

type Bootstrap map[string]BootstrapItem

type Host struct {
	Files      []string `json:"files"`
	Bootstraps []string `json:"bootstraps"`
}

func (c *repoConfig) makeMaps() {
	if c.Files == nil {
		c.Files = make(StringMap)
	}

	if c.Renders == nil {
		c.Renders = make(map[string][]string)
	}

	if c.Bootstraps == nil {
		c.Bootstraps = make(map[string]Bootstrap)
	}

	if c.Hosts == nil {
		c.Hosts = make(map[string]Host)
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
	TemplateRoot   string     // Template directory inside of dotfiles repo
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
	c.TemplateRoot = filepath.Join(c.DotfilesRoot, "templates")

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
		c.TemplateRoot = filepath.Join(config.DotfilesRoot, "templates")
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

func (c *Config) Write() error {
	bytes, err := json.Marshal(c.content)
	if err != nil {
		return err
	}
	prettyContents := pretty.PrettyOptions(bytes, &pretty.Options{Indent: "    "})
	return ioutil.WriteFile(c.repoConfig, prettyContents, 0644)
}

func (c *Config) GetAllFiles() StringMap {
	return c.content.Files
}

func (c *Config) GetFilesForTarget(target string) StringMap {
	if target == "" {
		target = c.Target
	}

	ret := make(StringMap)
	all := c.GetAllFiles()

	host, ok := c.content.Hosts[target]
	if !ok {
		return ret
	}

	for _, file := range host.Files {
		ret[file] = all[file]
	}

	return ret
}

func (c *Config) AddFile(template string, destination string) (string, error) {
	if template == "" {
		template = util.ToTemplateName(destination)
	}
	if c.IsKnownFile(template) {
		return "", errors.New(fmt.Sprintf("template name %q already exists", template))
	}
	newDest := util.ReplacePrefix(destination, c.Home, "~")
	c.content.Files[template] = newDest
	return template, nil
}

func (c *Config) AddTargetFile(target string, name string) error {
	if !c.IsKnownFile(name) {
		return fmt.Errorf("Unknown template of %q", name)
	}
	host, ok := c.content.Hosts[target]
	if !ok {
		host = Host{}
	}
	host.Files = append(host.Files, name)
	c.content.Hosts[target] = host
	return nil
}

func (c *Config) RemoveTargetFile(target string, name string) error {
	host, ok := c.content.Hosts[target]
	if !ok {
		return fmt.Errorf("unknown target %q", target)
	}

	newFiles, err := c.removeItem(host.Files, name)
	if err != nil {
		return fmt.Errorf("remove target file: %w", err)
	}
	host.Files = newFiles
	c.content.Hosts[target] = host
	return nil
}

func (c *Config) GetAllTemplateNames() []string {
	names := make([]string, 0, len(c.content.Files))
	for name := range c.content.Files {
		names = append(names, name)
	}
	return names
}

func (c *Config) GetAllTemplateNamesForTarget(target string) []string {
	host, ok := c.content.Hosts[target]
	if !ok {
		host = Host{}
	}
	return host.Files
}

func (c *Config) IsKnownFile(name string) bool {
	_, ok := c.content.Files[name]
	return ok
}

func (c *Config) GetTemplateFromFullPath(path string) (string, error) {
	for template, dest := range c.GetAllFiles() {
		fullDest := util.ReplacePrefix(dest, "~/", c.Home)
		if fullDest == path {
			return filepath.Join(c.TemplateRoot, template), nil
		}
	}
	return "", fmt.Errorf("Path %q is not managed by godot", path)
}

func (c *Config) ListAllFiles(w io.Writer) error {
	return c.writeStringMap(w, c.GetAllFiles())
}

func (c *Config) ListTargetFiles(target string, w io.Writer) error {
	return c.writeStringMap(w, c.GetFilesForTarget(target))
}

func (c *Config) writeStringMap(w io.Writer, m StringMap) error {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(w, 0, 1, 0, ' ', tabwriter.AlignRight)
	for _, key := range keys {
		_, err := fmt.Fprintln(tw, fmt.Sprintf("%v\t => %v", key, m[key]))
		if err != nil {
			return err
		}
	}

	return tw.Flush()
}

func (c *Config) GetAllBootstraps() map[string]Bootstrap {
	return c.content.Bootstraps
}

func (c *Config) GetBootstrapsForTarget(target string) map[string]Bootstrap {
	if target == "" {
		target = c.Target
	}

	all := c.GetAllBootstraps()
	ret := make(map[string]Bootstrap)

	host, ok := c.content.Hosts[target]
	if !ok {
		return ret
	}

	for _, key := range host.Bootstraps {
		ret[key] = all[key]
	}

	return ret
}

func (c *Config) AddBootstrapItem(item, manager, pkg, location string) {
	itemMap, ok := c.content.Bootstraps[item]
	if !ok {
		itemMap = make(map[string]BootstrapItem)
	}
	itemMap[manager] = BootstrapItem{Name: pkg, Location: location}
	c.content.Bootstraps[item] = itemMap
}

func (c *Config) isValidBootstrap(name string) bool {
	_, ok := c.content.Bootstraps[name]
	return ok
}

func (c *Config) AddTargetBootstrap(target string, name string) error {
	if !c.isValidBootstrap(name) {
		return fmt.Errorf("Unknown bootstrap item of %q", name)
	}

	current, ok := c.content.Hosts[target]
	if !ok {
		current = Host{}
	}
	current.Bootstraps = append(current.Bootstraps, name)
	c.content.Hosts[target] = current
	return nil
}

func (c *Config) RemoveTargetBootstrap(target string, name string) error {
	if target == "" {
		target = c.Target
	}

	host, ok := c.content.Hosts[target]
	if !ok {
		return fmt.Errorf("Unknown target of %q", target)
	}

	new, err := c.removeItem(host.Bootstraps, name)
	if err != nil {
		return err
	}
	host.Bootstraps = new
	c.content.Hosts[target] = host

	return nil
}

func (c *Config) removeItem(slice []string, item string) ([]string, error) {
	newSlice := make([]string, 0, len(slice))
	found := false
	for _, val := range slice {
		if val == item {
			found = true
			continue
		}
		newSlice = append(newSlice, val)
	}

	var err error
	if !found {
		err = fmt.Errorf("item %q not found", item)
	}
	return newSlice, err
}
