package lib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"
)

func replaceTilde(s string, replacement string) string {
	if !strings.Contains(s, "~") {
		return s
	}
	return strings.ReplaceAll(s, "~", replacement)
}

func ensureContainingDir(destpath string) error {
	dir := filepath.Dir(destpath)
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return fmt.Errorf("error creating containing directories: %w", err)
	}
	return nil
}

func runCmd(bin string, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func normalizeTag(tag string) (string, error) {
	out, err := filenamify.Filenamify(tag, filenamify.Options{
		Replacement: "-",
	})
	if err != nil {
		return "", fmt.Errorf("error converting tag to filesystem-safe name: %v", err)
	}
	return out, nil
}

func getDestination(conf UserConfig, name string, tag string) (string, error) {
	normTag, err := normalizeTag(tag)
	if err != nil {
		return "", err
	}
	return path.Join(conf.BinaryDir, name+"-"+normTag), nil
}

func getSymlinkName(conf UserConfig, name string, tag string) (string, error) {
	normTag, err := normalizeTag(tag)
	if err != nil {
		return "", err
	}
	dest, err := getDestination(conf, name, tag)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(dest, "-"+normTag), nil
}

func copyToDestination(src string, dest string) error {
	sfile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening binary file: %w", err)
	}
	defer sfile.Close()

	if err := ensureContainingDir(dest); err != nil {
		return err
	}

	dfile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dfile.Close()

	_, err = io.Copy(dfile, sfile)
	if err != nil {
		return fmt.Errorf("error copying binary to destination: %w", err)
	}

	if err := os.Chmod(dest, 0755); err != nil {
		return fmt.Errorf("error chmoding destination file: %w", err)
	}
	return nil
}

func createSymlink(src string, dest string) error {
	exists, err := pathExists(dest)
	if err != nil {
		return fmt.Errorf("Error checking path existance: %w", err)
	}
	if exists {
		err := os.Remove(dest)
		if err != nil {
			return fmt.Errorf("Error removing existing file: %w", err)
		}
	}
	err = os.Symlink(src, dest)
	if err != nil {
		return fmt.Errorf("Error symlinking binary to tagged version: %w", err)
	}
	return nil
}

type downloadOpts struct {
	Name         string
	DownloadName string
	FinalDest    string
	Url          string
	RequestFunc  func(*requests.Builder)
	SearchFunc   searchFunc
	SymlinkName  string
}

func downloadAndSymlinkBinary(opts downloadOpts) error {
	exists, err := pathExists(opts.FinalDest)
	if err != nil {
		return fmt.Errorf("unable to check existance of %v: %w", opts.FinalDest, err)
	}
	if exists {
		log.Infof("%v already downloaded, skipping", opts.Name)
		return nil
	}

	log.Infof("Downloading %v", opts.Name)

	dir, err := os.MkdirTemp("", "godot-")
	if err != nil {
		return fmt.Errorf("unable to make temp directory")
	}
	defer os.RemoveAll(dir)

	filepath := path.Join(dir, opts.DownloadName)
	req := requests.
		URL(opts.Url).
		ToFile(filepath)

	if opts.RequestFunc != nil {
		opts.RequestFunc(req)
	}
	err = req.Fetch(context.TODO())
	if err != nil {
		return fmt.Errorf("error downloading from url: %w", err)
	}

	extractDir := path.Join(dir, "extract")
	binary, err := extractBinary(filepath, extractDir, "", opts.SearchFunc)
	if err != nil {
		return err
	}
	if err := copyToDestination(binary, opts.FinalDest); err != nil {
		return err
	}
	if err := createSymlink(opts.FinalDest, opts.SymlinkName); err != nil {
		return err
	}

	return nil
}
