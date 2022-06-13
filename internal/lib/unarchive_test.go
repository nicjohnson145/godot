package lib

import (
	"testing"
	"github.com/stretchr/testify/require"
	"path/filepath"
)

func TestExtractBinary(t *testing.T) {
	t.Run("tar_gz", func(t *testing.T) {
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		extractDir := t.TempDir()

		got := extractBinary("testdata/fzf.tar.gz", extractDir, "")
		want := filepath.Join(extractDir, "fzf")
		require.Equal(t, want, got)
	})

	t.Run("zip", func(t *testing.T) {
		restore, noFatal := NoFatals(t)
		defer noFatal(t)
		defer restore()

		extractDir := t.TempDir()

		got := extractBinary("testdata/fzf.zip", extractDir, "")
		want := filepath.Join(extractDir, "fzf")
		require.Equal(t, want, got)
	})
}
