package lib

import (
	"fmt"

	"github.com/rs/zerolog"
)

type SelfUpdateOpts struct {
	Logger zerolog.Logger
	CurrentVersion string
	IgnoreVault bool
}

func SelfUpdate(opts SelfUpdateOpts) error {
	conf, err := NewOverrideableConfig(ConfigOverrides{
		IgnoreVault: opts.IgnoreVault,
	})
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	return selfUpdateWithConfig(conf, opts.CurrentVersion, opts.Logger)
}

func selfUpdateWithConfig(conf UserConfig, currentVersion string, logger zerolog.Logger) error {
	godot := GithubRelease{
		Name:           "godot",
		Repo:           "nicjohnson145/godot",
		IsArchive:      false,
		AssetPatterns: map[string]map[string]string{
			"darwin": {
				"amd64": "^godot_darwin_amd64$",
				"arm64": "^godot_darwin_arm64$",
			},
			"linux": {
				"amd64": "^godot_linux_amd64$",
				"arm64": "^godot_linux_arm64$",
			},
			"windows": {
				"amd64": "^godot_windows_amd64.exe$",
			},
		},
	}

	latest, err := godot.GetLatestRelease(conf)
	if err != nil {
		return fmt.Errorf("error determining latest release: %w", err)
	}
	latest = latest[1:]

	if latest == currentVersion {
		logger.Info().Str("version", currentVersion).Msg("current version is latest tag. nothing to do")
		return nil
	}

	logger.Info().Str("version", latest).Msg("newer version found, updating")
	godot.Tag = "v" + latest
	if err := godot.Execute(conf, SyncOpts{}, GodotConfig{}); err != nil {
		return fmt.Errorf("error executing self update: %w", err)
	}
	return nil
}
