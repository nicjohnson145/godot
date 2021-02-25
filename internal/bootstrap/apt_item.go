package bootstrap

import "fmt"

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
	_, _, err := runCmd("/bin/sh", "-c", fmt.Sprintf("sudo apt install -y %v", i.Name))
	return err
}
