package lib

type SyncOpts struct {
	Quick bool
}

func Sync(opts SyncOpts) {
	syncFromConf(NewConfig(), opts)
}

func syncFromConf(userConf UserConfig, opts SyncOpts) {
	EnsureDotfilesRepo(userConf)
	executors := getExecutors(
		NewTargetConfig(userConf),
		userConf,
	)

	for _, ex := range executors {
		ex.Execute(userConf, opts)
	}
}
