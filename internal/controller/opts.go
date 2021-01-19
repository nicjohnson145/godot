package controller

import (
	"github.com/nicjohnson145/godot/internal/builder"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/nicjohnson145/godot/internal/util"
)

type SyncOpts struct {
	Force bool
	NoGit bool
}

type EditOpts struct {
	NoGit  bool
	NoSync bool
}

type ImportOpts struct {
	NoGit bool
	NoAdd bool
}

type AddOpts struct {
	NoGit bool
}

type RemoveOpts struct {
	NoGit bool
}

type ControllerOpts struct {
	HomeDirGetter util.HomeDirGetter
	Repo          repo.Repo
	Config        *config.Config
	Builder       *builder.Builder
}
