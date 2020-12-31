package util

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

type HomeDirGetter interface {
	GetHomeDir() (string, error)
}

type OSHomeDir struct{}

func (o *OSHomeDir) GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		err = fmt.Errorf("could not get value of current user %v", err)
	}
	dir := usr.HomeDir
	return dir, err
}

func ReplacePrefix(path string, prefix string, newPrefix string) string {
	newPath := path
	if strings.HasPrefix(path, prefix) {
		newPath = filepath.Join(newPrefix, path[len(prefix):])
	}
	return newPath
}

func ToTemplateName(path string) string {
	name := filepath.Base(path)
	if strings.HasPrefix(name, ".") {
		return "dot_" + name[1:]
	}
	return name
}
