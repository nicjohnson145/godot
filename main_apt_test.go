// +build apt_integration

package main

import (
	"github.com/stretchr/testify/require"
	"os/exec"
	"testing"
)

func TestBootstrapApt(t *testing.T) {
	paths, opts := setupConfigs(t, Setup{
		ConfData: map[string]interface{}{
			"hosts": map[string]interface{}{
				"host1": map[string][]string{
					"bootstraps": {"ripgrep"},
				},
			},
			"bootstraps": map[string]interface{}{
				"ripgrep": map[string]interface{}{
					"apt": map[string]string{
						"name": "ripgrep",
					},
				},
			},
		},
		Target: "host1",
	})
	defer paths.Remove()

	cmd := exec.Command("rg", "--version")
	err := cmd.Run()
	require.Error(t, err)

	_, _, err = runCmd(t, opts, "sync")
	require.NoError(t, err)

	cmd = exec.Command("rg", "--version")
	err = cmd.Run()
	require.NoError(t, err)
}
