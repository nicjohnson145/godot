package lib

import (
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func makeSubDir(t *testing.T, dir string, d string) string {
	t.Helper()
	sub := path.Join(dir, d)
	err := os.MkdirAll(sub, 0744)
	require.NoError(t, err)
	return sub
}

func setupForConfigFile(t *testing.T, name string, content string) UserConfig {
	t.Helper()
	dir := t.TempDir()

	templates := makeSubDir(t, dir, "templates")
	output := makeSubDir(t, dir, "output")
	home := makeSubDir(t, dir, "home")

	err := ioutil.WriteFile(
		path.Join(templates, name),
		[]byte(content),
		0744,
	)
	require.NoError(t, err)

	return UserConfig{
		CloneLocation: dir,
		HomeDir:       home,
		BuildLocation: output,
	}
}

func requireContents(t *testing.T, path string, expected string) {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(b))
}

func cleanFuncsMap(t *testing.T) {
	t.Helper()

	delete(funcs, funcNameIsInstalled)
	delete(funcs, funcNameVaultLookup)
}

func TestConfigFileExecute(t *testing.T) {
	restore, noFatal := NoFatals(t)
	defer noFatal(t)
	defer restore()
	defer cleanFuncsMap(t)

	conf := setupForConfigFile(t, "dot_conf", "Hello from {{ .Target }}")
	conf.Target = "foobar"

	f := ConfigFile{
		Name:        "dot_conf",
		Destination: "~/.config/conf",
	}
	f.Execute(conf, SyncOpts{}, Target{})

	requireContents(t, path.Join(conf.HomeDir, ".config", "conf"), "Hello from foobar")
}

func TestIsInstalled(t *testing.T) {
	restore, noFatal := NoFatals(t)
	defer noFatal(t)
	defer restore()

	content := dedent.Dedent(`
		{{- if IsInstalled "blarg" -}}
		blarg installed
		{{- else -}}
		blarg not installed
		{{- end}}
		`)

	conf := setupForConfigFile(t, "install_conf", content)
	f := ConfigFile{
		Name:        "install_conf",
		Destination: "~/.config/install_conf",
	}
	outPath := path.Join(conf.HomeDir, ".config", "install_conf")

	t.Run("installed", func(t *testing.T) {
		defer cleanFuncsMap(t)
		f.Execute(conf, SyncOpts{}, Target{GithubReleases: []string{"blarg"}})
		requireContents(t, outPath, "blarg installed\n")
	})

	t.Run("not_installed", func(t *testing.T) {
		defer cleanFuncsMap(t)
		f.Execute(conf, SyncOpts{}, Target{GithubReleases: []string{"blarg2"}})
		requireContents(t, outPath, "blarg not installed\n")
	})
}
