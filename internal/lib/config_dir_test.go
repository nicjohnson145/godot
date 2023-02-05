package lib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sort"

	"github.com/stretchr/testify/require"
)

func buildDirectoryStructure(t *testing.T, structure map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for path, content := range structure {
		// If it ends with a slash, make an empty directory
		if strings.HasSuffix(path, "/") {
			require.NoError(t, os.MkdirAll(filepath.Join(root, path), 0755))
			continue
		}

		// Otherwise write a file with that name and content
		containingDir := filepath.Dir(path)
		require.NoError(t, os.MkdirAll(filepath.Join(root, containingDir), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(root, path), []byte(content), 0644))
	}

	return root
}

func TestGetFiles(t *testing.T) {
	tmp := buildDirectoryStructure(t, map[string]string{
		"templates/some-config/top-file": "top-file",
		"templates/some-config/some-sub-dir/some-file": "Hello World",
		"output/": "",
		"home/": "",
	})

	userConf := UserConfig{
		CloneLocation: tmp,
		HomeDir:       filepath.Join(tmp, "home"),
		BuildLocation: filepath.Join(tmp, "output"),
	}

	confDir := ConfigDir{
		Name: "my-config",
		DirName: "some-config",
		Destination: "~/.config/some-config",
	}

	got, err := confDir.getFiles(userConf)
	require.NoError(t, err)
	sort.Strings(got)
	require.Equal(
		t,
		[]string{
			filepath.Join("some-config", "some-sub-dir", "some-file"),
			filepath.Join("some-config", "top-file"),
		},
		got,
	)
}

func TestExecute(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		defer cleanFuncsMap(t)

		root := buildDirectoryStructure(t, map[string]string{
			"templates/some-config/top-file": "Hello World",
			"templates/some-config/some-sub-dir/some-file": "Hello {{ .Target }}",
			"output/": "",
			"home/": "",
		})

		userConf := UserConfig{
			CloneLocation: root,
			HomeDir: filepath.Join(root, "home"),
			BuildLocation: filepath.Join(root, "output"),
		}

		confDir := ConfigDir{
			Name: "my-config",
			DirName: "some-config",
			Destination: "~/.config/some-config",
		}

		require.NoError(t, confDir.Execute(userConf, SyncOpts{}, GodotConfig{}))
		requireContents(t, filepath.Join(root, "home", ".config", "some-config", "top-file"), "Hello World")
		requireContents(t, filepath.Join(root, "home", ".config", "some-config", "some-sub-dir", "some-file"), "Hello {{ .Target }}")
	})
}
