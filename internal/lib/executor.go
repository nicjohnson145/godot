package lib

import (
	log "github.com/sirupsen/logrus"
)

type Executor interface {
	Execute(UserConfig, SyncOpts)
	Type() string
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

func collectNamedOptions[T Executor](selectedOptions []string, allOptions []T, str string) []Executor {
	executors := []Executor{}
	for _, name := range selectedOptions {
		if idx, found := getByName(name, allOptions); !found {
			log.Fatalf("Unknown %v %v", str, name)
		} else {
			executors = append(executors, allOptions[idx])
		}
	}
	return executors
}

func toPtrList[T any](list []T) []*T {
	ptrs := []*T{}
	for _, i := range list {
		x := i
		ptrs = append(ptrs, &x)
	}
	return ptrs
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

	executors = append(
		executors,
		collectNamedOptions(target.ConfigFiles, toPtrList(targetConf.ConfigFiles), "config file")...,
	)
	executors = append(
		executors,
		collectNamedOptions(target.GithubReleases, toPtrList(targetConf.GithubReleases), "github release")...,
	)
	executors = append(
		executors,
		collectNamedOptions(target.GitRepos, toPtrList(targetConf.GitRepos), "git repository")...,
	)
	executors = append(
		executors,
		collectNamedOptions(target.SystemPackages, toPtrList(targetConf.SystemPackages), "system package")...,
	)

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
