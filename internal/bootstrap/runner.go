package bootstrap

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
	return nil
}
