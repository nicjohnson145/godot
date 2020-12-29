package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"reflect"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/tidwall/gjson"
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

	t.Run("target used to extract relevant files from repo config", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotfiles, removeDots := help.CreateTempDir(t, "dotfiles")
		defer removeDots()

		userConf := fmt.Sprintf(`{"target": "my_host", "dotfiles_root": "%v"}`, dotfiles)
		help.WriteConfig(t, home, userConf)

		confData := fmt.Sprintf(`{
			"all_files": {"dot_zshrc": "~/.zshrc", "init.vim": "~/.config/nvim/init.vim"},
			"renders": {
				"my_host": ["init.vim"],
				"other_host": ["dot_zshrc", "init.vim"]
			}
		}`)
		help.WriteRepoConf(t, dotfiles, confData)

		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		if len(c.Files) != 1 {
			t.Errorf("Expected 1 file, got %v", len(c.Files))
		}

		f := c.Files[0]
		expectedDest := filepath.Join(home, ".config/nvim/init.vim")
		if f.DestinationPath != expectedDest {
			t.Errorf("incorrect destination, got %q want %q", f.DestinationPath, expectedDest)
		}

		expectedTemplate := filepath.Join(dotfiles, "templates", "init.vim")
		if f.TemplatePath != expectedTemplate {
			t.Errorf("incorrect template path, got %q want %q", f.TemplatePath, expectedTemplate)
		}
	})

	t.Run("add files to repo config", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotPath := filepath.Join(home, "dotfiles")
		err := os.Mkdir(dotPath, 0744)
		if err != nil {
			t.Fatalf("error creating dir, %v", err)
		}

		userConf := fmt.Sprintf(`{"target": "my_host", "dotfiles_root": "%v"}`, dotPath)
		help.WriteConfig(t, home, userConf)

		confData := fmt.Sprintf(`{
			"all_files": {"dot_zshrc": "~/.zshrc"},
			"renders": {
				"other_host": ["dot_zshrc"]
			}
		}`)
		help.WriteRepoConf(t, dotPath, confData)

		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		c.AddFile("some_config", "~/.some_config")
		err = c.Write()
		if err != nil {
			t.Fatalf("error writing config, %v", err)
		}

		contents := help.ReadFile(t, filepath.Join(dotPath, "config.json"))
		value := gjson.Get(contents, "all_files")

		actual := make(map[string]string)
		value.ForEach(func(key, value gjson.Result) bool {
			actual[key.String()] = value.String()
			return true
		})

		expected := map[string]string{
			"dot_zshrc": "~/.zshrc",
			"some_config": "~/.some_config",
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("all files incorrect, got %v want %v", actual, expected)
		}
	})

	t.Run("add file to target", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotPath := filepath.Join(home, "dotfiles")
		err := os.Mkdir(dotPath, 0744)
		if err != nil {
			t.Fatalf("error creating dir, %v", err)
		}

		userConf := fmt.Sprintf(`{"target": "my_host", "dotfiles_root": "%v"}`, dotPath)
		help.WriteConfig(t, home, userConf)

		confData := fmt.Sprintf(`{
			"all_files": {"dot_zshrc": "~/.zshrc"},
			"renders": {
				"other_host": ["dot_zshrc"]
			}
		}`)
		help.WriteRepoConf(t, dotPath, confData)

		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		err = c.AddToTarget("my_host", "dot_zshrc")
		if err != nil {
			t.Fatalf("error adding, %v", err)
		}
		err = c.Write()
		if err != nil {
			t.Fatalf("error writing config, %v", err)
		}

		contents := help.ReadFile(t, filepath.Join(dotPath, "config.json"))
		value := gjson.Get(contents, "renders.my_host")
		var actual []string
		value.ForEach(func(key, value gjson.Result) bool {
			actual = append(actual, value.String())
			return true
		})

		expected := []string{"dot_zshrc"}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("target files incorrect, got %v want %v", actual, expected)
		}
	})
}

