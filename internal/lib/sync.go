package lib

import (
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

type SyncOpts struct {
	Quick     bool
	Ignore    []string
	NoVault   bool
	Executors []string
}

func Sync(opts SyncOpts) {
	syncFromConf(
		NewOverrideableConfig(ConfigOverrides{
			IgnoreVault: opts.NoVault,
		}),
		opts,
	)
}

func syncFromConf(userConf UserConfig, opts SyncOpts) {
	EnsureDotfilesRepo(userConf)
	executors := getExecutors(
		NewTargetConfig(userConf),
		userConf,
	)

	executorsSpecified := len(opts.Executors) != 0

	for _, ex := range executors {
		if lo.Contains(opts.Ignore, ex.GetName()) {
			log.Infof("Ignoring %v due to command line arg", ex.GetName())
			continue
		}
		if executorsSpecified && !lo.Contains(opts.Executors, ex.Type()) {
			log.Infof("Skipping %v due to executors command line arg", ex.GetName())
			continue
		}

		ex.Execute(userConf, opts)
	}
}
