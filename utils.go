package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/stretchr/testify/require"
)

type Setup struct {
	ConfData interface{}
	Target   string
}

type Paths struct {
	Dotfiles string
	Home     string
	Remove   func()
}

func setupConfigs(t *testing.T, setup Setup) (Paths, Components) {
	t.Helper()

	home, dotpath, remove := help.SetupDirectories(t, setup.Target)
	data, err := json.Marshal(setup.ConfData)
	require.NoError(t, err)
	help.WriteRepoConf(t, dotpath, string(data))

	o := Components{
		HomeDirGetter: &help.TempHomeDir{HomeDir: home},
		Repo: repo.NoopRepo{},
	}
	p := Paths{
		Home:     home,
		Dotfiles: dotpath,
		Remove:   remove,
	}

	return p, o
}

func runCmd(t *testing.T, opts Components, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	t.Helper()

	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)

	main := New(opts)
	main.rootCmd.SetOut(stdOut)
	main.rootCmd.SetErr(stdErr)
	main.rootCmd.SetArgs(args)

	_, err := main.rootCmd.ExecuteC()
	return stdOut, stdErr, err
}

