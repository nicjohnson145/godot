package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/nicjohnson145/godot/internal/lib"
	"github.com/stretchr/testify/require"
)

type Setup struct {
	ConfData        interface{}
	Target          string
	TemplateContent map[string]string
	PackageManagers []string
}

type Paths struct {
	Dotfiles string
	Home     string
	Remove   func()
}

func setupConfigs(t *testing.T, setup Setup) (Paths, lib.ControllerOpts) {
	t.Helper()

	home, dotpath, remove := help.SetupDirectories(t, setup.Target, setup.PackageManagers...)
	data, err := json.Marshal(setup.ConfData)
	require.NoError(t, err)
	help.WriteRepoConf(t, dotpath, string(data))

	for tmp, content := range setup.TemplateContent {
		help.WriteData(t, filepath.Join(dotpath, "templates", tmp), content)
	}

	o := lib.ControllerOpts{
		HomeDirGetter: &help.TempHomeDir{HomeDir: home},
		Repo:          lib.NoopRepo{},
	}
	p := Paths{
		Home:     home,
		Dotfiles: dotpath,
		Remove:   remove,
	}

	return p, o
}

func runCmd(t *testing.T, opts lib.ControllerOpts, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	t.Helper()

	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)

	main := New(opts)
	main.rootCmd.SetOut(stdOut)
	main.rootCmd.SetErr(stdErr)
	main.rootCmd.SetArgs(args)

	err := main.rootCmd.Execute()
	return stdOut, stdErr, err
}
