package controller

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/nicjohnson145/godot/internal/repo"
)

func getController(t *testing.T, home string) Controller {
	t.Helper()
	return NewController(ControllerOpts{
		HomeDirGetter: &help.TempHomeDir{HomeDir: home},
		Repo:          repo.NoopRepo{},
	})
}

func TestSync(t *testing.T) {}

func TestImport(t *testing.T) {
	setup := func(t *testing.T) (string, string, string, func()) {
		t.Helper()

		home, dotPath, remove := help.SetupFullConfig(t, "home", nil)

		confPath := filepath.Join(home, ".some_conf")
		help.WriteData(t, confPath, "contents")

		return home, confPath, dotPath, remove
	}

	t.Run("import and add", func(t *testing.T) {
		home, confPath, dotPath, remove := setup(t)
		defer remove()

		controller := getController(t, home)
		err := controller.Import(confPath, "", ImportOpts{NoAdd: false})
		help.Ensure(t, err)

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotPath, "templates"), []string{"dot_some_conf"})
		help.AssertTargetContents(t, dotPath, "home", []string{"dot_some_conf"})
		help.AssertPresentInAllFiles(t, dotPath, map[string]string{"dot_some_conf": "~/.some_conf"})
	})

	t.Run("import as", func(t *testing.T) {
		home, confPath, dotPath, remove := setup(t)
		defer remove()

		controller := getController(t, home)
		err := controller.Import(confPath, "other_name", ImportOpts{NoAdd: true})
		help.Ensure(t, err)

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotPath, "templates"), []string{"other_name"})
		help.AssertTargetContents(t, dotPath, "home", []string{})
		help.AssertPresentInAllFiles(t, dotPath, map[string]string{"other_name": "~/.some_conf"})
	})
}

func TestTarget(t *testing.T) {
	setup := func(t *testing.T, target string) (string, string, func()) {
		t.Helper()

		home, dotpath, remove := help.SetupDirectories(t, target)
		help.WriteRepoConf(t, dotpath, `{
			"files": {
				"dot_zshrc": "~/.zshrc",
				"some_conf": "~/some_conf",
				"last_conf": "~/last_conf"
			},
			"hosts": {
				"host": {
					"files": ["dot_zshrc"]
				},
				"other": {
					"files": ["some_conf", "last_conf"]
				}
			}
		}`)

		return home, dotpath, remove
	}

	t.Run("list_all", func(t *testing.T) {
		home, _, remove := setup(t, "home")
		defer remove()

		c := getController(t, home)
		w := bytes.NewBufferString("")

		help.Ok(t, c.ListAllFiles(w))

		got := w.String()
		want := strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
			"last_conf => ~/last_conf",
			"some_conf => ~/some_conf",
		}, "\n") + "\n"

		help.Equals(t, want, got)
	})

	t.Run("show target", func(t *testing.T) {
		home, _, remove := setup(t, "home")
		defer remove()

		c := getController(t, home)

		w := bytes.NewBufferString("")
		c.TargetShowFiles("other", w)

		got := w.String()
		want := strings.Join([]string{
			"last_conf => ~/last_conf",
			"some_conf => ~/some_conf",
		}, "\n") + "\n"

		help.Equals(t, want, got)

		w = bytes.NewBufferString("")
		c.TargetShowFiles("host", w)

		got = w.String()
		want = strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
		}, "\n") + "\n"

		help.Equals(t, want, got)
	})

	t.Run("show target no target", func(t *testing.T) {
		home, _, remove := setup(t, "other")
		defer remove()

		c := getController(t, home)

		w := bytes.NewBufferString("")
		c.TargetShowFiles("", w)

		got := w.String()
		want := strings.Join([]string{
			"last_conf => ~/last_conf",
			"some_conf => ~/some_conf",
		}, "\n") + "\n"

		help.Equals(t, want, got)

		w = bytes.NewBufferString("")
		c.TargetShowFiles("host", w)

		got = w.String()
		want = strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
		}, "\n") + "\n"

		help.Equals(t, want, got)
	})

	t.Run("add target", func(t *testing.T) {
		home, dotpath, remove := setup(t, "home")
		defer remove()
		c := getController(t, home)
		err := c.TargetAddFile("host", []string{"last_conf"})
		help.Ensure(t, err)
		help.AssertTargetContents(t, dotpath, "host", []string{"dot_zshrc", "last_conf"})
	})

	t.Run("add target no target", func(t *testing.T) {
		home, dotpath, remove := setup(t, "host")
		defer remove()
		c := getController(t, home)
		err := c.TargetAddFile("", []string{"last_conf"})
		help.Ensure(t, err)
		help.AssertTargetContents(t, dotpath, "host", []string{"dot_zshrc", "last_conf"})
	})

	t.Run("remove target", func(t *testing.T) {
		home, dotpath, remove := setup(t, "home")
		defer remove()
		c := getController(t, home)
		err := c.TargetRemoveFile("other", []string{"last_conf"})
		help.Ensure(t, err)
		help.AssertTargetContents(t, dotpath, "other", []string{"some_conf"})
	})

	t.Run("remove target no target", func(t *testing.T) {
		home, dotpath, remove := setup(t, "other")
		defer remove()
		c := getController(t, home)
		err := c.TargetRemoveFile("", []string{"last_conf"})
		help.Ensure(t, err)
		help.AssertTargetContents(t, dotpath, "other", []string{"some_conf"})
	})
}

func TestEditFile(t *testing.T) {
	setup := func(t *testing.T) (string, string, func()) {
		t.Helper()

		home, dotpath, remove := help.SetupDirectories(t, "host")
		help.WriteRepoConf(t, dotpath, `{
			"files": {
				"some_conf": "~/some_conf"
			},
			"hosts": {
				"host": {
					"files": ["some_conf"]
				}
			}
		}`)
		return home, dotpath, remove
	}

	t.Run("basic_edit", func(t *testing.T) {
		home, dotpath, remove := setup(t)
		defer remove()

		// "Mocking" editor so ensure the file isn't already setup
		os.Setenv("EDITOR", "touch")
		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotpath, "templates"), []string{})
		help.AssertDirectoryContents(t, home, []string{".config", "dotfiles"})

		c := getController(t, home)
		err := c.EditFile([]string{filepath.Join(home, "some_conf")}, EditOpts{NoSync: false})

		help.Ensure(t, err)

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotpath, "templates"), []string{"some_conf"})
		help.AssertDirectoryContents(t, home, []string{".config", "dotfiles", "some_conf"})
	})

	t.Run("edit_no_sync", func(t *testing.T) {
		home, dotpath, remove := setup(t)
		defer remove()

		// "Mocking" editor so ensure the file isn't already setup
		os.Setenv("EDITOR", "touch")
		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotpath, "templates"), []string{})
		help.AssertDirectoryContents(t, home, []string{".config", "dotfiles"})

		c := getController(t, home)
		err := c.EditFile([]string{filepath.Join(home, "some_conf")}, EditOpts{NoSync: true})
		help.Ensure(t, err)

		help.AssertDirectoryContentsRecursive(t, filepath.Join(dotpath, "templates"), []string{"some_conf"})
		help.AssertDirectoryContents(t, home, []string{".config", "dotfiles"})
	})
}
