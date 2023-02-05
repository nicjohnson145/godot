package lib

import (
	"github.com/rs/zerolog"
)

var _ Executor = (*Bundle)(nil)

type Bundle struct {
	Name  string `yaml:"-"`
	Items []string `yaml:"items"`
}

func (b *Bundle) SetLogger(_ zerolog.Logger) {
	// No-op since this is really just a container executor
}

func (b *Bundle) Execute(_ UserConfig, _ SyncOpts, _ GodotConfig) error {
	// Intentional noop, all the inner executors do the actual work
	return nil
}

func (b *Bundle) GetName() string {
	return b.Name
}

func (b *Bundle) SetName(n string) {
	b.Name = n
}

func (b *Bundle) Type() ExecutorType {
	return ExecutorTypeBundle
}

func (b *Bundle) Validate() error {
	// TODO: should validate contents here
	return nil
}
