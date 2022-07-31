package lib

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestConfigFileExecute(t *testing.T) {
	restore, noFatal := NoFatals(t)
	defer noFatal(t)
	defer restore()

	dir := t.TempDir()

	makeSubDir := func(d string) string {
		sub := path.Join(dir, d)
		err := os.MkdirAll(sub, 0744)
		require.NoError(t, err)
		return sub
	}

	templates := makeSubDir("templates")
	output := makeSubDir("output")
	home := makeSubDir("home")

	err := ioutil.WriteFile(
		path.Join(templates, "dot_conf"),
		[]byte("Hello from {{ .Target }}"),
		0744,
	)

	f := ConfigFile{
		Name:        "dot_conf",
		Destination: "~/.config/conf",
	}
	f.Execute(UserConfig{
		CloneLocation: dir,
		HomeDir:       home,
		BuildLocation: output,
		Target:        "foobar",
	}, SyncOpts{}, Target{})

	b, err := ioutil.ReadFile(path.Join(home, ".config", "conf"))
	require.NoError(t, err)
	require.Equal(t, "Hello from foobar", string(b))
}
