package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"reflect"

	"github.com/nicjohnson145/godot/internal/help"
)

func setup (t *testing.T, conf string) (*Config, func()) {
	home, dotpath, remove := help.SetupDirectories(t, "host1")
	help.WriteRepoConf(t, dotpath, conf)
	c := NewConfig(&help.TempHomeDir{HomeDir: home})
	return c, remove
}

func baseSetup(t *testing.T) (*Config, func()) {
	return setup(t, `{
		"files": {
			"dot_zshrc": "~/.zshrc",
			"some_conf": "~/some_conf",
			"odd_conf": "/etc/odd_conf"
		},
		"bootstraps": {
			"ripgrep": {
				"brew": {
					"name": "rip-grep"
				},
				"apt": {
					"name": "ripgrep"
				}
			},
			"pyenv": {
				"brew": {
					"name": "pyenv"
				},
				"git": {
					"name": "https://github.com/pyenv/pyenv.git",
					"location": "~/.pyenv"
				}
			}
		},
		"hosts": {
			"host1": {
				"files": ["dot_zshrc", "some_conf"],
				"bootstraps": ["ripgrep"]
			},
			"host2": {
				"files": ["some_conf"],
				"bootstraps": ["ripgrep", "pyenv"]
			}
		}
	}`)
}

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

	t.Run("missing repo config can be added to", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotfiles, removeDots := help.CreateTempDir(t, "dotfiles")
		defer removeDots()

		userConf := fmt.Sprintf(`{"target": "my_host", "dotfiles_root": "%v"}`, dotfiles)
		help.WriteConfig(t, home, userConf)

		c := NewConfig(&help.TempHomeDir{HomeDir: home})

		_, err := c.ManageFile(filepath.Join(home, ".some_conf"))
		help.Ensure(t, err)

		err = c.AddToTarget("my_target", "dot_some_conf")
		help.Ensure(t, err)

		err = c.Write()
		help.Ensure(t, err)

		help.AssertAllFiles(t, dotfiles, map[string]string{"dot_some_conf": "~/.some_conf"})
		help.AssertTargetContents(t, dotfiles, "my_target", []string{"dot_some_conf"})
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

	t.Run("get_template_names_from_target", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc",
				"other_file": "~/other_file"
			},
			"renders": {
				"home": ["dot_zshrc"]
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		got := c.GetTemplatesNamesForTarget("home")
		want := []string{"dot_zshrc"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("incorrect files, got %q want %q", got, want)
		}
	})

	t.Run("get_all_template_names", func(t *testing.T) {
		home, dotPath, remove := help.SetupDirectories(t, "home")
		defer remove()

		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc",
				"other_file": "~/other_file"
			},
			"renders": {
				"home": ["dot_zshrc"]
			}
		}`)
		c := NewConfig(&help.TempHomeDir{HomeDir: home})
		got := c.GetAllTemplateNames()
		sort.Strings(got)
		want := []string{"dot_zshrc", "other_file"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("incorrect files, got %q want %q", got, want)
		}
	})
}

func TestFiles(t *testing.T) {
	t.Run("GetAllFiles_empty", func(t *testing.T) {
		c, remove := setup(t, "{}")
		defer remove()

		got := c.GetAllFiles()
		want := map[string]string{}
		help.Equals(t, want, got)
	})

	t.Run("GetAllFiles", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		got := c.GetAllFiles()
		want := map[string]string{
			"dot_zshrc": "~/.zshrc",
			"some_conf": "~/some_conf",
			"odd_conf": "/etc/odd_conf",
		}
		help.Equals(t, want, got)
	})

	getFilesForTarget := []struct{
		name string
		host string
		want map[string]string
	}{
		{
			name: "missing_host",
			host: "not_a_host",
			want: map[string]string{},
		},
		{
			name: "valid_host",
			host: "host1",
			want: map[string]string{"dot_zshrc": "~/.zshrc", "some_conf": "~/some_conf"},
		},
	}
	for _, tc := range getFilesForTarget{
		t.Run("GetFilesForTarget_" + tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			got := c.GetFilesForTarget(tc.host)
			help.Equals(t, tc.want, got)
		})
	}
}

func TestBootstrapping(t *testing.T) {
	assertBootstrap := func(t *testing.T, got []string, want []string) {
		t.Helper()

		if len(got) == 0 && len(want) == 0 {
			return
		}

		sort.Strings(got)

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bootstraps not equal, got %q want %q", got, want)
		}
	}

	t.Run("GetAllBootstraps_empty", func(t *testing.T) {
		c, remove := setup(t, "{}")
		defer remove()
		got := c.GetAllBootstraps()

		if len(got) != 0 {
			t.Fatalf("bootstrap map should be empty, got %d", len(got))
		}
	})

	t.Run("GetAllBootstraps", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()
		got := c.GetAllBootstraps()

		want := map[string]Bootstrap{
			"ripgrep": {
				"brew": {Name: "rip-grep"},
				"apt": {Name: "ripgrep"},
			},
			"pyenv": {
				"brew": {Name: "pyenv"},
				"git": {
					Name: "https://github.com/pyenv/pyenv.git",
					Location: "~/.pyenv",
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("All bootstraps not equal, got %q want %q", got, want)
		}
	})

	getTargetTests := []struct {
		name string
		host string
		want map[string]Bootstrap
	}{
		{
			name: "host2",
			host: "host2",
			want: map[string]Bootstrap{
				"ripgrep": {
					"brew": {Name: "rip-grep"},
					"apt": {Name: "ripgrep"},
				},
				"pyenv": {
					"brew": {Name: "pyenv"},
					"git": {
						Name: "https://github.com/pyenv/pyenv.git",
						Location: "~/.pyenv",
					},
				},
			},
		},
		{
			name: "host1",
			host: "",
			want: map[string]Bootstrap{
				"ripgrep": {
					"brew": {Name: "rip-grep"},
					"apt": {Name: "ripgrep"},
				},
			},
		},
		{
			name: "not_valid",
			host: "not_valid",
			want: map[string]Bootstrap{},
		},
	}
	for _, tc := range getTargetTests {
		t.Run("GetBootstrapsForTarget_"+tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			got := c.GetBootstrapsForTarget(tc.host)

			if len(tc.want) == 0 && len(got) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("bootstraps for %q not equal, got %q want %q", tc.host, got, tc.want)
			}
		})
	}

	addBootstrapItem := []struct {
		name    string
		initial string
		want    map[string]Bootstrap
	}{
		{
			name:    "empty_conf",
			initial: "{}",
			want: map[string]Bootstrap{
				"ripgrep": {
					"apt": {Name: "ripgrep"},
				},
			},
		},
		{
			name:    "bootstrap_exists",
			initial: `{"bootstraps": {}}`,
			want: map[string]Bootstrap{
				"ripgrep": {
					"apt": {Name: "ripgrep"},
				},
			},
		},
		{
			name:    "item_exists_different_manager",
			initial: `{"bootstraps": {"ripgrep": {"brew": {"name": "rip-grep"}}}}`,
			want: map[string]Bootstrap{
				"ripgrep": {
					"apt": {Name: "ripgrep"},
					"brew": {Name: "rip-grep"},
				},
			},
		},
		{
			name:    "item_exists_same_manager",
			initial: `{"bootstraps": {"ripgrep": {"apt": {"name": "rip-grep"}}}}`,
			want: map[string]Bootstrap{
				"ripgrep": {
					"apt": {Name: "ripgrep"},
				},
			},
		},
	}
	for _, tc := range addBootstrapItem {
		t.Run("AddBootstrapItem_"+tc.name, func(t *testing.T) {
			c, remove := setup(t, tc.initial)
			defer remove()

			c.AddBootstrapItem("ripgrep", "apt", "ripgrep", "")
			got := c.GetAllBootstraps()

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("all bootstraps not equal, got %q want %q", got, tc.want)
			}
		})
	}

	addBootstrapTarget := []struct {
		name        string
		shouldError bool
		host        string
		item        string
		want        []string
	}{
		{
			name:        "bootstrap_item_missing",
			shouldError: true,
			host:        "host1",
			item:        "not_an_item",
		},
		{
			name:        "target_doesnt_exist",
			shouldError: false,
			host:        "host3",
			item:        "ripgrep",
			want:        []string{"ripgrep"},
		},
		{
			name:        "target_exists",
			shouldError: false,
			host:        "host1",
			item:        "pyenv",
			want:        []string{"pyenv", "ripgrep"},
		},
	}
	for _, tc := range addBootstrapTarget {
		t.Run("add_bootstrap_target_"+tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			err := c.AddTargetBootstrap(tc.host, tc.item)
			if tc.shouldError {
				help.ShouldError(t, err)
				return
			} else {
				help.Ensure(t, err)
			}

			got := c.content.Hosts[tc.host].Bootstraps
			assertBootstrap(t, got, tc.want)
		})
	}

	removeBootstrapTarget := []struct {
		name            string
		shouldError     bool
		host            string
		item            string
		want            []string
	}{
		{
			name:        "unknown_target",
			shouldError: true,
			host:        "not_a_host",
			item:        "not_an_item",
		},
		{
			name:            "valid_remove_still_items_left",
			shouldError:     false,
			host:            "host2",
			item:            "pyenv",
			want:            []string{"ripgrep"},
		},
		{
			name:            "valid_remove_last_item",
			shouldError:     false,
			host:            "host1",
			item:            "ripgrep",
			want:            []string{},
		},
	}
	for _, tc := range removeBootstrapTarget {
		t.Run("remove_bootstrap_target_"+tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			err := c.RemoveTargetBootstrap(tc.host, tc.item)
			if tc.shouldError {
				help.ShouldError(t, err)
				return
			} else {
				help.Ensure(t, err)
			}

			host, _ := c.content.Hosts[tc.host]
			assertBootstrap(t, host.Bootstraps, tc.want)
		})
	}
}
