package lib

import (
	log "github.com/sirupsen/logrus"
)

type Executor interface {
	Execute(UserConfig)
	GetName() string
}

func Sync() {
	syncFromConf(NewConfig())
}

func syncFromConf(userConf UserConfig) {
	EnsureDotfilesRepo(userConf)
	executors := getExecutors(
		NewTargetConfig(userConf),
		userConf,
	)

	for _, ex := range executors {
		ex.Execute(userConf)
	}
}

func getByName[T Executor](name string, objs []T) (int, bool) {
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

func getExecutors(targetConf TargetConfig, userConf UserConfig) []Executor {
	target, ok := targetConf.Targets[userConf.Target]
	if !ok {
		log.Fatalf("Target %v has no configuration", userConf.Target)
	}

	executors := []Executor{}

	executors = append(
		executors,
		collectNamedOptions(target.ConfigFiles, targetConf.ConfigFiles, "config file")...,
	)
	executors = append(
		executors,
		collectNamedOptions(target.GithubReleases, targetConf.GithubReleases, "github release")...,
	)
	executors = append(
		executors,
		collectNamedOptions(target.GitRepos, targetConf.GitRepos, "git repository")...,
	)

	return executors
}
