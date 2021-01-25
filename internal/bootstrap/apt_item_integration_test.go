// +build apt_integration integration

package bootstrap

import (
	"os/exec"
	"testing"
)

func ensure(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAptItem(t *testing.T) {
	item := NewAptItem("ripgrep")

	// Ripgrep should not be installed initially
	installed, err := item.Check()
	ensure(t, err)
	if installed {
		t.Fatalf("ripgrep should not show as installed")
	}

	// Now we install it
	err = item.Install()
	ensure(t, err)

	// And now it should be installed
	installed, err = item.Check()
	ensure(t, err)
	if !installed {
		t.Fatalf("ripgrep should show as installed")
	}

	// Double check that it's installed
	cmd := exec.Command("rg", "--version")
	err = cmd.Run()
	ensure(t, err)

	// Now ensure a bad install will error out
	item = NewAptItem("not_a_binary_at_all")
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
