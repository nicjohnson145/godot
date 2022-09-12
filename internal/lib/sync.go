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

func executorsFromOpts(opts SyncOpts) []ExecutorType {
	var executorStrings []string
	if len(opts.Executors) > 0 {
		executorStrings = opts.Executors
	} else {
		executorStrings = ExecutorTypeNames()
	}
	if opts.Quick {
		executorStrings = lo.Filter(executorStrings, func(s string, _ int) bool { return s != ExecutorTypeSysPackages.String() })
	}

	return lo.Map(executorStrings, func(s string, _ int) ExecutorType {
		val, err := ParseExecutorType(s)
		if err != nil {
			log.Fatal(err)
		}
		return val
	})
}

func syncFromConf(userConf UserConfig, opts SyncOpts) {
	EnsureDotfilesRepo(userConf)
	tConf := NewTargetConfig(userConf)
	executors := getExecutors(tConf, userConf)
	executorTypes := executorsFromOpts(opts)

	for _, ex := range executors {
		if lo.Contains(opts.Ignore, ex.GetName()) {
			log.Infof("Ignoring %v due to command line arg", ex.GetName())
			continue
		}
		if !lo.Contains(executorTypes, ex.Type()) {
			log.Infof("Skipping %v due to command line arg", ex.GetName())
			continue
		}

		ex.Execute(userConf, opts, tConf)
	}
}
