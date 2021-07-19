package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBootstrapOps(t *testing.T) {
	paths, opts := setupConfigs(t, Setup{
		ConfData: map[string]interface{}{
			"hosts": map[string]interface{}{
				"host1": map[string][]string{
					"bootstraps": {"jq"},
				},
				"host2": map[string][]string{
					"bootstraps": {"pyenv"},
				},
			},
			"bootstraps": map[string]interface{}{
				"jq": map[string]interface{}{
					"apt": map[string]string{
						"name": "jq",
					},
				},
				"pyenv": map[string]interface{}{
					"brew": map[string]string{
						"name": "pyenv",
					},
					"git": map[string]string{
						"name":     "https::/github.com/pyenv/pyenv",
						"location": "~/.pyenv",
					},
				},
			},
		},
		Target: "host1",
	})
	defer paths.Remove()

	// List out all the bootstraps
	stdOut, _, err := runCmd(t, opts, "list", "bootstraps")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"   jq => apt",
			"pyenv => brew, git",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// List out all the bootstraps for host1
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"jq => apt",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// List out all the bootstraps for host2
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"pyenv => brew, git",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Add a couple new ones
	_, _, err = runCmd(t, opts, "add", "bootstrap", "ripgrep", "-m", "brew", "-n", "ripgrep")
	require.NoError(t, err)
	_, _, err = runCmd(t, opts, "add", "bootstrap", "exa", "-m", "brew", "-n", "exa")
	require.NoError(t, err)

	// List out all the bootstraps
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"    exa => brew",
			"     jq => apt",
			"  pyenv => brew, git",
			"ripgrep => brew",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Set host1 to use ripgrep and host2 to use exa
	_, _, err = runCmd(t, opts, "use", "bootstrap", "ripgrep")
	require.NoError(t, err)
	_, _, err = runCmd(t, opts, "use", "bootstrap", "exa", "--target=host2")
	require.NoError(t, err)

	// List out all the bootstraps for host1
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"     jq => apt",
			"ripgrep => brew",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// List out all the bootstraps for host2
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"  exa => brew",
			"pyenv => brew, git",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Remove the new usages
	_, _, err = runCmd(t, opts, "cease", "bootstrap", "ripgrep")
	require.NoError(t, err)
	_, _, err = runCmd(t, opts, "cease", "bootstrap", "exa", "--target=host2")
	require.NoError(t, err)

	// List out all the bootstraps for host1
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"jq => apt",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// List out all the bootstraps for host2
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"pyenv => brew, git",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// Add a usage back, but all at once
	_, _, err = runCmd(t, opts, "use", "bootstrap", "exa", "--target=ALL")

	// List out all the bootstraps for host1
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"exa => brew",
			" jq => apt",
			"",
		}, "\n"),
		stdOut.String(),
	)

	// List out all the bootstraps for host2
	stdOut, _, err = runCmd(t, opts, "list", "bootstraps", "--target=host2")
	require.NoError(t, err)
	require.Equal(
		t,
		strings.Join([]string{
			"  exa => brew",
			"pyenv => brew, git",
			"",
		}, "\n"),
		stdOut.String(),
	)
}
