package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestConfig(t *testing.T) {
	t.Run("missing config file panics", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Code did not panic")
			}
		}()

		NewConfig(&help.TempHomeDir{HomeDir: dir})
	})

	t.Run("missing target panics", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Code did not panic")
			}
		}()

		help.WriteConfig(t, dir, `{}`)

		NewConfig(&help.TempHomeDir{HomeDir: dir})
	})

	t.Run("build target pulled from file", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()

		help.WriteConfig(t, dir, `{"target": "my_host"}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: dir})

		if c.Target != "my_host" {
			t.Errorf("incorrect target, got %q want %q", c.Target, "my_host")
		}
	})

	t.Run("dotfiles root inferred if missing", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()

		help.WriteConfig(t, dir, `{"target": "my_host"}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: dir})

		expected := filepath.Join(dir, "dotfiles")
		if c.DotfilesRoot != expected {
			t.Errorf("dotfiles root not inferred, got %q want %q", c.DotfilesRoot, expected)
		}
	})

	t.Run("dotfiles root can be overridden", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()

		help.WriteConfig(t, dir, `{"target": "my_host", "dotfiles_root": "some_path"}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: dir})

		expected := "some_path"
		if c.DotfilesRoot != expected {
			t.Errorf("dotfiles root not pulled from file, got %q want %q", c.DotfilesRoot, expected)
		}
	})

	t.Run("missing repo config means no files", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotfiles, removeDots := help.CreateTempDir(t, "dotfiles")
		defer removeDots()

		userConf := fmt.Sprintf(`{"target": "my_host", "dotfiles_root": "%v"}`, dotfiles)
		help.WriteConfig(t, home, userConf)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		if len(c.Files) != 0 {
			t.Errorf("missing repo config should result in 0 files, got %v", len(c.Files))
		}
	})
}
