package lib

import (
	log "github.com/sirupsen/logrus"
	"github.com/samber/lo"
)

type Executor interface {
	Execute(UserConfig, SyncOpts)
	Namer
}

type Namer interface {
	GetName() string
}

func getByName[T Namer](name string, objs []T) (int, bool) {
	for i, o := range objs {
		if o.GetName() == name {
			return i, true
		}
	}
	return -1, false
}


func getExecutors(targetConf TargetConfig, userConf UserConfig) []Executor {
	target, ok := targetConf.Targets[userConf.Target]
	if !ok {
		log.Fatalf("Target %v has no configuration", userConf.Target)
	}

	return getExecutorsFromTarget(target, targetConf)
}

func getExecutorsFromTarget(target Target, targetConf TargetConfig) []Executor {
	executors := []Executor{}

	for _, name := range target.ConfigFiles {
		conf, ok := lo.Find(targetConf.ConfigFiles, func(c ConfigFile) bool {
			return c.GetName() == name
		})
		if !ok {
			log.Fatalf("Unknown ConfigFile %v", name)
		}
		executors = append(executors, &conf)
	}

	for _, name := range target.GithubReleases {
		conf, ok := lo.Find(targetConf.GithubReleases, func(c GithubRelease) bool {
			return c.GetName() == name
		})
		if !ok {
			log.Fatalf("Unknown GithubRelease %v", name)
		}
		executors = append(executors, &conf)
	}

	for _, name := range target.GitRepos {
		conf, ok := lo.Find(targetConf.GitRepos, func(c GitRepo) bool {
			return c.GetName() == name
		})
		if !ok {
			log.Fatalf("Unknown GitRepo %v", name)
		}
		executors = append(executors, &conf)
	}

	for _, name := range target.SystemPackages {
		conf, ok := lo.Find(targetConf.SystemPackages, func(c SystemPackage) bool {
			return c.GetName() == name
		})
		if !ok {
			log.Fatalf("Unknown SystemPackage %v", name)
		}
		executors = append(executors, &conf)
	}

	for _, bundleName := range target.Bundles {
		idx, found := getByName(bundleName, targetConf.Bundles)
		if !found {
			log.Fatalf("Bundle %v not found", bundleName)
		}

		executors = append(
			executors,
			getExecutorsFromTarget(targetConf.Bundles[idx].ToTarget(), targetConf)...,
		)
	}

	// Deduplicate executors, in case multiple bundles require the same thing
	return deduplicate(executors)
}

func deduplicate(executors []Executor) []Executor {
	found := map[string]struct{}{}

	ret := []Executor{}

	for _, e := range executors {
		if _, ok := found[e.GetName()]; !ok {
			found[e.GetName()] = struct{}{}
			ret = append(ret, e)
		}
	}

	return ret
}
