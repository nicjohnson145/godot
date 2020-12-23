package file

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type HomeDirGetter interface {
	GetHomeDir() (string, error)
}

type ConfigGetter interface {
	GetConfig()
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

func substituteTilde(f *File, home string) {
	if strings.HasPrefix(f.DestinationPath, "~/") {
		f.DestinationPath = filepath.Join(home, f.DestinationPath[2:])
	}
}

type renderer struct {
	Files []File
	DotfilesRoot string
}

func NewRenderer(files []File, home HomeDirGetter, root string) (*renderer, error) {

	homeDir, err := home.GetHomeDir()
	if err != nil {
		err = fmt.Errorf("error getting home directory: %v", err)
		return nil, err
	}
	for i := range files {
		substituteTilde(&files[i], homeDir)
	}

	r := &renderer{
		Files: files,
		DotfilesRoot: root,
	}
	return r, nil
}

func (r  *renderer) ensureBuildDir() (string, error) {
	dir := filepath.Join(r.DotfilesRoot, "build")
	err := os.MkdirAll(dir, 700)
	if err != nil {
		err := fmt.Errorf("error creating build directory in %q, %v", dir, err)
		return "", err
	}
	return dir, nil
}

func (r *renderer) Render() error {
	buildDir, err := r.ensureBuildDir()
	if err != nil {
		return err
	}
	for _, fl := range r.Files {
		err := fl.Build(buildDir)
		if err != nil {
			err = fmt.Errorf("error rendering file %q, %v", fl.DestinationPath, err)
			return err
		}
	}
	return nil
}
