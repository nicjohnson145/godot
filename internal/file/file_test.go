package file

import (
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestFile(t *testing.T) {
	t.Run("files without templating are copied and symlinked", func(t *testing.T) {
		src, removeSrc := help.CreateTempDir(t, "src")
		defer removeSrc()

		build, removeBuild := help.CreateTempDir(t, "build")
		defer removeBuild()

		home, removeHome := help.CreateTempDir(t, "home")
		defer removeHome()

		help.WriteData(t, filepath.Join(src, "some_file"), "the contents")

		f := File{
			DestinationPath: filepath.Join(home, ".some_other_file"),
			TemplatePath:    filepath.Join(src, "some_file"),
		}
		err := f.Build(build)
		if err != nil {
			t.Fatalf("error building, %v", err)
		}

		want := []string{".some_other_file"}
		help.AssertDirectoryContents(t, home, want)
		help.AssertSymlinkTo(
			t,
			filepath.Join(home, ".some_other_file"),
			filepath.Join(build, "some_file"),
		)
	})

	t.Run("missing destination sub dirs are created", func(t *testing.T) {
		src, removeSrc := help.CreateTempDir(t, "src")
		defer removeSrc()

		build, removeBuild := help.CreateTempDir(t, "build")
		defer removeBuild()

		home, removeHome := help.CreateTempDir(t, "home")
		defer removeHome()

		help.WriteData(t, filepath.Join(src, "some_file"), "the contents")

		f := File{
			DestinationPath: filepath.Join(home, "subdir", ".some_other_file"),
			TemplatePath:    filepath.Join(src, "some_file"),
		}
		err := f.Build(build)
		if err != nil {
			t.Fatalf("error building, %v", err)
		}

		want := []string{
			"subdir",
			"subdir/.some_other_file",
		}
		help.AssertDirectoryContents(t, home, want)
	})
}
