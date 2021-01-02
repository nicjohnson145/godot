package cmd

import (
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func setup(t *testing.T) (string, string, string, func()) {
	t.Helper()
	
	home, dotPath, remove := help.SetupFullConfig(t, "home", nil)

	confPath := filepath.Join(home, ".some_conf")
	help.WriteData(t, confPath, "contents")

	return home, confPath, dotPath, remove
}

func TestManage(t *testing.T) {
	t.Run("import and add", func(t *testing.T) {
		home, confPath, dotPath, remove := setup(t)
		defer remove()

		getter := &help.TempHomeDir{HomeDir: home}
		err := importFile(getter, confPath, "", true)
		if err != nil {
			t.Fatal(err)
		}

		help.AssertDirectoryContents(t, filepath.Join(dotPath, "templates"), []string{"dot_some_conf"})
		help.AssertTargetContents(t, dotPath, "home", []string{"dot_some_conf"})
	})
}
