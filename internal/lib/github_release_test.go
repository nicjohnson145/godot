package lib

import (
	"encoding/json"
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
			g := GithubRelease{Tag: tc.tag}
			require.Equal(t, tc.want, g.normalizeTag())
		})
	}
}

func TestGetAssetAutoDetection(t *testing.T) {

	testData := []struct {
		name      string
		path      string
		os        string
		arch      string
		want      string
		IsArchive bool
	}{
		{
			name: "rg-linux-amd64",
			path: "rg-13.0.0.json",
			os:   "linux",
			arch: "amd64",
			want: "ripgrep-13.0.0-x86_64-unknown-linux-musl.tar.gz",
			IsArchive: true,
		},
		{
			name: "zoxide-linux-amd64",
			path: "zoxide-v0.8.2.json",
			os:   "linux",
			arch: "amd64",
			want: "zoxide-0.8.2-x86_64-unknown-linux-musl.tar.gz",
			IsArchive: true,
		},
		{
			name: "rg-mac-amd64",
			path: "rg-13.0.0.json",
			os:   "darwin",
			arch: "amd64",
			want: "ripgrep-13.0.0-x86_64-apple-darwin.tar.gz",
			IsArchive: true,
		},
		{
			name: "zoxide-mac-amd64",
			path: "zoxide-v0.8.2.json",
			os:   "darwin",
			arch: "amd64",
			want: "zoxide-0.8.2-x86_64-apple-darwin.tar.gz",
			IsArchive: true,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			restore, noFatal := NoFatals(t)
			defer noFatal(t)
			defer restore()

			data, err := os.ReadFile("testdata/asset-json/" + tc.path)
			require.NoError(t, err)

			var resp releaseResponse
			err = json.Unmarshal(data, &resp)
			require.NoError(t, err)
			g := GithubRelease{}
			got := g.getAsset(resp, tc.os, tc.arch).Name
			require.Equal(t, tc.want, got)
			// Should also properly set that the release is an archive
			require.Equal(t, tc.IsArchive, g.IsArchive)
		})
	}
}
