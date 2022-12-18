package lib

import (
	"fmt"
	"runtime"
	"os"
	"strings"
	"path"
	log "github.com/sirupsen/logrus"
	"github.com/carlmjohnson/requests"
	"context"
)

var _ Executor = (*Golang)(nil)

type Golang struct {
	Name string `yaml:"-"`
	Version string `yaml:"version" mapstructure:"version"`
}

func (g *Golang) Type() ExecutorType {
	return ExecutorTypeGolang
}

func (g *Golang) GetName() string {
	return g.Name
}

func (g *Golang) SetName(s string) {
	g.Name = s
}

func (g *Golang) Execute(_ UserConfig, _ SyncOpts, _ GodotConfig) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("golang installations only supported on linux")
	}

	log.Infof("ensuring golang %v", g.Version)

	log.Debug("checking installed version")
	out, _, err := runCmd("go", "version")
	if err == nil && g.getVersionFromOutput(out) == g.Version {
		log.Infof("go version %v already installed, skipping", g.Version)
		return nil
	}

	log.Debug("removing existing installation")
	// If we either don't have it, or it's the wrong version, make sure to clean out the old version
	// first per the golang docs. Run this command through the shell so we can elivate privileges
	_, _, err = runCmd("rm", "-rf", "/usr/local/go")
	if err != nil {
		return fmt.Errorf("error removing old golang installation: %w", err)
	}

	dir, err := os.MkdirTemp("", "godot-")
	if err != nil {
		return fmt.Errorf("unable to make temp directory")
	}
	defer os.RemoveAll(dir)

	log.Debug("downloading release tarball")
	filepath := path.Join(dir, g.getTarballName())
	err = requests.
		URL(fmt.Sprintf("https://go.dev/dl/%v", g.getTarballName())).
		ToFile(filepath).
		Fetch(context.Background())
	if err != nil {
		return fmt.Errorf("error downloading tarball: %w", err)
	}

	log.Debug("extracting tarball")
	_, _, err = runCmd("/bin/sh", "-c", fmt.Sprintf("sudo tar -C /usr/local -xzf %v", filepath))
	if err != nil {
		return fmt.Errorf("error unpacking tarball: %w", err)
	}

	return nil
}

func (g *Golang) getVersionFromOutput(out string) string {
	parts := strings.Split(out, " ")
	version := parts[2]
	return version[2:]
}

func (g *Golang) getTarballName() string {
	return fmt.Sprintf(
		"go%v.linux-%v.tar.gz",
		g.Version,
		runtime.GOARCH,
	)
}
