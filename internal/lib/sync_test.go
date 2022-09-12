package lib

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecutorsFromOpts(t *testing.T) {
	testData := []struct {
		name string
		opts SyncOpts
		want []ExecutorType
	}{
		{
			name: "nothing_given",
			opts: SyncOpts{},
			want: []ExecutorType{
				ExecutorTypeConfigFiles,
				ExecutorTypeGithubReleases,
				ExecutorTypeGitRepos,
				ExecutorTypeSysPackages,
				ExecutorTypeUrlDownloads,
			},
		},
		{
			name: "executors_given",
			opts: SyncOpts{
				Executors: []string{
					ExecutorTypeGitRepos.String(),
					ExecutorTypeConfigFiles.String(),
				},
			},
			want: []ExecutorType{
				ExecutorTypeGitRepos,
				ExecutorTypeConfigFiles,
			},
		},
		{
			name: "nothing_given_quick",
			opts: SyncOpts{
				Quick: true,
			},
			want: []ExecutorType{
				ExecutorTypeConfigFiles,
				ExecutorTypeGithubReleases,
				ExecutorTypeGitRepos,
				ExecutorTypeUrlDownloads,
			},
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			got := executorsFromOpts(tc.opts)
			require.Equal(t, tc.want, got)
		})
	}
}
