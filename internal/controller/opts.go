package controller

import (
	"github.com/nicjohnson145/godot/internal/bootstrap"
	"github.com/nicjohnson145/godot/internal/builder"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/repo"
	"github.com/nicjohnson145/godot/internal/util"
)

type SyncOpts struct {
	Force       bool
	NoBootstrap bool
}

type EditOpts struct {
	NoSync bool
}

type ControllerOpts struct {
	HomeDirGetter util.HomeDirGetter
	Repo          repo.Repo
	Config        *config.Config
	Builder       *builder.Builder
	Runner        bootstrap.Runner
	NoGit         bool
}
