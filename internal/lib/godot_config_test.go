package lib

import (
	"os"
	"path"
	"sort"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestExecutorsForTarget(t *testing.T) {
	setupConf := func(t *testing.T, dataPath string) string {
		t.Helper()

		confBytes, err := os.ReadFile(dataPath)

		dir := t.TempDir()
		confPath := path.Join(dir, "config.yaml")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(confPath, confBytes, 0777))

		return confPath
	}

	t.Run("good_config", func(t *testing.T) {
		confPath := setupConf(t, "./testdata/godot-config/good_config.yaml")

		conf, err := NewGodotConfig(confPath)
		require.NoError(t, err)

		target1Executors, err := conf.ExecutorsForTarget("target1")
		require.NoError(t, err)

		target1Names := lo.Map(target1Executors, func(e Executor, _ int) string {
			return e.GetName()
		})
		sort.Strings(target1Names)
		require.Equal(
			t,
			[]string{
				"bundle1",
				"bundle2",
				"conf1",
				"conf2",
				"conf3",
				"conf4",
			},
			target1Names,
		)
	})

	t.Run("bad_executor_name", func(t *testing.T) {
		confPath := setupConf(t, "./testdata/godot-config/bad_executor_name.yaml")
		_, err := NewGodotConfig(confPath)
		require.Error(t, err)
		// Both errors should be reported
		require.Contains(t, err.Error(), "error with target target1: unknown executor some-bad-executor")
		require.Contains(t, err.Error(), "error with target target1: unknown executor some-other-bad-executor")
	})

	t.Run("bad_executor_type", func(t *testing.T) {
		confPath := setupConf(t, "./testdata/godot-config/bad_executor_type.yaml")
		_, err := NewGodotConfig(confPath)
		require.Error(t, err)
		// should yell about a failure to unmarshal the type
		require.Contains(t, err.Error(), "some-bad-type")
	})

	t.Run("bad_bundle_item", func(t *testing.T) {
		confPath := setupConf(t, "./testdata/godot-config/bad_executor_in_bundle.yaml")
		_, err := NewGodotConfig(confPath)
		require.Error(t, err)
		// should yell about a failure to unmarshal the type
		require.Contains(t, err.Error(), "missingconf")
	})
}

func ensureExecutorParsed[T any](t *testing.T, gEx GodotExecutor, expected T) {
	t.Helper()

	ex, err := gEx.AsExecutor()
	require.NoError(t, err)

	concrete, ok := ex.(T)
	require.True(t, ok, "unable to cast to %T", expected)
	require.Equal(t, expected, concrete)
}

func TestAsExecutor(t *testing.T) {
	t.Run("config_file", func(t *testing.T) {
		ensureExecutorParsed(
			t,
			GodotExecutor{
				Name: "e1",
				Type: ExecutorTypeConfigFile,
				Spec: map[string]any{
					"template-name": "foo-tmpl",
					"destination": "foo-dest",
				},
			},
			&ConfigFile{
				Name: "e1",
				TemplateName: "foo-tmpl",
				Destination: "foo-dest",
			},
		)
	})

	t.Run("git_repo", func(t *testing.T) {
		ensureExecutorParsed(
			t,
			GodotExecutor{
				Name: "e1",
				Type: ExecutorTypeGitRepo,
				Spec: map[string]any{
					"url": "https://foo.com",
					"location": "foo-location",
					"private": true,
					"ref": map[string]any{
						"tag": "v1.0.0",
						"commit": "abcd-efgh",
					},
				},
			},
			&GitRepo{
				Name: "e1",
				URL: "https://foo.com",
				Location: "foo-location",
				Private: true,
				Ref: Ref{
					Tag: "v1.0.0",
					Commit: "abcd-efgh",
				},
			},
		)
	})

	t.Run("github_release", func(t *testing.T) {
		ensureExecutorParsed(
			t,
			GodotExecutor{
				Name: "e1",
				Type: ExecutorTypeGithubRelease,
				Spec: map[string]any{
					"repo": "http://some-repo.git",
					"tag": "LATEST",
					"is-archive": true,
					"regex": "abcd",
					"asset-patterns": map[string]any{
						"darwin": map[string]any{
							"amd64": "mac-amd64-pattern",
						},
					},
				},
			},
			&GithubRelease{
				Name: "e1",
				Repo: "http://some-repo.git",
				Tag: "LATEST",
				IsArchive: true,
				Regex: "abcd",
				AssetPatterns: map[string]map[string]string{
					"darwin": {
						"amd64": "mac-amd64-pattern",
					},
				},
			},
		)
	})

	t.Run("sys_package", func(t *testing.T) {
		ensureExecutorParsed(
			t,
			GodotExecutor{
				Name: "e1",
				Type: ExecutorTypeSysPackage,
				Spec: map[string]any{
					"apt": "apt-name",
					"brew": "brew-name",
				},
			},
			&SystemPackage{
				Name: "e1",
				AptName: "apt-name",
				BrewName: "brew-name",
			},
		)
	})

	t.Run("url_download", func(t *testing.T) {
		ensureExecutorParsed(
			t,
			GodotExecutor{
				Name: "e1",
				Type: ExecutorTypeUrlDownload,
				Spec: map[string]any{
					"tag": "some-tag",
					"mac-url": "mac-url",
					"linux-url": "linux-url",
					"windows-url": "windows-url",
				},
			},
			&UrlDownload{
				Name: "e1",
				Tag: "some-tag",
				MacUrl: "mac-url",
				LinuxUrl: "linux-url",
				WindowsUrl: "windows-url",
			},
		)
	})
}

func TestAsExecutorSmokes(t *testing.T) {
	for _, kind := range _ExecutorTypeValue {
		if kind == ExecutorTypeBundle {
			continue
		}
		gEx := GodotExecutor{
			Name: "ex",
			Type: kind,
			Spec: map[string]any{},
		}
		_, err := gEx.AsExecutor()
		require.NoError(t, err)
	}
}

