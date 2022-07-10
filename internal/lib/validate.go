package lib

import (
	mset "github.com/deckarep/golang-set/v2"
	log "github.com/sirupsen/logrus"
)

func Validate(filepath string) {
	exists, err := pathExists(filepath)
	if err != nil {
		log.Fatalf("Error checking existance of file: %v", err)
	}

	if !exists {
		log.Fatalf("Path %v does not exist", filepath)
	}

	validateConfig(NewTargetConfigFromPath(filepath))
}

func validateConfig(targetConfig TargetConfig) {
	names := mset.NewSet[string]()

	hasDupes := false
	for _, n := range targetConfig.ConfigFiles {
		if names.Contains(n.GetName()) {
			log.Errorf("Duplicate name %v found", n.GetName())
			hasDupes = true
		}
		names.Add(n.GetName())
	}
	for _, n := range targetConfig.GithubReleases {
		if names.Contains(n.GetName()) {
			log.Errorf("Duplicate name %v found", n.GetName())
			hasDupes = true
		}
		names.Add(n.GetName())
	}
	for _, n := range targetConfig.GitRepos {
		if names.Contains(n.GetName()) {
			log.Errorf("Duplicate name %v found", n.GetName())
			hasDupes = true
		}
		names.Add(n.GetName())
	}
	for _, n := range targetConfig.SystemPackages {
		if names.Contains(n.GetName()) {
			log.Errorf("Duplicate name %v found", n.GetName())
			hasDupes = true
		}
		names.Add(n.GetName())
	}
	for _, n := range targetConfig.Bundles {
		if names.Contains(n.GetName()) {
			log.Errorf("Duplicate name %v found", n.GetName())
			hasDupes = true
		}
		names.Add(n.GetName())
	}

	if hasDupes {
		log.Fatal("Duplicate names are not allowed")
	}
}

