package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/file"
	"github.com/nicjohnson145/godot/internal/util"
)

type Builder struct {
	Getter util.HomeDirGetter
}

func (b *Builder) Build(force bool) error {
	conf := config.NewConfig(b.Getter)

	vars := file.TemplateVars{
		Target: conf.Target,
	}

	dir, err := b.ensureBuildDir(*conf)
	if err != nil {
		return err
	}

	for _, fl := range conf.GetTargetFiles() {
		err = fl.Render(dir, vars, force)
		if err != nil {
			err = fmt.Errorf("error rendering template %q, %v", fl.DestinationPath, err)
			return err
		}

		err = fl.Symlink(dir)
		if err != nil {
			err = fmt.Errorf("error symlinking template %q, %v", fl.DestinationPath, err)
			return err
		}
	}

	return nil
}

func (b *Builder) ensureBuildDir(conf config.Config) (string, error) {
	buildDir := filepath.Join(conf.DotfilesRoot, "build")
	err := os.MkdirAll(buildDir, 0744)
	return buildDir, err
}
