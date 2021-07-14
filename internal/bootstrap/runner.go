package bootstrap

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r Runner) RunAll(items []Item) {
	for _, item := range items {
		r.runSingleItem(item)
	}
}

func (r Runner) RunSingle(item Item) {
	r.runSingleItem(item)
}

func (r Runner) runSingleItem(item Item) {
	// Check if the item is already installed, and bail early
	installed, err := item.Check()
	if err != nil {
		logrus.Error(fmt.Sprintf("Error checking installation status of %v: %v", item.GetName(), err))
		return
	}
	if installed {
		return
	}

	// Otherwise, install it
	err = item.Install()
	if err != nil {
		logrus.Error(fmt.Sprintf("Error installing %v: %v", item.GetName(), err))
	}
}
