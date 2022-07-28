package lib

import (
	"bytes"
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
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

func copyToDestination(src string, dest string) error {
	sfile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening binary file: %w", err)
	}
	defer sfile.Close()

	ensureContainingDir(dest)

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

	dir, err := ioutil.TempDir("", "godot-")
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
