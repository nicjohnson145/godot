package help

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const SAMPLE_CONFIG = `{
	"files": {
		"dot_zshrc": "~/.zshrc",
		"some_conf": "~/some_conf",
		"odd_conf": "/etc/odd_conf"
	},
	"bootstraps": {
		"ripgrep": {
			"brew": {
				"name": "rip-grep"
			},
			"apt": {
				"name": "ripgrep"
			}
		},
		"pyenv": {
			"brew": {
				"name": "pyenv"
			},
			"git": {
				"name": "https://github.com/pyenv/pyenv.git",
				"location": "~/.pyenv"
			}
		}
	},
	"hosts": {
		"host1": {
			"files": ["dot_zshrc", "some_conf"],
			"bootstraps": ["ripgrep"]
		},
		"host2": {
			"files": ["some_conf"],
			"bootstraps": ["ripgrep", "pyenv"]
		},
		"host3": {
			"files": ["dot_zshrc"]
		},
		"host4": {
			"bootstraps": ["pyenv"]
		}
	}
}`

func CreateTempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "test-"+pattern)
	if err != nil {
		t.Fatalf("could not create temp directory %v", err)
	}

	os.Chmod(dir, 0777)

	remove := func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}

	return dir, remove
}

func AssertDirectoryContents(t *testing.T, dir string, want []string) {
	t.Helper()

	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err)

	got := []string{}
	for _, fl := range files {
		got = append(got, fl.Name())
	}

	if len(got) == 0 && len(want) == 0 {
		return
	}

	sort.Strings(got)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("directory listings not equal got %v want %v", got, want)
	}
}

func AssertDirectoryContentsRecursive(t *testing.T, dir string, want []string) {
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

	if len(allPaths) == 0 && len(want) == 0 {
		return
	}

	if !reflect.DeepEqual(allPaths, want) {
		t.Fatalf("directory listings not equal got %v want %v", allPaths, want)
	}
}

func AssertFileContents(t *testing.T, path string, want string) {
	t.Helper()
	contents := ReadFile(t, path)

	if contents != want {
		t.Errorf("incorrect file contents, got %q want %q", contents, want)
	}
}

func ReadFile(t *testing.T, path string) string {
	t.Helper()

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading, %v", err)
	}
	return string(bytes)
}

func WriteData(t *testing.T, path string, data string) {
	t.Helper()

	dirs := filepath.Dir(path)
	err := os.MkdirAll(dirs, 0744)
	require.NoError(t, err)
	err = ioutil.WriteFile(path, []byte(data), 0744)
	require.NoError(t, err)
}

func Touch(t *testing.T, path string) {
	t.Helper()
	WriteData(t, path, "")
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
	os.MkdirAll(godotDir, os.ModePerm)
	WriteData(t, filepath.Join(godotDir, "config.json"), contents)
}

func WriteRepoConf(t *testing.T, dotDir string, contents string) {
	t.Helper()

	path := filepath.Join(dotDir, "config.json")
	WriteData(t, path, contents)
}

func SetupFullConfig(t *testing.T, target string, data *string) (string, string, func()) {
	t.Helper()
	home, dotPath, remove := SetupDirectories(t, target)
	WriteRepoConf(t, dotPath, SAMPLE_CONFIG)

	if data != nil {
		WriteData(t, filepath.Join(dotPath, "templates", "dot_zshrc"), *data)
	}

	return home, dotPath, remove
}

func SetupDirectories(t *testing.T, target string, managers ...string) (string, string, func()) {
	t.Helper()

	home, remove := CreateTempDir(t, "home")

	dotPath := filepath.Join(home, "dotfiles")
	err := os.Mkdir(dotPath, 0744)
	if err != nil {
		t.Errorf("error making dir, %v", err)
	}

	err = os.Mkdir(filepath.Join(dotPath, "templates"), 0744)
	if err != nil {
		t.Errorf("error making dir, %v", err)
	}
	mgrs := []string{"apt", "git"}
	if len(managers) != 0 {
		mgrs = managers
	}
	bytes, err := json.Marshal(map[string]interface{}{
		"target":           target,
		"dotfiles_root":    dotPath,
		"package_managers": mgrs,
		"github_user": "some_user",
	})
	require.NoError(t, err)
	WriteConfig(t, home, string(bytes))

	return home, dotPath, remove
}

func AssertTargetContents(t *testing.T, dotPath string, target string, want []string) {
	t.Helper()

	contents := ReadFile(t, filepath.Join(dotPath, "config.json"))
	value := gjson.Get(contents, fmt.Sprintf("hosts.%v.files", target))
	var actual []string
	value.ForEach(func(key, value gjson.Result) bool {
		actual = append(actual, value.String())
		return true
	})

	if len(actual) == 0 && len(want) == 0 {
		return
	}

	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("target files incorrect, got %v want %v", actual, want)
	}
}

func GetAllFiles(t *testing.T, dotPath string) map[string]string {
	t.Helper()

	contents := ReadFile(t, filepath.Join(dotPath, "config.json"))
	value := gjson.Get(contents, "files")

	actual := make(map[string]string)
	value.ForEach(func(key, value gjson.Result) bool {
		actual[key.String()] = value.String()
		return true
	})

	return actual
}

func AssertAllFiles(t *testing.T, dotPath string, want map[string]string) {
	t.Helper()

	got := GetAllFiles(t, dotPath)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("all files not equal, got %q want %q", got, want)
	}
}

func AssertPresentInAllFiles(t *testing.T, dotpath string, want map[string]string) {
	t.Helper()

	allFiles := GetAllFiles(t, dotpath)
	for key, value := range want {
		actualVal, ok := allFiles[key]
		if !ok || value != actualVal {
			t.Fatalf("k/v pair %q %q not present in all files, %v", key, value, allFiles)
		}
	}
}

func AssertStringInSlice(t *testing.T, str string, slice []string) {
	if StringInSlice(str, slice) {
		return
	}
	t.Fatalf("%v not in slice %v", str, slice)
}

func AssertStringNotInSlice(t *testing.T, str string, slice []string) {
	if !StringInSlice(str, slice) {
		return
	}
	t.Fatalf("%v in slice %v", str, slice)
}

func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func RunCmd(bin string, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func GetReturnCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode(), nil
	} else {
		return -1, err
	}
}
