package lib

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"runtime"
)

var _ Executor = (*GoInstall)(nil)

type GoInstall struct {
	Name string `yaml:"-"`
	Package string `yaml:"package" mapstructure:"package"`
	Version string `yaml:"version" mapstructure:"version"`
}

func (g *GoInstall) Type() ExecutorType {
	return ExecutorTypeGoInstall
}

func (g *GoInstall) GetName() string {
	return g.Name
}

func (g *GoInstall) SetName(s string) {
	g.Name = s
}

func (g *GoInstall) Execute(_ UserConfig, _ SyncOpts, _ GodotConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("go-install currently only supports linux")
	}
	log.Infof("go installing %v", g.Package)

	version := "latest"
	if g.Version != "" {
		version = g.Version
	}
	_, _, err := runCmd("/usr/local/go/bin/go", "install", g.Package + "@" + version)
	if err != nil {
		return fmt.Errorf("error installing package: %w", err)
	}
	return nil
}
