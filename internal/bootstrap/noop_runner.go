package bootstrap

type NoopRunner struct {
	RunAllArgs    [][]Item
	RunSingleArgs []Item
	RunAllErr     error
	RunSingleErr  error
}

func (n *NoopRunner) RunAll(items []Item) error {
	n.RunAllArgs = append(n.RunAllArgs, items)
	return n.RunAllErr
}

func (n *NoopRunner) RunSingle(item Item) error {
	n.RunSingleArgs = append(n.RunSingleArgs, item)
	return n.RunSingleErr
}
