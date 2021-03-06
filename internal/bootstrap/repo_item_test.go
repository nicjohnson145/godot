// +build !brew_integration
// +build !apt_integration

package bootstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
	"github.com/stretchr/testify/require"
)

func repoSetup(t *testing.T, isGit bool) (repoItem, func()) {
	t.Helper()
	dir, remove := help.CreateTempDir(t, "git-checkout")
	if isGit {
		err := os.Mkdir(filepath.Join(dir, ".git"), 0600)
		require.NoError(t, err)
	}

	return NewRepoItem("https://somerepo.github.com", dir), remove
}

func TestCheck(t *testing.T) {

	t.Run("Check_directory_exists_is_git", func(t *testing.T) {
		r, remove := repoSetup(t, true)
		defer remove()

		exists, err := r.Check()
		require.NoError(t, err)

		require.Equal(t, exists, true, "Repo should show as existing")
	})

	t.Run("Check_directory_exists_not_git", func(t *testing.T) {
		r, remove := repoSetup(t, false)
		defer remove()

		_, err := r.Check()
		require.Error(t, err)
	})

	t.Run("Check_directory_missing", func(t *testing.T) {
		r := NewRepoItem("https://somerepo.github.com", "/some/path")
		exists, err := r.Check()
		require.NoError(t, err)

		require.Equal(t, exists, false, "Repo should show as not existing")
	})
}
