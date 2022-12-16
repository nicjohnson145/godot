package lib

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
)

const (
	targetName = "my-target"
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

	err := os.WriteFile(
		path.Join(templates, name),
		[]byte(content),
		0744,
	)
	require.NoError(t, err)

	return UserConfig{
		CloneLocation: dir,
		HomeDir:       home,
		BuildLocation: output,
		Target:        targetName,
	}
}

func requireContents(t *testing.T, path string, expected string) {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(b))
}

func cleanFuncsMap(t *testing.T) {
	t.Helper()

	delete(funcs, funcNameIsInstalled)
	delete(funcs, funcNameVaultLookup)
}

func TestConfigFileExecute(t *testing.T) {
	defer cleanFuncsMap(t)

	conf := setupForConfigFile(t, "dot_conf", "Hello from {{ .Target }}")

	f := ConfigFile{
		TemplateName: "dot_conf",
		Destination:  "~/.config/conf",
	}
	require.NoError(t, f.Execute(conf, SyncOpts{}, GodotConfig{}))

	requireContents(t, path.Join(conf.HomeDir, ".config", "conf"), fmt.Sprintf("Hello from %v", targetName))
}

func TestIsInstalled(t *testing.T) {
	content := dedent.Dedent(`
		{{- if IsInstalled "blarg" -}}
		blarg installed
		{{- else -}}
		blarg not installed
		{{- end}}
		`)

	conf := setupForConfigFile(t, "install_conf", content)
	f := ConfigFile{
		TemplateName: "install_conf",
		Destination:  "~/.config/install_conf",
	}
	outPath := path.Join(conf.HomeDir, ".config", "install_conf")

	t.Run("installed", func(t *testing.T) {
		defer cleanFuncsMap(t)

		require.NoError(t, f.Execute(
			conf,
			SyncOpts{},
			GodotConfig{
				Executors: map[string]GodotExecutor{
					"blarg": {
						Name: "blarg",
						Type: ExecutorTypeConfigFile,
						Spec: map[string]any{
							"name":        "blarg",
							"destination": "~/.config/blarg",
						},
					},
					"blarg2": {
						Name: "blarg2",
						Type: ExecutorTypeConfigFile,
						Spec: map[string]any{
							"name":        "blarg2",
							"destination": "~/.config/blarg",
						},
					},
				},
				Targets: map[string][]string{
					targetName: []string{"blarg"},
				},
			},
		))
		requireContents(t, outPath, "blarg installed\n")
	})

	t.Run("not_installed", func(t *testing.T) {
		defer cleanFuncsMap(t)

		require.NoError(t, f.Execute(
			conf,
			SyncOpts{},
			GodotConfig{
				Executors: map[string]GodotExecutor{
					"blarg": {
						Name: "blarg",
						Type: ExecutorTypeConfigFile,
						Spec: map[string]any{
							"name":        "blarg",
							"destination": "~/.config/blarg",
						},
					},
					"blarg2": {
						Name: "blarg2",
						Type: ExecutorTypeConfigFile,
						Spec: map[string]any{
							"name":        "blarg2",
							"destination": "~/.config/blarg",
						},
					},
				},
				Targets: map[string][]string{
					targetName: []string{"blarg2"},
				},
			},
		))
		requireContents(t, outPath, "blarg not installed\n")
	})

	t.Run("installed_via_bundle", func(t *testing.T) {
		defer cleanFuncsMap(t)

		require.NoError(t, f.Execute(
			conf,
			SyncOpts{},
			GodotConfig{
				Executors: map[string]GodotExecutor{
					"blarg": {
						Name: "blarg",
						Type: ExecutorTypeConfigFile,
						Spec: map[string]any{
							"name":        "blarg",
							"destination": "~/.config/blarg",
						},
					},
					"blarg-bundle": {
						Name: "blarg-bundle",
						Type: ExecutorTypeBundle,
						Spec: map[string]any{
							"items": []string{"blarg"},
						},
					},
				},
				Targets: map[string][]string{
					targetName: []string{"blarg-bundle"},
				},
			},
		))
		requireContents(t, outPath, "blarg installed\n")
	})

}
