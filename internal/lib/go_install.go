package lib

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

var _ Executor = (*GoInstall)(nil)

type GoInstall struct {
	Name    string         `yaml:"-"`
	Package string         `yaml:"package" mapstructure:"package"`
	Version string         `yaml:"version" mapstructure:"version"`
	log     zerolog.Logger `yaml:"-"`
}

func (g *GoInstall) SetLogger(log zerolog.Logger) {
	g.log = log
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

func (g *GoInstall) Validate() error {
	var errs *multierror.Error

	if g.Package == "" {
		errs = multierror.Append(errs, fmt.Errorf("package is required"))
	}

	return errs.ErrorOrNil()
}

func (g *GoInstall) Execute(_ UserConfig, _ SyncOpts, _ GodotConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("go-install currently only supports linux")
	}
	g.log.Info().Str("package", g.Package).Msg("go installing")

	version := "latest"
	if g.Version != "" {
		version = g.Version
	}
	_, _, err := runCmd("/usr/local/go/bin/go", "install", g.Package+"@"+version)
	if err != nil {
		return fmt.Errorf("error installing package: %w", err)
	}
	return nil
}
