package lib

import (
	"testing"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"sort"
)

func TestGetExecutors(t *testing.T) {
	targetConf := TargetConfig{
		ConfigFiles: []ConfigFile{
			{Name: "conf-1"},
			{Name: "conf-2"},
			{Name: "conf-3"},
		},
		GithubReleases: []GithubRelease{
			{Name: "github-release-1"},
			{Name: "github-release-2"},
			{Name: "github-release-3"},
		},
		GitRepos: []GitRepo{
			{Name: "git-repo-1"},
			{Name: "git-repo-2"},
			{Name: "git-repo-3"},
		},
		SystemPackages: []SystemPackage{
			{Name: "sys-package-1"},
			{Name: "sys-package-2"},
			{Name: "sys-package-3"},
		},
		Targets: map[string]Target{
			"test1": {
				ConfigFiles: []string{
					"conf-1",
				},
				GithubReleases: []string{
					"github-release-1",
				},
				GitRepos: []string{
					"git-repo-1",
				},
				SystemPackages: []string{
					"sys-package-1",
				},
			},
		},
	}
	userConf := UserConfig{
		Target: "test1",
	}

	executorNames := lo.Map(getExecutors(targetConf, userConf), func(e Executor, i int) string {
		return e.GetName()
	})
	want := []string{
		"conf-1",
		"git-repo-1",
		"github-release-1",
		"sys-package-1",
	}

	sort.Strings(executorNames)

	require.Equal(t, want, executorNames)
}
