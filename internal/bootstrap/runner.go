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
	// Check if the item is already installed, and bail early
	installed, err := item.Check()
	if err != nil {
		return err
	}
	if installed {
		return nil
	}

	// Otherwise, install it
	return item.Install()
}
