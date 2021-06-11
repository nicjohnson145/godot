package controller

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"sort"

	"github.com/nicjohnson145/godot/internal/bootstrap"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/help"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/stretchr/testify/require"
)

const TARGET = "host1"

type testData struct {
	C       *Controller
	Home    string
	Dotpath string
	Remove  func()
}

func writeFile(t *testing.T, dotpath string, tmpl string, contents string) {
	t.Helper()

	err := ioutil.WriteFile(
		filepath.Join(dotpath, "templates", tmpl),
		[]byte(contents),
		0644,
	)
	require.NoError(t, err)
}

func writeFiles(t *testing.T, dotpath string, tmplData map[string]string) {
	t.Helper()

	for k, v := range tmplData {
		writeFile(t, dotpath, k, v)
	}
}

func setup(t *testing.T, conf string, tmplData map[string]string) testData {
	t.Helper()

	home, dotpath, remove := help.SetupDirectories(t, TARGET)
	help.WriteRepoConf(t, dotpath, conf)
	writeFiles(t, dotpath, tmplData)
	c := NewController(ControllerOpts{
		HomeDirGetter: &help.TempHomeDir{HomeDir: home},
		Repo:          repo.NoopRepo{},
		Runner:        &bootstrap.NoopRunner{},
	})

	return testData{
		C:       c,
		Home:    home,
		Dotpath: dotpath,
		Remove:  remove,
	}
}

func baseSetup(t *testing.T) testData {
	t.Helper()
	return setup(
		t,
		help.SAMPLE_CONFIG,
		map[string]string{
			"dot_zshrc": "dot_zshrc",
			"some_conf": "some_conf",
			"odd_conf":  "odd_conf",
		},
	)
}

func TestSync(t *testing.T) {

	assertZshRc := func(t *testing.T, obj testData) {
		help.AssertFileContents(t, filepath.Join(obj.Home, ".zshrc"), "dot_zshrc")
	}

	assertSomeConf := func(t *testing.T, obj testData) {
		help.AssertFileContents(t, filepath.Join(obj.Home, "some_conf"), "some_conf")
	}

	t.Run("happy_path", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		help.AssertDirectoryContents(t, obj.Home, []string{".config", "dotfiles"})

		err := obj.C.Sync(SyncOpts{Force: false})
		require.NoError(t, err)

		help.AssertDirectoryContents(
			t,
			obj.Home,
			[]string{
				".config",
				"dotfiles",
				"some_conf",
				".zshrc",
			},
		)

		assertZshRc(t, obj)
		assertSomeConf(t, obj)

		// Make sure our runner was called with the relavant configs
		want := [][]config.BootstrapImpl{
			{
				{Name: "apt", Item: config.BootstrapItem{Name: "ripgrep"}},
			},
		}
		r := obj.C.runner.(*bootstrap.NoopRunner)
		require.Equal(t, want, r.RunAllArgs)

	})

	t.Run("file_exists_not_symlink", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		help.Touch(t, filepath.Join(obj.Home, ".zshrc"))

		// Should error out
		err := obj.C.Sync(SyncOpts{Force: false})
		require.Error(t, err)

		// But some_conf should still be created
		assertSomeConf(t, obj)
	})

	t.Run("file_exists_not_symlink_force", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		help.Touch(t, filepath.Join(obj.Home, ".zshrc"))

		err := obj.C.Sync(SyncOpts{Force: true})
		require.NoError(t, err)

		assertZshRc(t, obj)
		assertSomeConf(t, obj)
	})

	t.Run("error_bootstrapping", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		obj.C.runner.(*bootstrap.NoopRunner).RunAllErr = errors.New("bad")

		err := obj.C.Sync(SyncOpts{})
		require.Error(t, err)

		// Symlinks should still be there
		assertZshRc(t, obj)
		assertSomeConf(t, obj)
	})
}

func TestImport(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		confPath := filepath.Join(obj.Home, ".new_conf")
		help.WriteData(t, confPath, "my new conf")
		err := obj.C.Import(confPath, "")
		require.NoError(t, err)
		require.Equal(
			t,
			config.StringMap{
				"dot_zshrc":    "~/.zshrc",
				"some_conf":    "~/some_conf",
				"odd_conf":     "/etc/odd_conf",
				"dot_new_conf": "~/.new_conf",
			},
			obj.C.config.GetRawContent().Files,
		)
	})

	t.Run("name_collision_errors", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		confPath := filepath.Join(obj.Home, "subfolder", "some_conf")
		help.WriteData(t, confPath, "my new conf")
		err := obj.C.Import(confPath, "")
		require.Error(t, err)
		require.Equal(
			t,
			config.StringMap{
				"dot_zshrc": "~/.zshrc",
				"some_conf": "~/some_conf",
				"odd_conf":  "/etc/odd_conf",
			},
			obj.C.config.GetRawContent().Files,
		)
	})

	t.Run("name_collision_with_as_allowed", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		confPath := filepath.Join(obj.Home, "subfolder", "some_conf")
		help.WriteData(t, confPath, "my new conf")
		err := obj.C.Import(confPath, "sub_some_conf")
		require.NoError(t, err)
		require.Equal(
			t,
			config.StringMap{
				"dot_zshrc":     "~/.zshrc",
				"some_conf":     "~/some_conf",
				"sub_some_conf": "~/subfolder/some_conf",
				"odd_conf":      "/etc/odd_conf",
			},
			obj.C.config.GetRawContent().Files,
		)
	})
}

func TestShowFilesEntry(t *testing.T) {

	showFiles := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "no_target",
			target: "",
			want: strings.Join([]string{
				"dot_zshrc => ~/.zshrc",
				" odd_conf => /etc/odd_conf",
				"some_conf => ~/some_conf",
			}, "\n") + "\n",
		},
		{
			name:   "named_target",
			target: "host3",
			want: strings.Join([]string{
				"dot_zshrc => ~/.zshrc",
			}, "\n") + "\n",
		},
		{
			name:   "current_target",
			target: config.CURRENT,
			want: strings.Join([]string{
				"dot_zshrc => ~/.zshrc",
				"some_conf => ~/some_conf",
			}, "\n") + "\n",
		},
	}
	for _, tc := range showFiles {
		t.Run(tc.name, func(t *testing.T) {
			obj := baseSetup(t)
			defer obj.Remove()

			buf := bytes.NewBufferString("")
			err := obj.C.ShowFilesEntry(tc.target, buf)
			require.NoError(t, err)
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestTargetAddFile(t *testing.T) {
	t.Run("no_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetAddFile("", []string{"odd_conf"})
		require.NoError(t, err)

		require.Equal(
			t,
			[]string{"dot_zshrc", "some_conf", "odd_conf"},
			obj.C.config.GetRawContent().Hosts[TARGET].Files,
		)
	})

	t.Run("with_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetAddFile("host2", []string{"odd_conf"})
		require.NoError(t, err)

		require.Equal(
			t,
			[]string{"some_conf", "odd_conf"},
			obj.C.config.GetRawContent().Hosts["host2"].Files,
		)
	})
}

func TestTargetRemoveFile(t *testing.T) {
	t.Run("no_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetRemoveFile("", []string{"dot_zshrc"})
		require.NoError(t, err)

		require.Equal(
			t,
			[]string{"some_conf"},
			obj.C.config.GetRawContent().Hosts[TARGET].Files,
		)
	})

	t.Run("with_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetRemoveFile("host2", []string{"some_conf"})
		require.NoError(t, err)

		require.Equal(
			t,
			[]string{},
			obj.C.config.GetRawContent().Hosts["host2"].Files,
		)
	})
}

func TestShowBootstrapEntry(t *testing.T) {
	showBootstraps := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "no target",
			target: "",
			want: strings.Join([]string{
				"  pyenv => brew, git",
				"ripgrep => apt, brew",
			}, "\n") + "\n",
		},
		{
			name:   "current target",
			target: config.CURRENT,
			want: strings.Join([]string{
				"ripgrep => apt, brew",
			}, "\n") + "\n",
		},
		{
			name:   "named target",
			target: "host4",
			want: strings.Join([]string{
				"pyenv => brew, git",
			}, "\n") + "\n",
		},
	}
	for _, tc := range showBootstraps {
		t.Run(tc.name, func(t *testing.T) {
			obj := baseSetup(t)
			defer obj.Remove()

			buf := bytes.NewBufferString("")
			err := obj.C.ShowBootstrapsEntry(tc.target, buf)
			require.NoError(t, err)

			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestAddBootstrapItem(t *testing.T) {
	t.Run("add_to_existing", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddBootstrapItem("ripgrep", "git", "http://repo.com", "~/.ripgrep")
		require.NoError(t, err)

		item, hasGit := obj.C.config.GetRawContent().Bootstraps["ripgrep"]["git"]
		require.Equal(t, hasGit, true, "git manager not present")
		require.Equal(t, config.BootstrapItem{Name: "http://repo.com", Location: "~/.ripgrep"}, item)
	})

	t.Run("add_new", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddBootstrapItem("new_package", "brew", "new_package", "")
		require.NoError(t, err)

		item, hasNew := obj.C.config.GetRawContent().Bootstraps["new_package"]
		require.Equal(t, hasNew, true, "new_package not added")
		require.Equal(t, config.Bootstrap{"brew": {Name: "new_package"}}, item)
	})

	t.Run("invalid manager", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddBootstrapItem("new_package", "not_a_manager", "new_package", "")
		require.Error(t, err)
	})
}

func TestAddTargetBootstrap(t *testing.T) {
	t.Run("no_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddTargetBootstrap("", []string{"pyenv"})
		require.NoError(t, err)

		want := []string{"ripgrep", "pyenv"}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts[TARGET].Bootstraps)
	})

	t.Run("current_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddTargetBootstrap(config.CURRENT, []string{"pyenv"})
		require.NoError(t, err)

		want := []string{"ripgrep", "pyenv"}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts[TARGET].Bootstraps)
	})

	t.Run("specific_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddTargetBootstrap("host3", []string{"pyenv"})
		require.NoError(t, err)

		want := []string{"pyenv"}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts["host3"].Bootstraps)
	})

	t.Run("not_a_bootstrap", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.AddTargetBootstrap("", []string{"not_a_bootstrap"})
		require.Error(t, err)

		want := []string{"ripgrep"}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts[TARGET].Bootstraps)
	})
}

func TestRemoveBootstrapTarget(t *testing.T) {
	t.Run("no_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.RemoveTargetBootstrap("", []string{"ripgrep"})
		require.NoError(t, err)

		want := []string{}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts[TARGET].Bootstraps)
	})

	t.Run("current_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.RemoveTargetBootstrap(config.CURRENT, []string{"ripgrep"})
		require.NoError(t, err)

		want := []string{}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts[TARGET].Bootstraps)
	})

	t.Run("specific_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.RemoveTargetBootstrap("host2", []string{"pyenv"})
		require.NoError(t, err)

		want := []string{"ripgrep"}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts["host2"].Bootstraps)
	})

	t.Run("not_a_bootstrap", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.RemoveTargetBootstrap("", []string{"not_a_bootstrap"})
		require.Error(t, err)

		want := []string{"ripgrep"}
		require.Equal(t, want, obj.C.config.GetRawContent().Hosts[TARGET].Bootstraps)
	})
}

func TestAllTarget(t *testing.T) {
	t.Run("files", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		beforeHosts := []string{}
		for h, host := range obj.C.config.GetRawContent().Hosts {
			t.Logf("Checking host pre add %v", h)
			beforeHosts = append(beforeHosts, h)
			help.AssertStringNotInSlice(t, "new_file", host.Files)
		}
		sort.Strings(beforeHosts)

		err := obj.C.Import("~/new_file", "new_file")
		require.NoError(t, err)

		err = obj.C.TargetAddFile("ALL", []string{"new_file"})
		require.NoError(t, err)

		afterAddHosts := []string{}
		for h, host := range obj.C.config.GetRawContent().Hosts {
			t.Logf("Checking host post add %v", h)
			afterAddHosts = append(afterAddHosts, h)
			help.AssertStringInSlice(t, "new_file", host.Files)
		}
		sort.Strings(afterAddHosts)
		require.Equal(t, afterAddHosts, beforeHosts)

		// Now remove the file from all hosts
		err = obj.C.TargetRemoveFile("ALL", []string{"new_file"})
		require.NoError(t, err)

		afterRemoveHosts := []string{}
		for h, host := range obj.C.config.GetRawContent().Hosts {
			t.Logf("Checking host post remove %v", h)
			afterRemoveHosts = append(afterRemoveHosts, h)
			help.AssertStringNotInSlice(t, "new_file", host.Files)
		}
		sort.Strings(afterRemoveHosts)
		require.Equal(t, afterRemoveHosts, beforeHosts)
	})

	t.Run("bootstraps", func(t *testing.T) {
		const name string = "new_bs"

		obj := baseSetup(t)
		defer obj.Remove()

		beforeHosts := []string{}
		for h, host := range obj.C.config.GetRawContent().Hosts {
			t.Logf("Checking host pre add %v", h)
			beforeHosts = append(beforeHosts, h)
			help.AssertStringNotInSlice(t, "new_file", host.Bootstraps)
		}
		sort.Strings(beforeHosts)

		err := obj.C.AddBootstrapItem(name, "apt", name, "")

		err = obj.C.AddTargetBootstrap("ALL", []string{name})
		require.NoError(t, err)

		afterAddHosts := []string{}
		for h, host := range obj.C.config.GetRawContent().Hosts {
			t.Logf("Checking host post add %v", h)
			afterAddHosts = append(afterAddHosts, h)
			help.AssertStringInSlice(t, name, host.Bootstraps)
		}
		sort.Strings(afterAddHosts)
		require.Equal(t, afterAddHosts, beforeHosts)

		err = obj.C.RemoveTargetBootstrap("ALL", []string{name})
		require.NoError(t, err)

		afterRemoveHosts := []string{}
		for h, host := range obj.C.config.GetRawContent().Hosts {
			t.Logf("Checking host post remove %v", h)
			afterRemoveHosts = append(afterRemoveHosts, h)
			help.AssertStringNotInSlice(t, name, host.Bootstraps)
		}
		sort.Strings(afterRemoveHosts)
		require.Equal(t, afterRemoveHosts, beforeHosts)
	})
}
