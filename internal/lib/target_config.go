package lib

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

type TargetConfig struct {
	ConfigFiles    []ConfigFile      `yaml:"config-files,flow"`
	GithubReleases []GithubRelease   `yaml:"github-releases,flow"`
	GitRepos       []GitRepo         `yaml:"git-repos,flow"`
	SystemPackages []SystemPackage   `yaml:"system-packages,flow"`
	Bundles        []Bundle          `yaml:"bundles,flow"`
	Targets        map[string]Target `yaml:"targets"`
}

type Target struct {
	ConfigFiles    []string `yaml:"config-files,flow"`
	GithubReleases []string `yaml:"github-releases,flow"`
	GitRepos       []string `yaml:"git-repos,flow"`
	SystemPackages []string `yaml:"system-packages,flow"`
	Bundles        []string `yaml:"bundles,flow"`
}

type Bundle struct {
	Name   string `yaml:"name"`
	Target `yaml:",inline"`
}

func (b Bundle) ToTarget() Target {
	return Target{
		ConfigFiles:    b.ConfigFiles,
		GithubReleases: b.GithubReleases,
		GitRepos:       b.GitRepos,
		SystemPackages: b.SystemPackages,
		Bundles:        b.Bundles,
	}
}

func (b Bundle) GetName() string {
	return b.Name
}

func NewTargetConfig(userConf UserConfig) TargetConfig {
	return NewTargetConfigFromPath(path.Join(userConf.CloneLocation, "config.yaml"))
}

func NewTargetConfigFromPath(path string) TargetConfig {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading target config: %v", err)
	}

	var conf TargetConfig
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		log.Fatalf("Error parsing target config: %v", err)
	}

	return conf
}
