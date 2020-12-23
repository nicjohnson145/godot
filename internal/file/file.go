package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File struct {
	DestinationPath string
	TemplatePath    string
}

func (f *File) render(buildDir string) error {
	bytes, err := ioutil.ReadFile(f.TemplatePath)
	if err != nil {
		err = fmt.Errorf("could not open %q for reading, %v", f.TemplatePath, err)
		return err
	}
	contents := string(bytes)

	destPath := filepath.Join(buildDir, filepath.Base(f.TemplatePath))

	err = ioutil.WriteFile(destPath, []byte(contents), 0700)
	if err != nil {
		err = fmt.Errorf("could not open %q for writing, %v", destPath, err)
		return err
	}

	return nil
}

func (f *File) symlink(buildDir string) error {
	src := filepath.Join(buildDir, filepath.Base(f.TemplatePath))
	destbase := filepath.Dir(f.DestinationPath)
	err := os.MkdirAll(destbase, 0700)
	if err != nil {
		err = fmt.Errorf("unable to create dir %q, %v", destbase, err)
		return err
	}
	err = os.Symlink(src, f.DestinationPath)
	if err != nil {
		err = fmt.Errorf("unable to symlink %q to %q, %v", src, f.DestinationPath, err)
		return err
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
