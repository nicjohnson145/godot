// +build brew_integration

package bootstrap

import (
	"os/exec"
	"testing"
)

func TestBrewItem(t *testing.T) {
	ensure := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	item := NewBrewItem("jq")

	installed, err := item.Check()
	ensure(t, err)
	if installed {
		t.Fatalf("jq should not already be installed")
	}

	err = item.Install()
	ensure(t, err)

	installed, err = item.Check()
	ensure(t, err)
	if !installed {
		t.Fatalf("jq should have been installed")
	}

	cmd := exec.Command("jq", "--version")
	err = cmd.Run()
	ensure(t, err)

	item = NewBrewItem("not_a_binary_at_all")
	installed, err = item.Check()
	ensure(t, err)
	if installed {
		t.Fatalf("not_a_binary_at_all should not show as installed")
	}

	err = item.Install()
	if err == nil {
		t.Fatal("installing a bad package should have errored")
	}
}
