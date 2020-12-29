package file

import (
	"fmt"
	"os"
	"path/filepath"
)

type Renderer struct {
	Files        []File
	DotfilesRoot string
}

func NewRenderer(files []File, root string) *Renderer {
	return &Renderer{
		Files:        files,
		DotfilesRoot: root,
	}
}

func (r *Renderer) ensureBuildDir() (string, error) {
	dir := filepath.Join(r.DotfilesRoot, "build")
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		err := fmt.Errorf("error creating build directory in %q, %v", dir, err)
		return "", err
	}
	return dir, nil
}

func (r *Renderer) Render() error {
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
