package controller

type SyncOpts struct {
	Force bool
	NoGit bool
}

type EditOpts struct {
	NoGit bool
	NoSync bool
}

type ImportOpts struct {
	NoGit bool
	NoAdd bool
}

type AddOpts struct {
	NoGit bool
}

type RemoveOpts struct {
	NoGit bool
}
