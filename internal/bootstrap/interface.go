package bootstrap

import (
	"github.com/nicjohnson145/godot/internal/config"
)

type Item interface {
	Check() (bool, error)
	Install() error
}

type Runner interface {
	RunAll([]config.BootstrapImpl) error
	RunSingle(config.BootstrapImpl) error
}
