package lib

import (
	"strings"
	"path/filepath"
	"os"
	log "github.com/sirupsen/logrus"
)

func replaceTilde(s string, replacement string) string {
	if !strings.Contains(s, "~") {
		return s
	}
	return strings.ReplaceAll(s, "~", replacement)
}

func ensureContainingDir(destpath string) {
	dir := filepath.Dir(destpath)
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		log.Fatalf("Error creating containing directories: %v", err)
	}
}
