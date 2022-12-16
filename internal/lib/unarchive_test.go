package lib

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractBinary(t *testing.T) {
	t.Run("tar_gz", func(t *testing.T) {
		extractDir := t.TempDir()

		got, err := extractBinary("testdata/fzf.tar.gz", extractDir, "", nil)
		require.NoError(t, err)
		want := filepath.Join(extractDir, "fzf")
		require.Equal(t, want, got)
	})

	t.Run("zip", func(t *testing.T) {
		extractDir := t.TempDir()

		got, err := extractBinary("testdata/fzf.zip", extractDir, "", nil)
		require.NoError(t, err)
		want := filepath.Join(extractDir, "fzf")
		require.Equal(t, want, got)
	})
}
