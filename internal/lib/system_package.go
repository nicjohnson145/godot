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
	Name     string `yaml:"-"`
	AptName  string `yaml:"apt" mapstructure:"apt"`
	BrewName string `yaml:"brew" mapstructure:"brew"`
}

func (s *SystemPackage) Type() ExecutorType {
	return ExecutorTypeSysPackage
}

func (s *SystemPackage) Execute(conf UserConfig, _ SyncOpts, _ GodotConfig) error {
	if conf.PackageManager == "" {
		return fmt.Errorf("eackage manager not configured, cannot install system packages")
	}

	log.Infof("Installing %v\n", s.Name)
	var err error
	switch conf.PackageManager {
	case PackageManagerApt:
		err = s.executeApt()
	case PackageManagerBrew:
		err = s.executeBrew()
	default:
		err = fmt.Errorf("enknown package manager %v", conf.PackageManager)
	}
	if err != nil {
		return fmt.Errorf("error during execution: %w", err)
	}

	return nil
}

func (s *SystemPackage) GetName() string {
	return s.Name
}

func (s *SystemPackage) SetName(n string) {
	s.Name = n
}

func (s *SystemPackage) executeApt() error {
	if s.AptName == "" {
		return fmt.Errorf("eo configured name for apt")
	}
	_, stderr, err := runCmd("/bin/sh", "-c", fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt install -y %v", s.AptName))
	if err != nil {
		return fmt.Errorf("error during installation: %v\n%v", err, stderr)
	}
	return nil
}

func (s *SystemPackage) executeBrew() error {
	if s.BrewName == "" {
		return fmt.Errorf("eo configured name for brew")
	}
	_, _, err := runCmd("brew", "install", s.BrewName)
	if err != nil {
		return fmt.Errorf("error during installation: %v", err)
	}
	return nil
}
