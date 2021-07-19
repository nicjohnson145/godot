package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithubReleaseOps(t *testing.T) {
	paths, opts := setupConfigs(t, Setup{
		ConfData: map[string]interface{}{
			"hosts": map[string]interface{}{
				"host1": map[string][]string{
					"github_releases": {},
				},
				"host2": map[string][]string{
					"github_releases": {},
				},
			},
			"github_releases": map[string]interface{}{
				"repo": map[string]interface{}{
					"repo_name": "some/repo",
					"patterns": map[string]string{
						"darwin": "repo_.*-apple-darwin.tar.gz",
						"linux": "repo_.*-unknown-linux-musl.tar.gz",
					},
					"download": map[string]string{
						"type": "tar_gz",
						"path": "bin/repo",
					},
				},
			},
		},
		Target: "host1",
	})
	defer paths.Remove()

	// Don't have the mechanisms for adding releases yet, so just use/cease/list


	// List out all the releases
	stdOut, _, err := runCmd(t, opts, "list", "github-release")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"repo (https://github.com/some/repo)",
			"     - DownloadType: tar_gz",
			"     - DownloadPath: bin/repo",
			"     - darwin Pattern: repo_.*-apple-darwin.tar.gz",
			"     - linux Pattern: repo_.*-unknown-linux-musl.tar.gz",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// host1 shouldn't have any releases
	stdOut, _, err = runCmd(t, opts, "list", "github-release", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"",
		}, "\n"),
		stdOut.String(),
	)

	// host2 shouldn't have any releases
	stdOut, _, err = runCmd(t, opts, "list", "github-release", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Add one to both
	_, _, err = runCmd(t, opts, "use", "github-release", "repo")
	require.NoError(t, err)
	_, _, err = runCmd(t, opts, "use", "github-release", "repo", "--target=host2")
	require.NoError(t, err)

	// host1 should have one now
	stdOut, _, err = runCmd(t, opts, "list", "github-release", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"repo",
			"     - Location: ~/bin",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// host2 should as well
	stdOut, _, err = runCmd(t, opts, "list", "github-release", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"repo",
			"     - Location: ~/bin",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Take them away
	_, _, err = runCmd(t, opts, "cease", "github-release", "repo")
	require.NoError(t, err)
	_, _, err = runCmd(t, opts, "cease", "github-release", "repo", "--target=host2")
	require.NoError(t, err)

	// host1 shouldn't have any releases
	stdOut, _, err = runCmd(t, opts, "list", "github-release", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"",
		}, "\n"),
		stdOut.String(),
	)

	// host2 shouldn't have any releases
	stdOut, _, err = runCmd(t, opts, "list", "github-release", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"",
		}, "\n"),
		stdOut.String(),
	)
}
