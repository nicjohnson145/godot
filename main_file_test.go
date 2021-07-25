package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/stretchr/testify/require"
)

func TestFileOps(t *testing.T) {
	paths, opts := setupConfigs(t, Setup{
		ConfData: map[string]interface{}{
			"files": map[string]string{
				"dot_zshrc": "~/.zshrc",
				"some_conf": "/etc/conf",
			},
			"hosts": map[string]interface{}{
				"host1": map[string][]string{
					"files": {"some_conf"},
				},
				"host2": map[string][]string{
					"files": {"dot_zshrc"},
				},
			},
		},
		Target: "host1",
	})
	defer paths.Remove()

	// List out all the files
	stdOut, _, err := runCmd(t, opts, "list", "files")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
			"some_conf => /etc/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Create a new file, and add it, letting the name be inferred
	help.WriteData(t, filepath.Join(paths.Home, ".new_conf"), "new conf data")
	// The shell would normally expand "~" for me, so mimic it here
	_, _, err = runCmd(t, opts, "add", "file", filepath.Join(paths.Home, ".new_conf"))
	require.NoError(t, err)

	// New file should be present
	stdOut, _, err = runCmd(t, opts, "list", "files")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_new_conf => ~/.new_conf",
			"   dot_zshrc => ~/.zshrc",
			"   some_conf => /etc/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)
	// Data should be imported from orignal file
	help.AssertFileContents(t, filepath.Join(paths.Dotfiles, "templates", "dot_new_conf"), "new conf data")

	// Create another new file, but use the as functionality this time
	help.WriteData(t, filepath.Join(paths.Home, "conf"), "home conf data")
	// The shell would normally expand "~" for me, so mimic it here
	_, _, err = runCmd(t, opts, "add", "file", filepath.Join(paths.Home, "conf"), "--as=home_conf")
	require.NoError(t, err)
	// New file should be present
	stdOut, _, err = runCmd(t, opts, "list", "files")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_new_conf => ~/.new_conf",
			"   dot_zshrc => ~/.zshrc",
			"   home_conf => ~/conf",
			"   some_conf => /etc/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)
	// Data should be imported from orignal file
	help.AssertFileContents(t, filepath.Join(paths.Dotfiles, "templates", "home_conf"), "home conf data")

	// Have host1 use the new conf
	_, _, err = runCmd(t, opts, "use", "file", "dot_new_conf")
	require.NoError(t, err)
	// And host2 use the home conf
	_, _, err = runCmd(t, opts, "use", "file", "home_conf", "--target=host2")
	require.NoError(t, err)

	// Prove they've used them
	stdOut, _, err = runCmd(t, opts, "list", "files", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_new_conf => ~/.new_conf",
			"   some_conf => /etc/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)
	stdOut, _, err = runCmd(t, opts, "list", "files", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
			"home_conf => ~/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Remove those files from each host now
	_, _, err = runCmd(t, opts, "cease", "file", "dot_new_conf")
	require.NoError(t, err)
	_, _, err = runCmd(t, opts, "cease", "file", "home_conf", "--target=host2")
	require.NoError(t, err)

	// And now they're gone
	stdOut, _, err = runCmd(t, opts, "list", "files", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"some_conf => /etc/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)
	stdOut, _, err = runCmd(t, opts, "list", "files", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Now add one back, but this time use the all arg to do everything at once
	_, _, err = runCmd(t, opts, "use", "file", "home_conf", "--target=ALL")
	require.NoError(t, err)

	// Prove they've used them
	stdOut, _, err = runCmd(t, opts, "list", "files", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"home_conf => ~/conf",
			"some_conf => /etc/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)
	stdOut, _, err = runCmd(t, opts, "list", "files", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"dot_zshrc => ~/.zshrc",
			"home_conf => ~/conf",
			"",
		}, "\n"),
		stdOut.String(),
	)
}
