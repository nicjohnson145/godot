package file

import (
	"testing"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/lib"
	"io/ioutil"
	"github.com/stretchr/testify/require"
	"path"
	"os"
)

func TestExecute(t *testing.T) {
	restore, noFatal := lib.NoFatals()
	defer restore()
	defer noFatal(t)

	dir, err := ioutil.TempDir("", "godot-config-file-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	makeSubDir := func(d string) string {
		sub := path.Join(dir, d)
		err = os.MkdirAll(sub, 0744)
		require.NoError(t, err)
		return sub
	}

	templates := makeSubDir("templates")
	output := makeSubDir("output")
	home := makeSubDir("home")

	err = ioutil.WriteFile(
		path.Join(templates, "dot_conf"),
		[]byte("Hello from {{ .Target }}"),
		0744,
	)

	f := ConfigFile{
		Name: "dot_conf",
		Destination: "~/.config/conf",
	}
	f.Execute(config.UserConfig{
		CloneLocation: dir,
		HomeDir: home,
		BuildLocation: output,
		Target: "foobar",
	})

	b, err := ioutil.ReadFile(path.Join(home, ".config", "conf"))
	require.NoError(t, err)
	require.Equal(t, "Hello from foobar", string(b))
}
