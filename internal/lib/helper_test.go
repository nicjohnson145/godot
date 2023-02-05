package lib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildDirectoryStructure(t *testing.T, structure map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for path, content := range structure {
		// If it ends with a slash, make an empty directory
		if strings.HasSuffix(path, "/") {
			require.NoError(t, os.MkdirAll(filepath.Join(root, path), 0755))
			continue
		}

		// Otherwise write a file with that name and content
		containingDir := filepath.Dir(path)
		require.NoError(t, os.MkdirAll(filepath.Join(root, containingDir), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(root, path), []byte(content), 0644))
	}

	return root
}

func requireContents(t *testing.T, path string, expected string) {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(b))
}

func cleanFuncsMap(t *testing.T) {
	t.Helper()

	delete(funcs, funcNameIsInstalled)
	delete(funcs, funcNameVaultLookup)
}

