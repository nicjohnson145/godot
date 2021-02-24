package controller

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/help"
	"github.com/nicjohnson145/godot/internal/repo"
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
	help.Ok(t, err)
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
		help.Ok(t, err)

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
	})

	t.Run("file_exists_not_symlink", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		help.Touch(t, filepath.Join(obj.Home, ".zshrc"))

		// Should error out
		err := obj.C.Sync(SyncOpts{Force: false})
		help.ShouldError(t, err)

		// TODO
		// But some_conf should still be created
		// assertSomeConf(t, obj)
	})

	t.Run("file_exists_not_symlink_force", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		help.Touch(t, filepath.Join(obj.Home, ".zshrc"))

		err := obj.C.Sync(SyncOpts{Force: true})
		help.Ok(t, err)

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
		help.Ok(t, err)
		help.Equals(
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
		help.ShouldError(t, err)
		help.Equals(
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
		help.Ok(t, err)
		help.Equals(
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
			name: "no_target",
			target: "",
			want: strings.Join([]string{
				"dot_zshrc => ~/.zshrc",
				" odd_conf => /etc/odd_conf",
				"some_conf => ~/some_conf",
			}, "\n") + "\n",
		},
		{
			name: "no_target",
			target: TARGET,
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
			help.Ok(t, err)
			help.Equals(t, tc.want, buf.String())
		})
	}
}

func TestTargetAddFile(t *testing.T) {
	t.Run("no_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetAddFile("", []string{"odd_conf"})
		help.Ok(t, err)

		help.Equals(
			t,
			[]string{"dot_zshrc", "some_conf", "odd_conf"},
			obj.C.config.GetRawContent().Hosts[TARGET].Files,
		)
	})

	t.Run("with_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetAddFile("host2", []string{"odd_conf"})
		help.Ok(t, err)

		help.Equals(
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
		help.Ok(t, err)

		help.Equals(
			t,
			[]string{"some_conf"},
			obj.C.config.GetRawContent().Hosts[TARGET].Files,
		)
	})

	t.Run("with_target", func(t *testing.T) {
		obj := baseSetup(t)
		defer obj.Remove()

		err := obj.C.TargetRemoveFile("host2", []string{"some_conf"})
		help.Ok(t, err)

		help.Equals(
			t,
			[]string{},
			obj.C.config.GetRawContent().Hosts["host2"].Files,
		)
	})
}
