package builder

import (
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/nicjohnson145/godot/internal/file"
)

type Builder struct {
	Getter util.HomeDirGetter
}

func (b *Builder) Build() error {
	conf := config.NewConfig(b.Getter)
	renderer := file.NewRenderer(conf.Files, conf.DotfilesRoot)
	err := renderer.Render()
	if err != nil {
		return err
	}
	return nil
}
