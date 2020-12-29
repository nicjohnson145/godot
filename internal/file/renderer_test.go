package file

import (
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestRenderer(t *testing.T) {
	t.Run("all files rendered and symlinked", func(t *testing.T) {
		src, removeSrc := help.CreateTempDir(t, "src")
		defer removeSrc()

		dotRoot, removeDotRoot := help.CreateTempDir(t, "dotfiles")
		defer removeDotRoot()

		home, removeHome := help.CreateTempDir(t, "home")
		defer removeHome()

		help.WriteData(t, filepath.Join(src, "first_file"), "first contents")
		help.WriteData(t, filepath.Join(src, "second_file"), "second contents")

		files := []File{
			{
				DestinationPath: filepath.Join(home, ".first_file"),
				TemplatePath:    filepath.Join(src, "first_file"),
			},
			{
				DestinationPath: filepath.Join(home, ".config", "second_file"),
				TemplatePath:    filepath.Join(src, "second_file"),
			},
		}
		
		r := NewRenderer(files, dotRoot)
		err := r.Render()
		if err != nil {
			t.Fatalf("error rendering, %v", err)
		}

		want := []string{
			".config",
			".config/second_file",
			".first_file",
		}
		help.AssertDirectoryContents(t, home, want)
	})
}
