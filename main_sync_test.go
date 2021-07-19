package main

import (
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	paths, opts := setupConfigs(t, Setup{
		ConfData: map[string]interface{}{
			"files": map[string]string{
				"dot_zshrc": "~/.zshrc",
				"some_conf": "~/.config/prog/conf",
				"my_conf": "~/my_conf",
			},
			"hosts": map[string]interface{}{
				"host1": map[string][]string{
					"files": {"some_conf", "dot_zshrc", "my_conf"},
				},
			},
		},
		Target: "host1",
		TemplateContent: map[string]string{
			"dot_zshrc": "zshrc content",
			"some_conf": "some conf content",
			"my_conf": "new my conf content",
		},
	})
	defer paths.Remove()

	// Touch the file, this should make sync error on the first try
	help.WriteData(t, filepath.Join(paths.Home, "my_conf"), "orig my conf content")

	// Create the symlinks
	_, _, err := runCmd(t, opts, "sync")
	require.Error(t, err)

	// Make sure they're pointing to the right place
	help.AssertSymlinkTo(
		t,
		filepath.Join(paths.Home, ".zshrc"),
		filepath.Join(paths.Dotfiles, "build", "dot_zshrc"),
	)
	help.AssertSymlinkTo(
		t,
		filepath.Join(paths.Home, ".config", "prog", "conf"),
		filepath.Join(paths.Dotfiles, "build", "some_conf"),
	)

	// And have the right content
	help.AssertFileContents(
		t,
		filepath.Join(paths.Home, ".zshrc"),
		"zshrc content",
	)
	help.AssertFileContents(
		t,
		filepath.Join(paths.Home, ".config", "prog", "conf"),
		"some conf content",
	)
	help.AssertFileContents(
		t,
		filepath.Join(paths.Home, "my_conf"),
		"orig my conf content",
	)

	// Now force it, to render over the file
	_, _, err = runCmd(t, opts, "sync", "-f")
	require.NoError(t, err)

	// Should have the new content now
	help.AssertFileContents(
		t,
		filepath.Join(paths.Home, "my_conf"),
		"new my conf content",
	)
}
