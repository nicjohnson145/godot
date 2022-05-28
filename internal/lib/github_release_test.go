package lib

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"sort"
	"testing"
)

// Best tests are real tests? question mark?
func TestGithubReleaseExecute(t *testing.T) {
	checkFiles := func(t *testing.T, dir string, want []string) {
		t.Helper()

		ioFiles, err := ioutil.ReadDir(dir)
		require.NoError(t, err)

		files := []string{}
		for _, f := range ioFiles {
			files = append(files, f.Name())
		}

		sort.Strings(files)

		require.Equal(t, want, files)
	}

	ghuser, userOk := os.LookupEnv("GITHUB_USER")
	ghpat, patOk := os.LookupEnv("GITHUB_PAT")

	if !userOk || !patOk {
		t.Skip("Either GITHUB_USER or GITHUB_PAT isn't set. Skipping to avoid flaky tests")
	}

	t.Run("single_binary", func(t *testing.T) {
		restore, noFatal := NoFatals()
		defer noFatal(t)
		defer restore()

		dir := t.TempDir()

		g := GithubRelease{
			Name:         "godot",
			Repo:         "nicjohnson145/godot",
			Type:         Binary,
			Tag:          "v2.4.1",
			LinuxPattern: "godot_linux_amd64",
			MacPattern:   "godot_darwin_amd64",
		}
		g.Execute(UserConfig{
			BinaryDir:  dir,
			GithubUser: ghuser,
			GithubAuth: BasicAuth(ghuser, ghpat),
		})

		checkFiles(t, dir, []string{"godot", "godot-v2.4.1"})
	})

	t.Run("tarball", func(t *testing.T) {
		restore, noFatal := NoFatals()
		defer noFatal(t)
		defer restore()

		dir := t.TempDir()

		g := GithubRelease{
			Name:         "rg",
			Repo:         "BurntSushi/ripgrep",
			Type:         TarGzDir,
			Tag:          "13.0.0",
			LinuxPattern: ".*unknown-linux-musl.*",
			MacPattern:   ".*apple-darwin.*",
			Path:         "rg",
		}
		g.Execute(UserConfig{
			BinaryDir:  dir,
			GithubUser: ghuser,
			GithubAuth: BasicAuth(ghuser, ghpat),
		})

		checkFiles(t, dir, []string{"rg", "rg-13.0.0"})
	})
}
