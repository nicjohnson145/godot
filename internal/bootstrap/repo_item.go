package bootstrap

type repoItem struct {
	Name string
	Location string
}

func NewRepoItem(name string, location string) repoItem {
	return repoItem{
		Name: name,
		Location: location,
	}
}

func (i repoItem) Check() (bool, error) {
	// Check if directory exists and has a .git
	return false, nil
}

func (i repoItem) Install() error {
	return nil
}
