package repo

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func initRepo(t *testing.T) (shellGitRepo, func()) {
	dir, remove := help.CreateTempDir(t, "repo")
	cmd := exec.Command("git", "init", dir)
	err := cmd.Run()
	help.Ensure(t, err)

	return NewShellGitRepo(dir), remove
}

func TestIsWorkdirClean(t *testing.T) {
	t.Run("clean_workdir", func(t *testing.T) {
		r, remove := initRepo(t)
		defer remove()

		got, err := r.isWorkdirClean()
		help.Ensure(t, err)
		want := true

		if got != want {
			t.Fatalf("incorrect clean state, got %v want %v", got, want)
		}
	})

	t.Run("untracked_file", func(t *testing.T) {
		r, remove := initRepo(t)
		defer remove()
		help.WriteData(t, filepath.Join(r.Path, "some_file"), "")

		got, err := r.isWorkdirClean()
		help.Ensure(t, err)
		want := false

		if got != want {
			t.Fatalf("incorrect clean state, got %v want %v", got, want)
		}
	})
}

func TestDirtyFiles(t *testing.T) {
	t.Run("untracked_file", func(t *testing.T) {
		r, remove := initRepo(t)
		defer remove()
		help.WriteData(t, filepath.Join(r.Path, "some_file"), "")

		got, err := r.dirtyFiles()
		help.Ensure(t, err)
		want := []string{"some_file"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("incorrect files, got %v want %v", got, want)
		}
	})
}
