package lib

var _ Executor = (*Bundle)(nil)

type Bundle struct {
	Name  string
	Items []string `json:"items"`
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
