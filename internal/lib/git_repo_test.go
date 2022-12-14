package lib

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitRepo(t *testing.T) {
	expectedCommitMsg := strings.Join([]string{
		"commit 02c8c0085385f7d65ba35556edfc58e0f48257eb",
		"Author: Scott Baker <scott.baker@directlink.coop>",
		"Date:   Mon Jun 14 12:44:48 2021 -0700",
		"",
		"    Bump the fatpacked version",
		"",
	}, "\n")

	checkMsg := func(t *testing.T, loc string) {
		stdout, _, err := runCmd("git", "-C", loc, "log", "-n", "1")
		require.NoError(t, err)

		require.Equal(t, expectedCommitMsg, stdout)
	}

	t.Run("as_tag", func(t *testing.T) {
		dir := t.TempDir()
		loc := path.Join(dir, "foo-repo")

		gr := GitRepo{
			Name:     "foo-repo",
			URL:      "https://github.com/so-fancy/diff-so-fancy.git",
			Location: loc,
			Ref: Ref{
				Tag: "v1.4.2",
			},
		}
		require.NoError(t, gr.Execute(UserConfig{}, SyncOpts{}, GodotConfig{}))
		checkMsg(t, loc)
	})

	t.Run("as_commit", func(t *testing.T) {
		dir := t.TempDir()
		loc := path.Join(dir, "foo-repo")

		gr := GitRepo{
			Name:     "foo-repo",
			URL:      "https://github.com/so-fancy/diff-so-fancy.git",
			Location: loc,
			Ref: Ref{
				Commit: "02c8c0085385f7d65ba35556edfc58e0f48257eb",
			},
		}
		require.NoError(t, gr.Execute(UserConfig{}, SyncOpts{}, GodotConfig{}))
		checkMsg(t, loc)
	})
}
