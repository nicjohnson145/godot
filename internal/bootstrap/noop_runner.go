package bootstrap

type NoopRunner struct {
	RunAllArgs    [][]Item
	RunSingleArgs []Item
}

func (n *NoopRunner) RunAll(items []Item) {
	n.RunAllArgs = append(n.RunAllArgs, items)
}

func (n *NoopRunner) RunSingle(item Item) {
	n.RunSingleArgs = append(n.RunSingleArgs, item)
}
