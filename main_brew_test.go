// +build brew_integration

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
					"bootstraps": {"jq"},
				},
			},
			"bootstraps": map[string]interface{}{
				"jq": map[string]interface{}{
					"brew": map[string]string{
						"name": "jq",
					},
				},
			},
		},
		Target: "host1",
		PackageManagers: []string{"brew"},
	})
	defer paths.Remove()

	cmd := exec.Command("jq", "--version")
	err := cmd.Run()
	require.Error(t, err)

	_, _, err = runCmd(t, opts, "sync")
	require.NoError(t, err)

	cmd = exec.Command("jq", "--version")
	err = cmd.Run()
	require.NoError(t, err)
}
