package lib

import (
	"bytes"
	"io"
	"path"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/flytam/filenamify"
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

func normalizeTag(tag string) string {
	out, err := filenamify.Filenamify(tag, filenamify.Options{
		Replacement: "-",
	})
	if err != nil {
		log.Fatalf("Error converting tag to filesystem-safe name: %v", err)
	}
	return out
}

func getDestination(conf UserConfig, name string, tag string) string {
	return path.Join(conf.BinaryDir, name+"-"+normalizeTag(tag))
}

func getSymlinkName(conf UserConfig, name string, tag string) string {
	return strings.TrimSuffix(getDestination(conf, name, tag), "-"+normalizeTag(tag))
}

func copyToDestination(src string, dest string) {
	sfile, err := os.Open(src)
	if err != nil {
		log.Fatal("Error opening binary file: ", err)
	}
	defer sfile.Close()

	ensureContainingDir(dest)

	dfile, err := os.Create(dest)
	if err != nil {
		log.Fatalf("Error creating destination file: %v", err)
	}
	defer dfile.Close()

	_, err = io.Copy(dfile, sfile)
	if err != nil {
		log.Fatalf("Error copying binary to destination: %v", err)
	}

	if err := os.Chmod(dest, 0755); err != nil {
		log.Fatalf("Error chmoding destination file: %v", err)
	}
}

func createSymlink(src string, dest string) {
	exists, err := pathExists(dest)
	if err != nil {
		log.Fatalf("Error checking path existance: %v", err)
	}
	if exists {
		err := os.Remove(dest)
		if err != nil {
			log.Fatalf("Error removing existing file: %v", err)
		}
	}
	err = os.Symlink(src, dest)
	if err != nil {
		log.Fatalf("Error symlinking binary to tagged version: %v", err)
	}
}

