package controller

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nicjohnson145/godot/internal/bootstrap"
	"github.com/nicjohnson145/godot/internal/help"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/stretchr/testify/require"
)

func setupGithubRelease(t *testing.T) testData {
	t.Helper()
	home, dotpath, remove := help.SetupDirectories(t, "host1")
	repoConf := `{
		"github_releases": {
			"repo": {
				"repo_name": "some/repo",
				"patterns": {
					"darwin": "repo_.*-apple-darwin.tar.gz",
					"linux": "repo_.*-unknown-linux-musl.tar.gz"
				},
				"download": {
					"type": "archive",
					"path": "bin/repo"
				}
			}
		},
		"hosts": {
			"host1": {
				"files": [],
				"bootstraps": [],
				"github_releases": []
			}
		}
	}`
	help.WriteRepoConf(t, dotpath, repoConf)
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

func TestGithubRelease(t *testing.T) {
	obj := setupGithubRelease(t)
	defer obj.Remove()

	// Show current releases
	expected := strings.Join(
		[]string{
			"repo (https://github.com/some/repo)",
			"     - DownloadType: archive",
			"     - DownloadPath: bin/repo",
			"     - darwin Pattern: repo_.*-apple-darwin.tar.gz",
			"     - linux Pattern: repo_.*-unknown-linux-musl.tar.gz",
		},
		"\n",
	) + "\n"
	b := bytes.NewBufferString("")
	err := obj.C.ShowGithubRelease("", b)
	require.NoError(t, err)
	require.Equal(t, expected, b.String())

	// Host1 should have no releases
	b = bytes.NewBufferString("")
	err = obj.C.ShowGithubRelease("host1", b)
	require.NoError(t, err)
	require.Equal(t, "", b.String())

	// Add the release to host1
	err = obj.C.TargetUseGithubRelease("host1", "repo", "", true)
	require.NoError(t, err)
	b = bytes.NewBufferString("")
	err = obj.C.ShowGithubRelease("host1", b)
	require.NoError(t, err)
	expected = strings.Join(
		[]string{
			"repo",
			"     - Location: ~/bin/repo",
			"     - Track Updates: true",
		},
		"\n",
	) + "\n"
	require.Equal(t, expected, b.String())

	// Remove the release again
	err = obj.C.TargetCeaseGithubRelease("host1", []string{"repo"})
	require.NoError(t, err)
	b = bytes.NewBufferString("")
	err = obj.C.ShowGithubRelease("host1", b)
	require.NoError(t, err)
	require.Equal(t, "", b.String())
}
