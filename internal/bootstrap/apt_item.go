package bootstrap

type aptItem struct {
	Name string
}

func NewAptItem(name string) Item {
	return aptItem{
		Name: name,
	}
}

func (i aptItem) Check() (bool, error) {
	_, _, err := runCmd("dpkg", "-l", i.Name)
	returnCode, err := getReturnCode(err)
	if err != nil {
		return false, err
	}
	return returnCode == 0, nil
}

func (i aptItem) Install() error {
	_, _, err := runCmd("apt", "install", "-y", i.Name)
	return err
}
