package lib

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

func TestGetExecutors(t *testing.T) {
	targetConf := TargetConfig{
		ConfigFiles: []ConfigFile{
			{Name: "conf-1"},
			{Name: "conf-2"},
			{Name: "conf-3"},
			{Name: "conf-4"},
		},
		GithubReleases: []GithubRelease{
			{Name: "github-release-1"},
			{Name: "github-release-2"},
			{Name: "github-release-3"},
			{Name: "github-release-4"},
		},
		GitRepos: []GitRepo{
			{Name: "git-repo-1"},
			{Name: "git-repo-2"},
			{Name: "git-repo-3"},
			{Name: "git-repo-4"},
		},
		SystemPackages: []SystemPackage{
			{Name: "sys-package-1"},
			{Name: "sys-package-2"},
			{Name: "sys-package-3"},
			{Name: "sys-package-4"},
		},
		Bundles: []Bundle{
			{
				Name: "bundle-1",
				Target: Target{
					ConfigFiles:    []string{"conf-1"},
					GithubReleases: []string{"github-release-1"},
					GitRepos:       []string{"git-repo-1"},
					SystemPackages: []string{"sys-package-1"},
				},
			},
			{
				Name: "bundle-2",
				Target: Target{
					ConfigFiles:    []string{"conf-2"},
					GithubReleases: []string{"github-release-2"},
					GitRepos:       []string{"git-repo-2"},
					SystemPackages: []string{"sys-package-2"},
					Bundles:        []string{"bundle-3"}, // Woooo, recursion
				},
			},
			{
				Name: "bundle-3",
				Target: Target{
					ConfigFiles:    []string{"conf-3"},
					GithubReleases: []string{"github-release-3"},
				},
			},
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
				Bundles: []string{
					"bundle-1", // This should be all duplicates
					"bundle-2",
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
		"conf-2",
		"conf-3",
		"git-repo-1",
		"git-repo-2",
		"github-release-1",
		"github-release-2",
		"github-release-3",
		"sys-package-1",
		"sys-package-2",
	}

	sort.Strings(executorNames)

	require.Equal(t, want, executorNames)
}
