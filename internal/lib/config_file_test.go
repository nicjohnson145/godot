package lib

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
)

const (
	targetName = "my-target"
)

func setupForConfigFile(t *testing.T, name string, content string) UserConfig {
	t.Helper()

	root := buildDirectoryStructure(t, map[string]string{
		"templates/" + name: content,
		"output/": "",
		"home/": "",
	})

	return UserConfig{
		CloneLocation: root,
		HomeDir:       filepath.Join(root, "home"),
		BuildLocation: filepath.Join(root, "output"),
		Target:        targetName,
	}
}

func TestConfigFileExecute(t *testing.T) {
	t.Run("with templates", func(t *testing.T) {
		defer cleanFuncsMap(t)
		conf := setupForConfigFile(t, "dot_conf", "Hello from {{ .Target }}")

		f := ConfigFile{
			TemplateName: "dot_conf",
			Destination:  "~/.config/conf",
		}
		require.NoError(t, f.Execute(conf, SyncOpts{}, GodotConfig{}))

		requireContents(t, path.Join(conf.HomeDir, ".config", "conf"), fmt.Sprintf("Hello from %v", targetName))
	})

	t.Run("without templates", func(t *testing.T) {
		defer cleanFuncsMap(t)
		conf := setupForConfigFile(t, "dot_conf", "Hello from {{ .Target }}")

		f := ConfigFile{
			TemplateName: "dot_conf",
			Destination:  "~/.config/conf",
			NoTemplate: true,
		}
		require.NoError(t, f.Execute(conf, SyncOpts{}, GodotConfig{}))

		requireContents(t, path.Join(conf.HomeDir, ".config", "conf"), "Hello from {{ .Target }}")
	})
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
