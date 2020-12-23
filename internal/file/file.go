package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type HomeDirGetter interface {
	GetHomeDir() string
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

type File struct {
	DestinationPath string
	TemplatePath    string
}

func NewFile(template string, destination string, home HomeDirGetter) *File {
	f := &File{
		TemplatePath: template,
		DestinationPath: destination,
	}
	substituteTilde(f, home)

	return f
}

func (f *File) render(buildDir string) error {
	bytes, err := ioutil.ReadFile(f.TemplatePath)
	if err != nil {
		err = fmt.Errorf("could not open %q for reading, %v", f.TemplatePath, err)
		return err
	}
	contents := string(bytes)

	destPath := filepath.Join(buildDir, filepath.Base(f.TemplatePath))

	err = ioutil.WriteFile(destPath, []byte(contents), 0644)
	if err != nil {
		err = fmt.Errorf("could not open %q for writing, %v", destPath, err)
		return err
	}

	return nil
}

func (f *File) symlink(buildDir string) error {
	src := filepath.Join(buildDir, filepath.Base(f.TemplatePath))
	err := os.Symlink(src, f.DestinationPath)
	if err != nil {
		err = fmt.Errorf("unable to symlink %q to %q, %v", src, f.DestinationPath, err)
	}
	return err
}

func (f *File) Build(buildDir string) error {
	err := f.render(buildDir)
	if err != nil {
		return err
	}
	err = f.symlink(buildDir)
	if err != nil {
		return err
	}
	return nil
}

func substituteTilde(f *File, home HomeDirGetter) {
	if strings.HasPrefix(f.DestinationPath, "~/") {
		f.DestinationPath = filepath.Join(home.GetHomeDir(), f.DestinationPath[2:])
	}
}
