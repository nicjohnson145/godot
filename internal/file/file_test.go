package file

import (
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestFile(t *testing.T) {
	t.Run("render writes to build dir", func(t *testing.T) {
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

		vars := TemplateVars{}
		err := f.Render(build, vars)
		if err != nil {
			t.Fatalf("error rendering, %v", err)
		}

		want := []string{"some_file"}
		help.AssertDirectoryContents(t, build, want)
		help.AssertFileContents(t, filepath.Join(build, "some_file"), "the contents")
	})

	t.Run("templated data is expanded", func(t *testing.T) {
		src, removeSrc := help.CreateTempDir(t, "src")
		defer removeSrc()

		build, removeBuild := help.CreateTempDir(t, "build")
		defer removeBuild()
		help.WriteData(t, filepath.Join(src, "some_file"), "the {{ .Target }} contents")

		f := File{
			TemplatePath:    filepath.Join(src, "some_file"),
		}

		vars := TemplateVars{Target: "host"}
		err := f.Render(build, vars)
		if err != nil {
			t.Fatalf("error rendering, %v", err)
		}

		want := []string{"some_file"}
		help.AssertDirectoryContents(t, build, want)
		help.AssertFileContents(t, filepath.Join(build, "some_file"), "the host contents")
	})
}
