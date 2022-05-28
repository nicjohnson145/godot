package lib

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func NoFatals() (func(), func(*testing.T)) {
	restore := func() {
		log.StandardLogger().ExitFunc = nil
	}

	fatal := false
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	noFatal := func(t *testing.T) {
		require.False(t, fatal, "There should be no calls to log.Fatal")
	}

	return restore, noFatal
}

func HasFatals() (func(), func(*testing.T)) {
	restore := func() {
		log.StandardLogger().ExitFunc = nil
	}

	fatal := false
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	noFatal := func(t *testing.T) {
		require.True(t, fatal, "There should be calls to log.Fatal")
	}

	return restore, noFatal
}
