package file

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func createTempDir(t *testing.T, pattern string) (string, func()) {
    t.Helper()

    dir, err := ioutil.TempDir("", pattern)
    if err != nil {
        t.Fatalf("could not create temp directory %v", err)
    }

    remove := func() {
        os.RemoveAll(dir)
    }

    return dir, remove
}

func assertDirectoryContents(t *testing.T, dir string, want []string) {
    t.Helper()

    files, err := ioutil.ReadDir(dir)
    if err != nil {
        t.Fatalf("could not read directory %v", err)
    }

    var found []string

    for _, file := range files {
        found = append(found, file.Name())
    }

    if !reflect.DeepEqual(found, want) {
        t.Errorf("directory listings not equal got %v want %v", found, want)
    }
}

func writeData(t *testing.T, path string, data []byte) {
    t.Helper()

    err := ioutil.WriteFile(path, data, 0777)
    if err != nil {
        t.Fatalf("could not write file %v", err)
    }
}

func TestFile(t *testing.T) {
    t.Run("render outputs to build directory", func(t *testing.T) {
        src, removeSrc := createTempDir(t, "src")
        defer removeSrc()

        build, removeBuild := createTempDir(t, "build")
        defer removeBuild()

        srcFile := path.Join(src, "file_1.txt")
        err := ioutil.WriteFile(srcFile, []byte("some data"), 0777)
        if err != nil {
            t.Fatalf("could not write file %v", err)
        }

        f := File{TemplatePath: srcFile}
        err = f.Render(build)
        if err != nil {
            t.Errorf("error rendering %v", err)
        }

        want := []string{"file_1.txt"}
        assertDirectoryContents(t, build, want)
    })

    t.Run("output filename can be overridden", func(t *testing.T) {
        src, removeSrc := createTempDir(t, "src")
        defer removeSrc()

        build, removeBuild := createTempDir(t, "build")
        defer removeBuild()

        srcFile := path.Join(src, "file_1.txt")
        writeData(t, srcFile, []byte("some data"))

        f := File{
            TemplatePath: srcFile,
            DestinationName: "other_name.txt",
        }
        err := f.Render(build)
        if err != nil {
            t.Errorf("error rendering %v", err)
        }

        want := []string{"other_name.txt"}
        assertDirectoryContents(t, build, want)
    })
}
