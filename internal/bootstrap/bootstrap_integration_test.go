// +build integration

package bootstrap

import (
	"testing"
)

func TestIntegration(t *testing.T) {
	t.Run("integration_thing", func(t *testing.T) {
		x := 2
		if x != 1 {
			t.Fatalf("bad")
		}
	})
}
