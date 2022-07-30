package lib

import (
	log "github.com/sirupsen/logrus"
)

func SelfUpdate(currentVersion string) {
	selfUpdateWithConfig(NewConfig(), currentVersion)
}

func selfUpdateWithConfig(conf UserConfig, currentVersion string) {
	// Always make the log output here visible
	log.SetLevel(log.InfoLevel)

	if currentVersion == "development" {
		log.Warn("Running development build, assuming latest and will not self update")
		return
	}

	godot := GithubRelease{
		Name:           "godot",
		Repo:           "nicjohnson145/godot",
		IsArchive:      false,
		MacPattern:     "^godot_darwin_amd64$",
		LinuxPattern:   "^godot_linux_amd64$",
		WindowsPattern: "^godot_windows_amd64.exe$",
	}
	latest := godot.GetLatestRelease(conf)[1:]

	if latest == currentVersion {
		log.Infof("Current version of %v is latest tag. Nothing to do", currentVersion)
		return
	}

	log.Infof("Newer version found, updating to %v", latest)
	godot.Tag = "v" + latest
	godot.Execute(conf, SyncOpts{})
}
