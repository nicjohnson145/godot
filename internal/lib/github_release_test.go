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
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		dir := t.TempDir()

		g := GithubRelease{
			Name:         "godot",
			Repo:         "nicjohnson145/godot",
			IsArchive:    false,
			Tag:          "v2.4.1",
			LinuxPattern: "godot_linux_amd64",
			MacPattern:   "godot_darwin_amd64",
		}
		g.Execute(UserConfig{
			BinaryDir:  dir,
			GithubUser: ghuser,
			GithubAuth: BasicAuth(ghuser, ghpat),
		}, SyncOpts{})

		checkFiles(t, dir, []string{"godot", "godot-v2.4.1"})
	})

	t.Run("tarball", func(t *testing.T) {
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		dir := t.TempDir()

		g := GithubRelease{
			Name:         "rg",
			Repo:         "BurntSushi/ripgrep",
			IsArchive:    true,
			Tag:          "13.0.0",
			LinuxPattern: ".*unknown-linux-musl.*",
			MacPattern:   ".*apple-darwin.*",
		}
		g.Execute(UserConfig{
			BinaryDir:  dir,
			GithubUser: ghuser,
			GithubAuth: BasicAuth(ghuser, ghpat),
		}, SyncOpts{})

		checkFiles(t, dir, []string{"rg", "rg-13.0.0"})
	})

	t.Run("tarball_regex_specified", func(t *testing.T) {
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		dir := t.TempDir()

		g := GithubRelease{
			Name:         "gh",
			Repo:         "cli/cli",
			IsArchive:    true,
			Tag:          "v2.12.1",
			Regex:        ".*/bin/gh$",
			LinuxPattern: ".*linux_amd64.tar.gz$",
			MacPattern:   ".*macOS_amd64.tar.gz$",
		}
		g.Execute(UserConfig{
			BinaryDir:  dir,
			GithubUser: ghuser,
			GithubAuth: BasicAuth(ghuser, ghpat),
		}, SyncOpts{})

		checkFiles(t, dir, []string{"gh", "gh-v2.12.1"})
	})
}

func TestTagNormalization(t *testing.T) {
	testData := []struct {
		name string
		tag  string
		want string
	}{
		{
			name: "kustomize",
			tag:  "kustomize/v4.5.4",
			want: "kustomize-v4.5.4",
		},
		{
			name: "regular",
			tag:  "v0.21.0",
			want: "v0.21.0",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, GithubRelease{Tag: tc.tag}.normalizeTag())
		})
	}
}
