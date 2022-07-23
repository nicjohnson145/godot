package lib

import (
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

const (
	PackageManagerApt  = "apt"
	PackageManagerBrew = "brew"
)

var validPackageManagers = []string{
	PackageManagerApt,
	PackageManagerBrew,
}

func isValidPackageManager(s string) bool {
	return lo.Contains(validPackageManagers, s)
}

var _ Executor = (*SystemPackage)(nil)

type SystemPackage struct {
	Name     string `yaml:"name"`
	AptName  string `yaml:"apt"`
	BrewName string `yaml:"brew"`
}

func (s *SystemPackage) Execute(conf UserConfig, opts SyncOpts) {
	if opts.Quick {
		log.Debugf("skipping system package %v due to quick flag", s.Name)
		return
	}

	if conf.PackageManager == "" {
		log.Fatal("Package manager not configured, cannot install system packages")
	}

	log.Infof("Installing %v\n", s.Name)
	switch conf.PackageManager {
	case PackageManagerApt:
		s.executeApt(conf)
	case PackageManagerBrew:
		s.executeBrew(conf)
	default:
		log.Fatalf("Unknown package manager %v", conf.PackageManager)
	}

}

func (s *SystemPackage) GetName() string {
	return s.Name
}

func (s *SystemPackage) executeApt(conf UserConfig) {
	if s.AptName == "" {
		log.Fatal("No configured name for apt")
	}
	_, stderr, err := runCmd("/bin/sh", "-c", fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt install -y %v", s.AptName))
	if err != nil {
		log.Fatalf("Error during installation: %v\n%v", err, stderr)
	}
}

func (s *SystemPackage) executeBrew(conf UserConfig) {
	if s.BrewName == "" {
		log.Fatal("No configured name for brew")
	}
	_, _, err := runCmd("brew", "install", s.BrewName)
	if err != nil {
		log.Fatalf("Error during installation: %v", err)
	}
}
