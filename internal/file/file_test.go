package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func writeData(t *testing.T, path string, data string) {
	t.Helper()

	err := ioutil.WriteFile(path, []byte(data), 0777)
	if err != nil {
		t.Fatalf("could not write file %v", err)
	}
}

func assertSymlinkTo(t *testing.T, link string, source string) {
	t.Helper()

	info, err := os.Lstat(link)
	if err != nil {
		t.Errorf("error getting link info for %q: %v", link, err)
	}

	if info.Mode() & os.ModeSymlink == 0 {
		t.Errorf("%q is not a symlink", link)
	}

	pointsTo, err := os.Readlink(link)
	if err != nil {
		t.Errorf("error reading symlink %q, %v", link, err)
	}

	if pointsTo != source {
		t.Errorf("symlink not pointing to correct file, got %q want %q", pointsTo, source)
	}
}

type TempHomeDir struct {
	HomeDir string
}

func (t *TempHomeDir) GetHomeDir() string {
	return t.HomeDir
}


func TestFile(t *testing.T) {
	t.Run("files without templating are copied and symlinked", func(t *testing.T) {
		src, removeSrc := createTempDir(t, "src")
		defer removeSrc()

		build, removeBuild := createTempDir(t, "build")
		defer removeBuild()

		home, removeHome := createTempDir(t, "home")
		defer removeHome()

		writeData(t, filepath.Join(src, "some_file"), "the contents")

		f := File{
			DestinationPath: filepath.Join(home, ".some_other_file"),
			TemplatePath: filepath.Join(src, "some_file"),
		}
		err := f.Build(build)
		if err != nil {
			t.Errorf("error building, %v", err)
		}

		want := []string{".some_other_file"}
		assertDirectoryContents(t, home, want)
		assertSymlinkTo(
			t,
			filepath.Join(home, ".some_other_file"),
			filepath.Join(build, "some_file"),
		)
	})

}

