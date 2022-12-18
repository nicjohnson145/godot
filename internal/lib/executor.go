package lib

//go:generate go-enum -f $GOFILE -marshal -names -flag

/*
ENUM(
config-file
github-release
git-repo
sys-package
url-download
bundle
golang
go-install
)
*/
type ExecutorType string

type Executor interface {
	Execute(UserConfig, SyncOpts, GodotConfig) error
	Type() ExecutorType
	Namer
}

type Namer interface {
	GetName() string
	SetName(string)
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

func applyOrdering(executors []Executor) []Executor {
	// We cant do a go install until we've installed go, so if we didn't sort these properly then
	// the first configuration run would fail
	installs := []Executor{}
	sortedExecutors := []Executor{}

	for _, e := range executors {
		if e.Type() == ExecutorTypeGoInstall {
			installs = append(installs, e)
		} else {
			sortedExecutors = append(sortedExecutors, e)
		}
	}

	sortedExecutors = append(sortedExecutors, installs...)
	
	return sortedExecutors
}
