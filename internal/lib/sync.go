package lib

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type SyncOpts struct {
	Quick     bool
	Ignore    []string
	NoVault   bool
	Executors []string
}

func (s *SyncOpts) Validate() error {
	for _, ex := range s.Executors {
		if _, err := ParseExecutorType(ex); err != nil {
			return err
		}
	}
	return nil
}

func Sync(opts SyncOpts, logger zerolog.Logger) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	conf, err := NewOverrideableConfig(ConfigOverrides{
		IgnoreVault: opts.NoVault,
	})
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	return syncFromConf(
		conf,
		opts,
		logger,
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
			// There should be no reason for this to fail right now, just panic
			panic(fmt.Errorf("unexpected error parsing executor: %w", err).Error())
		}
		return val
	})
}

func syncFromConf(userConf UserConfig, opts SyncOpts, logger zerolog.Logger) error {
	if err := ensureDotfilesRepo(userConf, logger); err != nil {
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
			logger.Debug().Str("name", ex.GetName()).Msg("ignoring due to command line arg")
			continue
		}
		if !lo.Contains(executorTypes, ex.Type()) {
			logger.Debug().Str("name", ex.GetName()).Msg("ignoring due to command line arg")
			continue
		}

		ex.SetLogger(logger)
		if err := ex.Execute(userConf, opts, godotConf); err != nil {
			return fmt.Errorf("error during execution of %v: %w", ex.GetName(), err)
		}
	}

	return nil
}

func ensureDotfilesRepo(conf UserConfig, logger zerolog.Logger) error {
	dotfiles := GitRepo{
		URL:         conf.DotfilesURL,
		Location:    conf.CloneLocation,
		Private:     true,
		TrackLatest: true,
	}
	dotfiles.SetLogger(logger)
	if err := dotfiles.Execute(conf, SyncOpts{}, GodotConfig{}); err != nil {
		return fmt.Errorf("error ensuring dotfiles repo: %w", err)
	}
	return nil
}
