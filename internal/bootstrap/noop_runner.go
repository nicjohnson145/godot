package bootstrap

import (
	"github.com/nicjohnson145/godot/internal/config"
)

type NoopRunner struct {
	RunAllArgs    [][]config.BootstrapImpl
	RunSingleArgs []config.BootstrapImpl
	RunAllErr     error
	RunSingleErr  error
}

func (n *NoopRunner) RunAll(items []config.BootstrapImpl) error {
	n.RunAllArgs = append(n.RunAllArgs, items)
	return n.RunAllErr
}

func (n *NoopRunner) RunSingle(item config.BootstrapImpl) error {
	n.RunSingleArgs = append(n.RunSingleArgs, item)
	return n.RunSingleErr
}
