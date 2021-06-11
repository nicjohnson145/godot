package config

import (
	"bytes"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"reflect"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/stretchr/testify/require"
)

const TARGET = "host1"

func setup(t *testing.T, conf string) (*Config, func()) {
	home, dotpath, remove := help.SetupDirectories(t, TARGET)
	help.WriteRepoConf(t, dotpath, conf)
	c := NewConfig(&help.TempHomeDir{HomeDir: home})
	return c, remove
}

func baseSetup(t *testing.T) (*Config, func()) {
	return setup(t, help.SAMPLE_CONFIG)
}

func getErrFunc(t *testing.T, shouldError bool) func(require.TestingT, error, ...interface{}) {
	t.Helper()
	if shouldError {
		return require.Error
	} else {
		return require.NoError
	}
}

func TestFiles(t *testing.T) {
	t.Run("GetAllFiles_empty", func(t *testing.T) {
		c, remove := setup(t, "{}")
		defer remove()

		got := c.GetAllFiles()
		want := StringMap{}
		require.Equal(t, want, got)
	})

	t.Run("GetAllFiles", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		got := c.GetAllFiles()
		want := StringMap{
			"dot_zshrc": "~/.zshrc",
			"some_conf": "~/some_conf",
			"odd_conf":  "/etc/odd_conf",
		}
		require.Equal(t, want, got)
	})

	getFilesForTarget := []struct {
		name string
		host string
		want StringMap
	}{
		{
			name: "missing_host",
			host: "not_a_host",
			want: StringMap{},
		},
		{
			name: "valid_host",
			host: "host1",
			want: StringMap{"dot_zshrc": "~/.zshrc", "some_conf": "~/some_conf"},
		},
	}
	for _, tc := range getFilesForTarget {
		t.Run("GetFilesForTarget_"+tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			got := c.GetFilesForTarget(tc.host)
			require.Equal(t, tc.want, got)
		})
	}

	addFile := []struct {
		name        string
		template    string
		destination string
		shouldError bool
		initial     string
		want        StringMap
	}{
		{
			name:        "happy_path",
			template:    "new_conf",
			destination: "~/new_conf",
			initial:     "{}",
			want:        StringMap{"new_conf": "~/new_conf"},
		},
		{
			name:        "no_template",
			template:    "",
			destination: "~/.new_conf",
			initial:     "{}",
			want:        StringMap{"dot_new_conf": "~/.new_conf"},
		},
		{
			name:        "template_exists",
			template:    "",
			destination: "~/.new_conf",
			shouldError: true,
			initial:     `{"files": {"dot_new_conf": "~/.not_this_file"}}`,
			want:        StringMap{"dot_new_conf": "~/.not_this_file"},
		},
	}
	for _, tc := range addFile {
		t.Run("AddFile_"+tc.name, func(t *testing.T) {
			c, remove := setup(t, tc.initial)
			defer remove()

			_, err := c.AddFile(tc.template, tc.destination)
			getErrFunc(t, tc.shouldError)(t, err)
			got := c.GetAllFiles()
			require.Equal(t, tc.want, got)
		})
	}

	addToTarget := []struct {
		name        string
		shouldError bool
		template    string
		want        StringMap
	}{
		{
			name:        "happy_path",
			shouldError: false,
			template:    "dot_zshrc",
			want:        StringMap{"dot_zshrc": "~/.zshrc"},
		},
		{
			name:        "unknown_template",
			shouldError: true,
			template:    "not_a_template",
			want:        StringMap{},
		},
	}
	for _, tc := range addToTarget {
		t.Run("AddTargetFile_"+tc.name, func(t *testing.T) {
			initial := `{"files": {"dot_zshrc": "~/.zshrc"}}`
			c, remove := setup(t, initial)
			defer remove()

			err := c.AddTargetFile(TARGET, tc.template)
			getErrFunc(t, tc.shouldError)(t, err)

			got := c.GetFilesForTarget(TARGET)
			require.Equal(t, tc.want, got)
		})
	}

	removeFromTarget := []struct {
		name        string
		shouldError bool
		target      string
		template    string
		want        StringMap
	}{
		{
			name:        "happy_path",
			shouldError: false,
			target:      "host1",
			template:    "dot_zshrc",
			want:        StringMap{"some_conf": "~/some_conf"},
		},
		{
			name:        "unknown_target",
			shouldError: true,
			target:      "not_a_target",
			template:    "dot_zshrc",
			want:        StringMap{},
		},
		{
			name:        "unknown_template",
			shouldError: true,
			target:      "host1",
			template:    "not_a_template",
			want:        StringMap{"dot_zshrc": "~/.zshrc", "some_conf": "~/some_conf"},
		},
	}
	for _, tc := range removeFromTarget {
		t.Run("RemoveTargetFile_"+tc.name, func(t *testing.T) {
			initial := `{
				"files": {
					"dot_zshrc": "~/.zshrc",
					"some_conf": "~/some_conf"
				},
				"hosts": {
					"host1": {
						"files": ["dot_zshrc", "some_conf"]
					}
				}
			}`
			c, remove := setup(t, initial)
			defer remove()

			err := c.RemoveTargetFile(tc.target, tc.template)
			getErrFunc(t, tc.shouldError)(t, err)

			got := c.GetFilesForTarget(tc.target)
			require.Equal(t, tc.want, got)
		})
	}

	t.Run("GetAllTemplateNames", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		got := c.GetAllTemplateNames()
		sort.Strings(got)
		want := []string{"dot_zshrc", "some_conf", "odd_conf"}
		sort.Strings(want)
		require.Equal(t, want, got)
	})

	t.Run("GetAllTemplateNamesForTarget", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		got := c.GetAllTemplateNamesForTarget("host1")
		sort.Strings(got)
		want := []string{"dot_zshrc", "some_conf"}
		sort.Strings(want)
		require.Equal(t, want, got)
	})

	t.Run("GetTemplateFromFullPath", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		got, err := c.GetTemplateFromFullPath(filepath.Join(c.Home, ".zshrc"))
		require.NoError(t, err)

		want := filepath.Join(c.DotfilesRoot, "templates", "dot_zshrc")
		require.Equal(t, want, got)
	})

	t.Run("GetTemplateFromFullPath_not_found", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		_, err := c.GetTemplateFromFullPath(filepath.Join(c.Home, "not_a_file"))
		require.Error(t, err)
	})

	t.Run("ListAllFiles", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		b := bytes.NewBufferString("")
		err := c.ListAllFiles(b)
		require.NoError(t, err)

		want := strings.Join(
			[]string{
				"dot_zshrc => ~/.zshrc",
				" odd_conf => /etc/odd_conf",
				"some_conf => ~/some_conf",
			},
			"\n",
		) + "\n"
		require.Equal(t, want, b.String())
	})

	t.Run("ListTargetFiles", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		b := bytes.NewBufferString("")
		err := c.ListTargetFiles("host1", b)
		require.NoError(t, err)

		want := strings.Join(
			[]string{
				"dot_zshrc => ~/.zshrc",
				"some_conf => ~/some_conf",
			},
			"\n",
		) + "\n"
		require.Equal(t, want, b.String())
	})
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
				"apt":  {Name: "ripgrep"},
			},
			"pyenv": {
				"brew": {Name: "pyenv"},
				"git": {
					Name:     "https://github.com/pyenv/pyenv.git",
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
					"apt":  {Name: "ripgrep"},
				},
				"pyenv": {
					"brew": {Name: "pyenv"},
					"git": {
						Name:     "https://github.com/pyenv/pyenv.git",
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
					"apt":  {Name: "ripgrep"},
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
					"apt":  {Name: "ripgrep"},
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
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			got := c.content.Hosts[tc.host].Bootstraps
			assertBootstrap(t, got, tc.want)
		})
	}

	removeBootstrapTarget := []struct {
		name        string
		shouldError bool
		host        string
		item        string
		want        []string
	}{
		{
			name:        "unknown_target",
			shouldError: true,
			host:        "not_a_host",
			item:        "not_an_item",
		},
		{
			name:        "unknown_item",
			shouldError: true,
			host:        "host1",
			item:        "not_an_item",
			want:        []string{"ripgrep"},
		},
		{
			name:        "valid_remove_still_items_left",
			shouldError: false,
			host:        "host2",
			item:        "pyenv",
			want:        []string{"ripgrep"},
		},
		{
			name:        "valid_remove_last_item",
			shouldError: false,
			host:        "host1",
			item:        "ripgrep",
			want:        []string{},
		},
	}
	for _, tc := range removeBootstrapTarget {
		t.Run("RemoveTargetBootstrap_"+tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			err := c.RemoveTargetBootstrap(tc.host, tc.item)
			getErrFunc(t, tc.shouldError)(t, err)

			host, _ := c.content.Hosts[tc.host]
			assertBootstrap(t, host.Bootstraps, tc.want)
		})
	}

	getBSTagetData := []struct {
		name   string
		target string
		want   []string
	}{
		{
			name:   "no_target",
			target: "",
			want:   []string{"ripgrep"},
		},
		{
			name:   "host2",
			target: "host2",
			want:   []string{"ripgrep", "pyenv"},
		},
	}
	for _, tc := range getBSTagetData {
		t.Run("GetBootstrapTargetsForTarget_"+tc.name, func(t *testing.T) {
			c, remove := baseSetup(t)
			defer remove()

			got := c.GetBootstrapTargetsForTarget(tc.target)
			require.Equal(t, tc.want, got)
		})
	}

	t.Run("ListAllBootstraps", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		b := bytes.NewBufferString("")
		err := c.ListAllBootstraps(b)
		require.NoError(t, err)

		want := strings.Join(
			[]string{
				"  pyenv => brew, git",
				"ripgrep => apt, brew",
			},
			"\n",
		) + "\n"
		require.Equal(t, want, b.String())
	})

	t.Run("ListBootstrapForTarget", func(t *testing.T) {
		c, remove := baseSetup(t)
		defer remove()

		b := bytes.NewBufferString("")
		err := c.ListBootstrapsForTarget(b, "host1")
		require.NoError(t, err)

		want := strings.Join(
			[]string{
				"ripgrep => apt, brew",
			},
			"\n",
		) + "\n"
		require.Equal(t, want, b.String())
	})

	t.Run("GetRelevantBootstrapImpls_noErrors", func(t *testing.T) {
		c, remove := setup(t, `{
			"files": {},
			"bootstraps": {
				"ripgrep": {
					"apt": {
						"name": "ripgrep"
					},
					"git": {
						"name": "https://github.com/ripgrep/ripgrep.git",
						"location": "~/.ripgrep"
					}
				},
				"pyenv": {
					"git": {
						"name": "https://github.com/pyenv/pyenv.git",
						"location": "~/.pyenv"
					},
					"yum": {
						"name": "broke_2"
					}
				},
				"broke_1": {
					"brew": {
						"name": "broke_1"
					}
				},
				"broke_2": {
					"brew": {
						"name": "broke_2"
					},
					"yum": {
						"name": "broke_2"
					}
				}
			},
			"hosts": {
				"host1": {
					"files": [],
					"bootstraps": ["ripgrep", "pyenv"]
				},
				"host2": {
					"files": [],
					"bootstraps": ["broke_1", "broke_2"]
				}
			}
		}`)
		defer remove()

		got, err := c.GetRelevantBootstrapImpls("host1")
		require.NoError(t, err)

		want := []BootstrapImpl{
			{Name: "apt", Item: BootstrapItem{Name: "ripgrep"}},
			{Name: "git", Item: BootstrapItem{Name: "https://github.com/pyenv/pyenv.git", Location: filepath.Join(c.Home, ".pyenv")}},
		}

		require.Equal(t, want, got)
	})

	t.Run("GetRelevantBootstrapImpls_errors", func(t *testing.T) {
		// Should show all errors
		c, remove := setup(t, `{
			"files": {},
			"bootstraps": {
				"ripgrep": {
					"apt": {
						"name": "ripgrep"
					},
					"git": {
						"name": "https://github.com/ripgrep/ripgrep.git",
						"location": "~/.ripgrep"
					}
				},
				"pyenv": {
					"git": {
						"name": "https://github.com/pyenv/pyenv.git",
						"location": "~/.pyenv"
					},
					"yum": {
						"name": "broke_2"
					}
				},
				"broke_1": {
					"brew": {
						"name": "broke_1"
					}
				},
				"broke_2": {
					"brew": {
						"name": "broke_2"
					},
					"yum": {
						"name": "broke_2"
					}
				}
			},
			"hosts": {
				"host1": {
					"files": [],
					"bootstraps": ["ripgrep", "pyenv"]
				},
				"host2": {
					"files": [],
					"bootstraps": ["broke_1", "broke_2"]
				}
			}
		}`)
		defer remove()

		_, err := c.GetRelevantBootstrapImpls("host2")
		require.Error(t, err)

		got := err.Error()
		want := strings.Join([]string{
			"2 errors occurred:",
			"\t* No suitable manager found for broke_1, broke_1's available managers are brew",
			"\t* No suitable manager found for broke_2, broke_2's available managers are brew, yum",
		}, "\n") + "\n\n"
		require.Equal(t, want, got)
	})
}
