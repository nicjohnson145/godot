package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestBuilder(t *testing.T) {
	t.Run("simple no-template integration", func(t *testing.T) {
		expected := "zsh contents"
		home, _, remove := help.SetupFullConfig(t, "host", &expected)
		defer remove()

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("templated integration", func(t *testing.T) {
		template := `{{ if oneOf . "host" "other" }}{{ .Target }}{{ end }} zsh contents`
		home, _, remove := help.SetupFullConfig(t, "host", &template)
		defer remove()

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		expected := "host zsh contents"
		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("build directory cleaned on each sync", func(t *testing.T) {
		expected := "contents"
		home, dotPath, remove := help.SetupFullConfig(t, "host", &expected)
		defer remove()

		build := filepath.Join(dotPath, "build")
		err := os.Mkdir(build, 0744)
		if err != nil {
			t.Fatal(err)
		}
		help.WriteData(t, filepath.Join(build, "some_file"), "some_data")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build(false)
		if err != nil {
			t.Fatal(err)
		}

		help.AssertDirectoryContents(t, build, []string{"dot_zshrc"})
	})

	t.Run("file exists as symlink", func(t *testing.T) {
		expected := "host zsh contents"
		home, dotPath, remove := help.SetupFullConfig(t, "host", &expected)
		defer remove()

		err := os.Mkdir(filepath.Join(dotPath, "build"), 0744)
		if err != nil {
			t.Fatalf("err creating dir, %v", err)
		}
		buildFile := filepath.Join(dotPath, "build/dot_zshrc")
		help.WriteData(t, buildFile, "orig")
		os.Symlink(filepath.Join(home, ".zshrc"), filepath.Join(dotPath, "build/dot_zshrc"))

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build(false)
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("file exists as file no force", func(t *testing.T) {
		expected := "zsh contents"
		home, _, remove := help.SetupFullConfig(t, "host", &expected)
		defer remove()

		destPath := filepath.Join(home, ".zshrc")
		help.WriteData(t, destPath, "")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		if err == nil {
			t.Fatalf("code should have errored")
		}

		help.AssertFileContents(t, destPath, "")
	})

	t.Run("file exists as file with force", func(t *testing.T) {
		expected := "zsh contents"
		home, _, remove := help.SetupFullConfig(t, "host", &expected)
		defer remove()

		destPath := filepath.Join(home, ".zshrc")
		help.WriteData(t, destPath, "")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(true)
		if err != nil {
			t.Fatalf("error building, %v", err)
		}

		help.AssertFileContents(t, destPath, "zsh contents")
	})

	t.Run("import file", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "host", nil)
		defer remove()

		expected := "contents"
		path := filepath.Join(home, ".some_conf")
		help.WriteData(t, path, expected)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		b.Import(path, "")

		help.AssertDirectoryContents(t, filepath.Join(dotPath, "templates"), []string{"dot_some_conf"})
		help.AssertFileContents(t, filepath.Join(dotPath, "templates", "dot_some_conf"), expected)
	})

	t.Run("import missing file", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "host", nil)
		defer remove()

		path := filepath.Join(home, ".some_conf")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		b.Import(path, "")

		help.AssertDirectoryContents(t, filepath.Join(dotPath, "templates"), []string{"dot_some_conf"})
		help.AssertFileContents(t, filepath.Join(dotPath, "templates", "dot_some_conf"), "")
	})

	t.Run("import file as", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "host", nil)
		defer remove()

		path := filepath.Join(home, ".some_conf")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		b.Import(path, "other_name")

		help.AssertDirectoryContents(t, filepath.Join(dotPath, "templates"), []string{"other_name"})
	})
}
