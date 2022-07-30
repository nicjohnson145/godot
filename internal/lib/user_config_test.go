package lib

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"testing"
)

var _ VaultClient = (*MockVaultClient)(nil)

type MockVaultClient struct {
	ReadKeyFunc     func(string, string) (string, error)
	InitializedFunc func() bool
}

func (m *MockVaultClient) Initialized() bool {
	if m.InitializedFunc != nil {
		return m.InitializedFunc()
	}
	return true
}

func (m *MockVaultClient) ReadKey(s string, s1 string) (string, error) {
	if m.ReadKeyFunc != nil {
		return m.ReadKeyFunc(s, s1)
	}
	return "", fmt.Errorf("No find")
}

func (m *MockVaultClient) ReadKeyOrDie(s string, s1 string) string {
	s, err := m.ReadKey(s, s1)
	if err != nil {
		panic(err)
	}
	return s
}

func TestAuthenticationSetup(t *testing.T) {
	dir := t.TempDir()

	conf := UserConfig{
		GithubUser: "testuser",
		VaultConfig: VaultConfig{
			GithubPatFromVault: true,
			GithubPatConfig: VaultPatConfig{
				Path: "path/to/key",
				Key:  "key1",
			},
		},
	}
	b, err := yaml.Marshal(conf)
	require.NoError(t, err)
	confPath := path.Join(dir, "some-conf.yaml")
	err = os.WriteFile(confPath, b, 0777)
	require.NoError(t, err)

	c := NewConfigFromPath(
		confPath,
		func(uc *UserConfig) {
			uc.VaultConfig.Client = &MockVaultClient{
				ReadKeyFunc: func(s1, s2 string) (string, error) {
					if s1 == "path/to/key" && s2 == "key1" {
						return "MY_GITHUB_PAT", nil
					}
					return "", fmt.Errorf("%v and %v not known", s1, s2)
				},
			}
		},
		ConfigOverrides{},
	)

	require.Equal(t, "MY_GITHUB_PAT", c.GithubPAT)
	require.NotEqual(t, "", c.GithubAuth)
}
