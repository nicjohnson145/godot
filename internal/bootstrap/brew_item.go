package bootstrap

import (
	"github.com/nicjohnson145/godot/internal/bootstrap/brewcache"
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
	b := brewcache.GetInstance()
	return b.IsInstalled(i.Name), nil
}

func (i brewItem) Install() error {
	_, _, err := help.RunCmd("brew", "install", i.Name)
	return err
}
