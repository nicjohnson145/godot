package lib

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func SelfUpdate(currentVersion string) error {
	conf, err := NewConfig()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	return selfUpdateWithConfig(conf, currentVersion)
}

func selfUpdateWithConfig(conf UserConfig, currentVersion string) error {
	// Always make the log output here visible
	log.SetLevel(log.InfoLevel)

	if currentVersion == "development" {
		log.Warn("Running development build, assuming latest and will not self update")
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
		log.Infof("Current version of %v is latest tag. Nothing to do", currentVersion)
		return nil
	}

	log.Infof("Newer version found, updating to %v", latest)
	godot.Tag = "v" + latest
	if err := godot.Execute(conf, SyncOpts{}, GodotConfig{}); err != nil {
		return fmt.Errorf("error executing self update: %w", err)
	}
	return nil
}
