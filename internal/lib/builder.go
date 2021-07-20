package lib

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/nicjohnson145/godot/internal/util"
)

type Builder struct {
	Getter util.HomeDirGetter
	Config *Config
}

func (b *Builder) Build(force bool) error {
	b.ensureConfig()

	vars := b.makeTemplateVars()

	// Before clearing the build directory, make sure nothing is going to error out, i.e fail safe
	for _, fl := range b.buildFileObjs(b.Config.GetFilesForTarget(b.Config.Target)) {
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

	var errs *multierror.Error
	// Actually render/symlink the files
	for _, fl := range b.buildFileObjs(b.Config.GetFilesForTarget(b.Config.Target)) {
		err = fl.Render(dir, vars, force)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("rendering: %w", err))
			continue
		}

		err = fl.Symlink(dir)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("symlinking: %w", err))
		}
	}

	return errs.ErrorOrNil()
}

func (b *Builder) buildFileObjs(m StringMap) []File {
	files := make([]File, 0, len(m))
	for tmpl, dest := range m {
		files = append(files, File{
			TemplatePath:    filepath.Join(b.Config.TemplateRoot, tmpl),
			DestinationPath: util.ReplacePrefix(dest, "~/", b.Config.Home),
		})
	}
	return files
}

func (b *Builder) makeTemplateVars() TemplateVars {
	return TemplateVars{
		Target:     b.Config.Target,
		Submodules: filepath.Join(b.Config.DotfilesRoot, "submodules"),
		Home:       b.Config.Home,
	}
}

func (b *Builder) ensureConfig() {
	if b.Config == nil {
		b.Config = NewConfig(b.Getter)
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

	if b.Config.IsKnownFile(as) {
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
