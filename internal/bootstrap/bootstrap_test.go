// +build !integration

package bootstrap

import (
	"testing"
)

func TestRegular(t *testing.T) {
	t.Run("thing", func(t *testing.T) {
		x := 1
		if x != 1 {
			t.Fatalf("bad")
		}
	})
}
