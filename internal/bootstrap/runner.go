package bootstrap

import (
	"bytes"
	"os/exec"
)

type runner struct {
	Items []Item
}

func NewRunner(items []Item) Runner {
	return runner{
		Items: items,
	}
}

func (r runner) RunAll() error {
	return nil
}

func (r runner) RunSingle(item Item) error {
	// Check if the item is already installed, and bail early
	installed, err := item.Check()
	if err != nil {
		return err
	}
	if installed {
		return nil
	}

	// Otherwise, install it
	return item.Install()
}

func runCmd(bin string, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func getReturnCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode(), nil
	} else {
		return -1, err
	}
}
