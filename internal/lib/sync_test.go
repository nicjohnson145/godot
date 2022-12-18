package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
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
				ExecutorTypeConfigFile,
				ExecutorTypeGithubRelease,
				ExecutorTypeGitRepo,
				ExecutorTypeSysPackage,
				ExecutorTypeUrlDownload,
				ExecutorTypeBundle,
				ExecutorTypeGolang,
				ExecutorTypeGoInstall,
			},
		},
		{
			name: "executors_given",
			opts: SyncOpts{
				Executors: []string{
					ExecutorTypeGitRepo.String(),
					ExecutorTypeConfigFile.String(),
				},
			},
			want: []ExecutorType{
				ExecutorTypeGitRepo,
				ExecutorTypeConfigFile,
			},
		},
		{
			name: "nothing_given_quick",
			opts: SyncOpts{
				Quick: true,
			},
			want: []ExecutorType{
				ExecutorTypeConfigFile,
				ExecutorTypeGithubRelease,
				ExecutorTypeGitRepo,
				ExecutorTypeUrlDownload,
				ExecutorTypeBundle,
				ExecutorTypeGolang,
				ExecutorTypeGoInstall,
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
