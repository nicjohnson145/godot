package bootstrap

import (
	"github.com/nicjohnson145/godot/internal/help"
)

type brewItem struct {
	Name string
}

func NewBrewItem(name string) brewItem {
	return brewItem{
		Name: name,
	}
}

func (i brewItem) Check() (bool, error) {
	_, _, err := help.RunCmd("brew", "list", i.Name)
	returnCode, err := help.GetReturnCode(err)
	if err != nil {
		return false, err
	}
	return returnCode == 0, nil
}

func (i brewItem) Install() error {
	_, _, err := help.RunCmd("brew", "install", i.Name)
	return err
}
