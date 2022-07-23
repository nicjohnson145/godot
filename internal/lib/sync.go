package lib

import (
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

type SyncOpts struct {
	Quick   bool
	Ignore  []string
	NoVault bool
}

func Sync(opts SyncOpts) {
	syncFromConf(NewConfig(), opts)
}

func syncFromConf(userConf UserConfig, opts SyncOpts) {
	EnsureDotfilesRepo(userConf)
	executors := getExecutors(
		NewTargetConfig(userConf),
		userConf,
	)

	for _, ex := range executors {
		if lo.Contains(opts.Ignore, ex.GetName()) {
			log.Infof("Ignoring %v due to command line arg", ex.GetName())
			continue
		}
		ex.Execute(userConf, opts)
	}
}
