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
	tConf := NewTargetConfig(userConf)
	executors := getExecutors(tConf, userConf)

	shouldFilter := len(opts.Executors) != 0 || opts.Quick
	executorsTypes := lo.Map(opts.Executors, func(t string, _ int) ExecutorType {
		val, err := ParseExecutorType(t)
		if err != nil {
			log.Fatal(err)
		}
		return val
	})
	if opts.Quick {
		if !lo.Contains(executorsTypes, ExecutorTypeSysPackages) {
			executorsTypes = append(executorsTypes, ExecutorTypeSysPackages)
		}
	}

	for _, ex := range executors {
		if lo.Contains(opts.Ignore, ex.GetName()) {
			log.Infof("Ignoring %v due to command line arg", ex.GetName())
			continue
		}
		if shouldFilter && !lo.Contains(executorsTypes, ex.Type()) {
			log.Infof("Skipping %v due to command line arg", ex.GetName())
			continue
		}

		ex.Execute(userConf, opts, tConf)
	}
}
