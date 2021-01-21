package bootstrap

import (
	"bytes"
	"os/exec"
)

type aptItem struct {
	Name string
}

func NewAptItem(name string) Item {
	return aptItem{
		Name: name,
	}
}

func (i aptItem) Check() (bool, error) {
	if _, _, err := i.runCmd("dpkg", "-l", i.Name); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			} else {
				return false, exitError
			}
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func (i aptItem) Install() error {
	_, _, err := i.runCmd("apt", "install", "-y", i.Name)
	return err
}

func (i aptItem) runCmd(bin string, args...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
