package lib

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestExtractBinary(t *testing.T) {
	t.Run("tar_gz", func(t *testing.T) {
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		extractDir := t.TempDir()

		got, err := extractBinary("testdata/fzf.tar.gz", extractDir, "", nil)
		require.NoError(t, err)
		want := filepath.Join(extractDir, "fzf")
		require.Equal(t, want, got)
	})

	t.Run("zip", func(t *testing.T) {
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		extractDir := t.TempDir()

		got, err := extractBinary("testdata/fzf.zip", extractDir, "", nil)
		require.NoError(t, err)
		want := filepath.Join(extractDir, "fzf")
		require.Equal(t, want, got)
	})
}
