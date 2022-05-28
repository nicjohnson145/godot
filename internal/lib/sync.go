package lib

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
