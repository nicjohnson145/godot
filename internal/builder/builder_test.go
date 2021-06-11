package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	t.Run("simple no-template integration", func(t *testing.T) {
		expected := "zsh contents"
		home, _, remove := help.SetupFullConfig(t, "host3", &expected)
		defer remove()

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		require.NoError(t, err)
		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("templated integration", func(t *testing.T) {
		template := `{{ if oneOf . "host3" "other" }}{{ .Target }}{{ end }} zsh contents`
		home, _, remove := help.SetupFullConfig(t, "host3", &template)
		defer remove()

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		require.NoError(t, err)

		expected := "host3 zsh contents"
		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("build directory cleaned on each sync", func(t *testing.T) {
		expected := "contents"
		home, dotPath, remove := help.SetupFullConfig(t, "host3", &expected)
		defer remove()

		build := filepath.Join(dotPath, "build")
		err := os.Mkdir(build, 0744)
		require.NoError(t, err)
		help.WriteData(t, filepath.Join(build, "some_file"), "some_data")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build(false)
		require.NoError(t, err)
		help.AssertDirectoryContentsRecursive(t, build, []string{"dot_zshrc"})
	})

	t.Run("file exists as symlink", func(t *testing.T) {
		expected := "host3 zsh contents"
		home, dotPath, remove := help.SetupFullConfig(t, "host3", &expected)
		defer remove()

		err := os.Mkdir(filepath.Join(dotPath, "build"), 0744)
		require.NoError(t, err)
		buildFile := filepath.Join(dotPath, "build/dot_zshrc")
		help.WriteData(t, buildFile, "orig")
		os.Symlink(filepath.Join(home, ".zshrc"), filepath.Join(dotPath, "build/dot_zshrc"))

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build(false)
		require.NoError(t, err)
		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("file exists as file no force", func(t *testing.T) {
		expected := "zsh contents"
		home, _, remove := help.SetupFullConfig(t, "host3", &expected)
		defer remove()

		destPath := filepath.Join(home, ".zshrc")
		help.WriteData(t, destPath, "")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		require.Error(t, err)
		help.AssertFileContents(t, destPath, "")
	})

	t.Run("file exists as file with force", func(t *testing.T) {
		expected := "zsh contents"
		home, _, remove := help.SetupFullConfig(t, "host3", &expected)
		defer remove()

		destPath := filepath.Join(home, ".zshrc")
		help.WriteData(t, destPath, "")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(true)
		require.NoError(t, err)
		help.AssertFileContents(t, destPath, "zsh contents")
	})

	t.Run("import file", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "host3", nil)
		defer remove()

		expected := "contents"
		path := filepath.Join(home, ".some_conf")
		help.WriteData(t, path, expected)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		b.Import(path, "")

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotPath, "templates"), []string{"dot_some_conf"})
		help.AssertFileContents(t, filepath.Join(dotPath, "templates", "dot_some_conf"), expected)
	})

	t.Run("import missing file", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "host3", nil)
		defer remove()

		path := filepath.Join(home, ".some_conf")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		b.Import(path, "")

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotPath, "templates"), []string{"dot_some_conf"})
		help.AssertFileContents(t, filepath.Join(dotPath, "templates", "dot_some_conf"), "")
	})

	t.Run("import file as", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "host3", nil)
		defer remove()

		path := filepath.Join(home, ".some_conf")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		b.Import(path, "other_name")

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotPath, "templates"), []string{"other_name"})
	})

	t.Run("template error doesnt delete build dir", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "host3")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"files": {
				"some_file": "~/some_file"
			},
			"hosts": {
				"host3": {
					"files": ["some_file"]
				}
			}
		}`)

		// Touch some file in the build dir
		buildDir := filepath.Join(dotPath, "build")
		err := os.Mkdir(buildDir, 0744)
		require.NoError(t, err)
		help.WriteData(t, filepath.Join(buildDir, "orphan_file"), "")

		// Write an invalid template
		template := filepath.Join(dotPath, "templates", "some_file")
		help.WriteData(t, template, "{{ .NotAValidKey }}")

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build(true)
		require.Error(t, err)
		help.AssertDirectoryContentsRecursive(t, buildDir, []string{"orphan_file"})
	})
}
