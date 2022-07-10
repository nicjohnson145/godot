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
	if dupes := checkNames(names, toPtrList(targetConfig.ConfigFiles)); dupes {
		hasDupes = true
	}
	if dupes := checkNames(names, toPtrList(targetConfig.GithubReleases)); dupes {
		hasDupes = true
	}
	if dupes := checkNames(names, toPtrList(targetConfig.GitRepos)); dupes {
		hasDupes = true
	}
	if dupes := checkNames(names, toPtrList(targetConfig.SystemPackages)); dupes {
		hasDupes = true
	}
	if dupes := checkNames(names, targetConfig.Bundles); dupes {
		hasDupes = true
	}

	if hasDupes {
		log.Fatal("Duplicate names are not allowed")
	}
}

func checkNames[T Namer](names mset.Set[string], options []T) bool {
	foundErr := false
	for _, f := range options {
		if names.Contains(f.GetName()) {
			log.Errorf("Duplicate name %v found", f.GetName())
			foundErr = true
		}
		names.Add(f.GetName())
	}
	return foundErr
}
