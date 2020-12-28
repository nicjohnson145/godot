package file

import (
	"path/filepath"
	"testing"
)

func TestFile(t *testing.T) {
	t.Run("files without templating are copied and symlinked", func(t *testing.T) {
		src, removeSrc := createTempDir(t, "src")
		defer removeSrc()

		build, removeBuild := createTempDir(t, "build")
		defer removeBuild()

		home, removeHome := createTempDir(t, "home")
		defer removeHome()

		writeData(t, filepath.Join(src, "some_file"), "the contents")

		f := File{
			DestinationPath: filepath.Join(home, ".some_other_file"),
			TemplatePath:    filepath.Join(src, "some_file"),
		}
		err := f.Build(build)
		if err != nil {
			t.Fatalf("error building, %v", err)
		}

		want := []string{".some_other_file"}
		assertDirectoryContents(t, home, want)
		assertSymlinkTo(
			t,
			filepath.Join(home, ".some_other_file"),
			filepath.Join(build, "some_file"),
		)
	})

	t.Run("missing destination sub dirs are created", func(t *testing.T) {
		src, removeSrc := createTempDir(t, "src")
		defer removeSrc()

		build, removeBuild := createTempDir(t, "build")
		defer removeBuild()

		home, removeHome := createTempDir(t, "home")
		defer removeHome()

		writeData(t, filepath.Join(src, "some_file"), "the contents")

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
		assertDirectoryContents(t, home, want)
	})
}
