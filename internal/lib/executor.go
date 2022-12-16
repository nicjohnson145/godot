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
