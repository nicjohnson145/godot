package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"reflect"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestConfig(t *testing.T) {
	t.Run("missing config file ok", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()
		c := NewConfig(&help.TempHomeDir{HomeDir: dir})

		name, err := os.Hostname()
		help.Ensure(t, err)

		if c.Target != name {
			t.Errorf("incorrect target inferred, got %q want %q", c.Target, name)
		}

		expected := filepath.Join(dir, "dotfiles")
		if c.DotfilesRoot != expected {
			t.Errorf("dotfiles root not inferred, got %q want %q", c.DotfilesRoot, expected)
		}
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

	t.Run("malformed config errors", func(t *testing.T) {
		dir, remove := help.CreateTempDir(t, "home")
		defer remove()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Code did not panic")
			}
		}()

		help.WriteConfig(t, dir, `{"target": "my_host"`)
		// Should panic
		NewConfig(&help.TempHomeDir{HomeDir: dir})
	})

	t.Run("missing repo config means no files", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotfiles, removeDots := help.CreateTempDir(t, "dotfiles")
		defer removeDots()

		userConf := fmt.Sprintf(`{"target": "my_host", "dotfiles_root": "%v"}`, dotfiles)
		help.WriteConfig(t, home, userConf)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		if len(c.GetTargetFiles()) != 0 {
			t.Errorf("missing repo config should result in 0 files, got %v", len(c.GetTargetFiles()))
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

		targetFiles := c.GetTargetFiles()
		if len(targetFiles) != 1 {
			t.Errorf("Expected 1 file, got %v", len(targetFiles))
		}

		f := targetFiles[0]
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
		home, dotPath, remove := help.SetupFullConfig(t, "home", nil)
		defer remove()

		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		c.ManageFile("~/.some_config")
		c.ManageFile("~/.config/init.vim")
		err := c.Write()
		help.Ensure(t, err)

		actual := help.GetAllFiles(t, dotPath)
		expected := map[string]string{
			"dot_zshrc":       "~/.zshrc",
			"dot_some_config": "~/.some_config",
			"init.vim":        "~/.config/init.vim",
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("all files incorrect, got %v want %v", actual, expected)
		}
	})

	t.Run("add files to repo config with home substitution", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "home", nil)
		defer remove()

		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		_, err := c.ManageFile(filepath.Join(home, ".some_config"))
		help.Ensure(t, err)
		err = c.Write()
		help.Ensure(t, err)

		actual := help.GetAllFiles(t, dotPath)
		expected := map[string]string{
			"dot_zshrc":       "~/.zshrc",
			"dot_some_config": "~/.some_config",
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("all files incorrect, got %v want %v", actual, expected)
		}
	})

	t.Run("is valid file", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			}
		}`)

		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		if c.IsValidFile("dot_zshrc") != true {
			t.Fatalf("dot_zshrc should be a valid file")
		}

		if c.IsValidFile("invalid_file") != false {
			t.Fatalf("invalid_file should not be a valid file")
		}
	})

	t.Run("add files to repo already exists", func(t *testing.T) {
		// should setup a "dot_zshrc" file
		home, _, remove := help.SetupFullConfig(t, "home", nil)
		defer remove()

		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		_, err := c.ManageFile("~/subdir/.zshrc")
		help.ShouldError(t, err)
		want := `template name "dot_zshrc" already exists`
		if err.Error() != want {
			t.Fatalf("incorrect error, got %q want %q", err, want)
		}
	})

	t.Run("add file to target", func(t *testing.T) {
		home, dotPath, remove := help.SetupFullConfig(t, "home", nil)
		defer remove()

		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		err := c.AddToTarget("my_host", "dot_zshrc")
		help.Ensure(t, err)
		err = c.Write()
		if err != nil {
			t.Fatalf("error writing config, %v", err)
		}

		help.AssertTargetContents(t, dotPath, "my_host", []string{"dot_zshrc"})
	})

	t.Run("list all files", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc",
				"some_conf": "~/some_conf"
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		s := bytes.NewBufferString("")
		c.ListAllFiles(s)

		want := `dot_zshrc => ~/.zshrc
some_conf => ~/some_conf
`

		got := s.String()
		if got != want {
			t.Errorf("incorrect data printed, got %q want %q", got, want)
		}
	})

	t.Run("list target files", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc",
				"some_conf": "~/some_conf",
				"other_conf": "~/other_conf"
			},
			"renders": {
				"home": ["dot_zshrc", "some_conf"],
				"work": ["other_conf"]
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		s := bytes.NewBufferString("")
		c.ListTargetFiles(c.Target, s)

		want := `Target: home
dot_zshrc => ~/.zshrc
some_conf => ~/some_conf
`

		got := s.String()
		if got != want {
			t.Errorf("incorrect data printed, got %q want %q", got, want)
		}
	})

	t.Run("remove_file_from_target", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc",
				"some_conf": "~/some_conf"
			},
			"renders": {
				"home": ["dot_zshrc", "some_conf"]
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		err := c.RemoveFromTarget("home", "dot_zshrc")
		help.Ensure(t, err)
		err = c.Write()
		help.Ensure(t, err)
		help.AssertTargetContents(t, dotPath, "home", []string{"some_conf"})
	})

	t.Run("get_template_path_from_full_path", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		want := filepath.Join(dotPath, "templates", "dot_zshrc")
		got, err := c.GetTemplateFromFullPath(filepath.Join(home, ".zshrc"))
		help.Ensure(t, err)
		if got != want {
			t.Fatalf("incorrect template path, got %q want %q", got, want)
		}
	})

	t.Run("get_template_path_from_full_path_not_found", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		_, err := c.GetTemplateFromFullPath(filepath.Join(home, ".missing_config"))
		help.ShouldError(t, err)
	})
}
