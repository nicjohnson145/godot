package lib

import (
	"fmt"

	"github.com/nicjohnson145/godot/internal/help"
)

var _ Item = (*aptItem)(nil)

type aptItem struct {
	Name string
}

func NewAptItem(name string) Item {
	return aptItem{
		Name: name,
	}
}

func (i aptItem) Check() (bool, error) {
	_, _, err := help.RunCmd("dpkg", "-l", i.Name)
	returnCode, err := help.GetReturnCode(err)
	if err != nil {
		return false, err
	}
	return returnCode == 0, nil
}

func (i aptItem) Install() error {
	_, _, err := help.RunCmd("/bin/sh", "-c", fmt.Sprintf("sudo apt install -y %v", i.Name))
	return err
}
