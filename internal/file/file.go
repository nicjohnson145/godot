package file

import (
	"fmt"
	"io/ioutil"
	"path"
)


type File struct {
    DestinationName string
    TemplatePath string
}



func (f *File) Render(buildDir string) error {
    bytes, err := ioutil.ReadFile(f.TemplatePath)
    if err != nil {
        err = fmt.Errorf("could not open %q for reading, %v", f.TemplatePath, err)
        return err
    }
    contents := string(bytes)

    var destPath string
    if f.DestinationName == "" {
        destPath = path.Join(buildDir, path.Base(f.TemplatePath))
    } else {
        destPath = path.Join(buildDir, f.DestinationName)
    }

    err = ioutil.WriteFile(destPath, []byte(contents), 0644)
    if err != nil {
        err = fmt.Errorf("could not open %q for writing, %v", destPath, err)
        return err
    }

    return nil
}
