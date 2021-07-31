package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"text/tabwriter"

	"github.com/jinzhu/copier"
	"github.com/tidwall/pretty"
)

func NewConfig(getter HomeDirGetter) *Config {
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
	c.content.setGithubReleaseNames()
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
		c.PackageManagers = []string{BREW, GIT}
	case "linux":
		c.PackageManagers = []string{APT, GIT}
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

	if len(config.PackageManagers) != 0 {
		c.PackageManagers = config.PackageManagers
	}

	if config.GithubUser != "" {
		c.GithubUser = config.GithubUser
	}
}

func (c *Config) readRepoConfig() {
	configPath := filepath.Join(c.DotfilesRoot, "config.json")
	c.repoConfig = configPath
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.content = RepoConfig{}
			return
		} else {
			panic(fmt.Errorf("error reading repo conf, %v", err))
		}
	}

	var content RepoConfig
	if err := json.Unmarshal(bytes, &content); err != nil {
		panic(err)
	}

	c.content = content
}

func (c *Config) GetRawContent() RepoConfig {
	newConf := RepoConfig{}
	copier.CopyWithOption(&newConf, &c.content, copier.Option{DeepCopy: true})
	return newConf
}

func (c *Config) Write() error {
	bytes, err := json.Marshal(c.content)
	if err != nil {
		return err
	}
	prettyContents := pretty.PrettyOptions(bytes, &pretty.Options{Indent: "    "})
	return ioutil.WriteFile(c.repoConfig, prettyContents, 0644)
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
		err = fmt.Errorf("%q: %w", item, NotFoundError)
	}
	return newSlice, err
}

func (c *Config) GetAllTargets() []string {
	targets := []string{}
	for target := range c.content.Hosts {
		targets = append(targets, target)
	}
	return targets
}

func IsValidPackageManager(candidate string) bool {
	for _, val := range ValidManagers {
		if candidate == val {
			return true
		}
	}
	return false
}
