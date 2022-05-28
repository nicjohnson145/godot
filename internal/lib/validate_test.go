package lib

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {

	t.Run("duplicateNames", func(t *testing.T) {
		restore, hasFatal := HasFatals()
		defer hasFatal(t)
		defer restore()

		duplicateNames := TargetConfig{
			ConfigFiles: []ConfigFile{
				{Name: "conf-1"},
			},
			GithubReleases: []GithubRelease{
				{Name: "github-release-1"},
				{Name: "github-release-2"},
			},
			Bundles: []Bundle{
				{
					Name: "github-release-1",
				},
			},
		}
		validateConfig(duplicateNames)
	})

	t.Run("no-duplicates", func(t *testing.T) {
		restore, noFatal := NoFatals()
		defer noFatal(t)
		defer restore()

		duplicateNames := TargetConfig{
			ConfigFiles: []ConfigFile{
				{Name: "conf-1"},
			},
			GithubReleases: []GithubRelease{
				{Name: "github-release-1"},
				{Name: "github-release-2"},
			},
			Bundles: []Bundle{
				{
					Name: "bundle-1",
				},
			},
		}
		validateConfig(duplicateNames)
	})
}
