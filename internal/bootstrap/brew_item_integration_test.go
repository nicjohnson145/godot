// +build brew_integration

package bootstrap

import (
	"os/exec"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/nicjohnson145/godot/internal/bootstrap/brewcache"
)

func TestBrewItem(t *testing.T) {
	item := NewBrewItem("jq")

	installed, err := item.Check()
	require.NoError(t, err)
	if installed {
		t.Fatalf("jq should not already be installed")
	}

	err = item.Install()
	require.NoError(t, err)

	// Reset the brewcache, cause singleton
	brewcache.GetInstance().Reset()

	installed, err = item.Check()
	require.NoError(t, err)

	if !installed {
		t.Fatalf("jq should have been installed")
	}

	cmd := exec.Command("jq", "--version")
	err = cmd.Run()
	require.NoError(t, err)

	// Reset the brewcache, cause singleton
	brewcache.GetInstance().Reset()

	item = NewBrewItem("not_a_binary_at_all")
	installed, err = item.Check()
	require.NoError(t, err)

	if installed {
		t.Fatalf("not_a_binary_at_all should not show as installed")
	}

	err = item.Install()
	require.Error(t, err)
}
