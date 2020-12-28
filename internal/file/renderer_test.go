package file

import (
	"path/filepath"
	"testing"
)

func TestRenderer(t *testing.T) {
	t.Run("all files rendered and symlinked", func(t *testing.T) {
		src, removeSrc := createTempDir(t, "src")
		defer removeSrc()

		dotRoot, removeDotRoot := createTempDir(t, "dotfiles")
		defer removeDotRoot()

		home, removeHome := createTempDir(t, "home")
		defer removeHome()

		writeData(t, filepath.Join(src, "first_file"), "first contents")
		writeData(t, filepath.Join(src, "second_file"), "second contents")

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
		
		r, err := NewRenderer(files, &TempHomeDir{HomeDir: home}, dotRoot)
		if err != nil {
			t.Fatalf("error creating renderer, %v", err)
		}
		err = r.Render()
		if err != nil {
			t.Fatalf("error rendering, %v", err)
		}

		want := []string{
			".config",
			".config/second_file",
			".first_file",
		}
		assertDirectoryContents(t, home, want)
	})
}
