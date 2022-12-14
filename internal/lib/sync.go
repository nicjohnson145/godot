package lib

import (
	"fmt"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

type SyncOpts struct {
	Quick     bool
	Ignore    []string
	NoVault   bool
	Executors []string
}

func Sync(opts SyncOpts) error {
	conf, err := NewOverrideableConfig(ConfigOverrides{
		IgnoreVault: opts.NoVault,
	})
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	return syncFromConf(
		conf,
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
		executorStrings = lo.Filter(executorStrings, func(s string, _ int) bool { return s != ExecutorTypeSysPackage.String() })
	}

	return lo.Map(executorStrings, func(s string, _ int) ExecutorType {
		val, err := ParseExecutorType(s)
		if err != nil {
			log.Fatal(err)
		}
		return val
	})
}

func syncFromConf(userConf UserConfig, opts SyncOpts) error {
	if err := EnsureDotfilesRepo(userConf); err != nil {
		return fmt.Errorf("error ensuring dotfiles repo: %w", err)
	}
	godotConf, err := NewGodotConfigFromUserConfig(userConf)
	if err != nil {
		return fmt.Errorf("error loading godot config; %w", err)
	}
	executors, err := godotConf.ExecutorsForTarget(userConf.Target)
	if err != nil {
		return fmt.Errorf("error fetching target configuration: %w", err)
	}
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

		if err := ex.Execute(userConf, opts, godotConf); err != nil {
			return fmt.Errorf("error during execution of %v: %w", ex.GetName(), err)
		}
	}

	return nil
}
