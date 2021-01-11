package builder

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/file"
	"github.com/nicjohnson145/godot/internal/util"
)

type Builder struct {
	Getter util.HomeDirGetter
	Config *config.Config
}

func (b *Builder) Build(force bool) error {
	b.ensureConfig()

	vars := b.makeTemplateVars()

	// Before clearing the build directory, make sure nothing is going to error out, i.e fail safe
	for _, fl := range b.Config.GetTargetFiles() {
		_, err := fl.Execute(vars)
		if err != nil {
			return err
		}
	}

	// Clean out the build directory
	dir, err := b.ensureBuildDir()
	if err != nil {
		return err
	}

	// Actually render/symlink the files
	for _, fl := range b.Config.GetTargetFiles() {
		err = fl.Render(dir, vars, force)
		if err != nil {
			return fmt.Errorf("error rendering: %w", err)
		}

		err = fl.Symlink(dir)
		if err != nil {
			return fmt.Errorf("error symlinking %q: %w", fl.DestinationPath, err)
		}
	}

	return nil
}

func (b *Builder) makeTemplateVars() file.TemplateVars {
	return file.TemplateVars{
		Target:     b.Config.Target,
		Submodules: filepath.Join(b.Config.DotfilesRoot, "submodules"),
		Home:       b.Config.Home,
	}
}

func (b *Builder) ensureConfig() {
	if b.Config == nil {
		b.Config = config.NewConfig(b.Getter)
	}
}

func (b *Builder) ensureBuildDir() (string, error) {
	buildDir := filepath.Join(b.Config.DotfilesRoot, "build")
	err := os.RemoveAll(buildDir)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(buildDir, 0744)
	return buildDir, err
}

func (b *Builder) Import(path string, as string) error {
	b.ensureConfig()

	readData := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		readData = false
	} else if err != nil {
		return err
	}

	if as == "" {
		as = util.ToTemplateName(path)
	}

	if b.Config.IsValidFile(as) {
		return errors.New("file %q already managed, use --as to give it a different name")
	}

	var content []byte
	if readData {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		content = bytes
	}

	return ioutil.WriteFile(filepath.Join(b.Config.DotfilesRoot, "templates", as), content, 0644)
}
