package help

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func CreateTempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "test-"+pattern)
	if err != nil {
		t.Fatalf("could not create temp directory %v", err)
	}

	remove := func() {
		os.RemoveAll(dir)
	}

	return dir, remove
}

func AssertDirectoryContents(t *testing.T, dir string, want []string) {
	t.Helper()

	var allPaths []string

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("error walking directory %q, %v", dir, err)
		}
		if path != dir {
			relpath, _ := filepath.Rel(dir, path)
			allPaths = append(allPaths, relpath)
		}
		return nil
	})
	sort.Strings(allPaths)

	if !reflect.DeepEqual(allPaths, want) {
		t.Fatalf("directory listings not equal got %v want %v", allPaths, want)
	}
}

func GetDirContents(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatalf("could not read directory %v", err)
	}

	var found []string

	for _, file := range files {
		found = append(found, file.Name())
	}
	return found
}

func WriteData(t *testing.T, path string, data string) {
	t.Helper()

	err := ioutil.WriteFile(path, []byte(data), 0700)
	if err != nil {
		t.Fatalf("could not write file %v", err)
	}
}

func AssertSymlinkTo(t *testing.T, link string, source string) {
	t.Helper()

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("error getting link info for %q: %v", link, err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%q is not a symlink", link)
	}

	pointsTo, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("error reading symlink %q, %v", link, err)
	}

	if pointsTo != source {
		t.Fatalf("symlink not pointing to correct file, got %q want %q", pointsTo, source)
	}
}

type TempHomeDir struct {
	HomeDir string
}

func (t *TempHomeDir) GetHomeDir() (string, error) {
	return t.HomeDir, nil
}

func WriteConfig(t *testing.T, basedir string, contents string) {
	t.Helper()

	godotDir := filepath.Join(basedir, ".config", "godot")
	os.MkdirAll(godotDir, 744)
	WriteData(t, filepath.Join(godotDir, "config.json"), contents)
}

func WriteRepoConf(t *testing.T, dotDir string, contents string) {
	t.Helper()

	path := filepath.Join(dotDir, "config.json")
	WriteData(t, path, contents)
}
