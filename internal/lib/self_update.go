package lib

import (
	"fmt"

	"github.com/rs/zerolog"
)

func SelfUpdate(currentVersion string, logger zerolog.Logger) error {
	conf, err := NewConfig()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	return selfUpdateWithConfig(conf, currentVersion, logger)
}

func selfUpdateWithConfig(conf UserConfig, currentVersion string, logger zerolog.Logger) error {
	if currentVersion == "development" {
		logger.Warn().Msg("Running development build, assuming latest and will not self update")
		return nil
	}

	godot := GithubRelease{
		Name:           "godot",
		Repo:           "nicjohnson145/godot",
		IsArchive:      false,
		MacPattern:     "^godot_darwin_amd64$",
		LinuxPattern:   "^godot_linux_amd64$",
		WindowsPattern: "^godot_windows_amd64.exe$",
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
