package lib

import (
	"strings"
	"path/filepath"
	"os"
	"os/exec"
	log "github.com/sirupsen/logrus"
	"bytes"
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

func runCmd(bin string, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
