package lib

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
)

const (
	Binary = "binary"
	TarGzDir = "targz_dir"
	TarGzBin = "targz_bin"
)

type releaseResponse struct {
	Assets []release `json:"assets"`
}

type release struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

var _ Executor = (*GithubRelease)(nil)

type GithubRelease struct {
	Name           string `yaml:"name"`
	Repo           string `yaml:"repo"`
	Type           string `yaml:"type"`
	Tag            string `yaml:"tag"`
	Path           string `yaml:"path"`
	MacPattern     string `yaml:"mac-pattern"`
	LinuxPattern   string `yaml:"linux-pattern"`
	WindowsPattern string `yaml:"windows-pattern"`
}

func (g GithubRelease) GetName() string {
	return g.Name
}

func (g GithubRelease) Execute(conf UserConfig) {
	// Check if the destination path is already there, if so don't redownload
	exists, err := pathExists(g.getDestination(conf))
	if err != nil {
		log.Fatalf("Unable to check existance of %v: %v", g.getDestination(conf), err)
	}
	if exists {
		log.Infof("%v already downloaded, skipping", g.Name)
		return
	}
	log.Info("Downloading ", g.Repo)

	dir, err := ioutil.TempDir("", "godot-")
	if err != nil {
		log.Fatal("Unable to make temp directory")
	}
	defer os.RemoveAll(dir)

	filepath := path.Join(dir, "release")
	release := g.getRelease(conf)

	req := requests.
		URL(release.DownloadUrl).
		ToFile(filepath)
	if conf.GithubAuth != "" {
		req = req.Header("Authorization", conf.GithubAuth)
	}
	err = req.Fetch(context.TODO())
	if err != nil {
		log.Fatalf("Error downloading release asset: %v", err)
	}

	switch g.Type {
	case TarGzDir:
		g.handleTarGzDir(conf, dir, filepath, release)
	case TarGzBin:
		g.handleTarGzBin(conf, dir, filepath)
	case Binary:
		g.handleBinary(conf, filepath)
	default:
		log.Fatalf("Unknown type of %v", g.Type)
	}
}

func (g GithubRelease) getRelease(conf UserConfig) release {
	var resp releaseResponse
	req := requests.
		URL(fmt.Sprintf("https://api.github.com/repos/%v/releases/tags/%v", g.Repo, g.Tag)).
		ToJSON(&resp)
	if conf.GithubAuth != "" {
		req = req.Header("Authorization", conf.GithubAuth)
	}
	err := req.Fetch(context.TODO())
	if err != nil {
		log.Fatalf("Error getting release %v for %v: %v", g.Tag, g.Repo, err)
	}

	pattern := g.getDownloadPattern()

	for _, r := range resp.Assets {
		if pattern.MatchString(r.Name) {
			return r
		}
	}

	log.Fatalf("No assets in %v:%v match the pattern %v", g.Tag, g.Repo, pattern)
	return release{}
}

func (g GithubRelease) getDownloadPattern() *regexp.Regexp {
	var pat string
	switch runtime.GOOS {
	case "windows":
		pat = g.WindowsPattern
	case "linux":
		pat = g.LinuxPattern
	case "darwin":
		pat = g.MacPattern
	}

	if pat == "" {
		log.Fatal(fmt.Sprintf("GithubRelease %v does not have configured download pattern for %v", g.Repo, runtime.GOOS))
	}

	exp, err := regexp.Compile(pat)
	if err != nil {
		log.Fatal(fmt.Sprintf("GithubRelease %v pattern is not a valid regular expression: %v", runtime.GOOS, err))
	}

	return exp
}

func (g GithubRelease) handleTarGz(conf UserConfig, tempdir string, downloadpath string, dir string) {
	file, err := os.Open(downloadpath)
	if err != nil {
		log.Fatalf("Error opening downloaded release: %v", err)
	}
	defer file.Close()

	outpath := path.Join(tempdir, "release-unpacked")
	if err := os.Mkdir(outpath, os.ModePerm); err != nil {
		log.Fatalf("Error creating temp directory: %v", err)
	}

	if err := Untar(file, outpath); err != nil {
		log.Fatalf("Error unpacking tarball: %v", err)
	}

	g.copyToDestination(path.Join(outpath, dir, g.Path), g.getDestination(conf))
}

func (g GithubRelease) handleTarGzDir(conf UserConfig, tempdir string, downloadpath string, release release) {
	g.handleTarGz(conf, tempdir, downloadpath, release.Name[0 : len(release.Name)-7])
}

func (g GithubRelease) handleTarGzBin(conf UserConfig, tempdir string, downloadpath string) {
	g.handleTarGz(conf, tempdir, downloadpath, "")
}

func (g GithubRelease) handleBinary(conf UserConfig, downloadpath string) {
	g.copyToDestination(downloadpath, g.getDestination(conf))
}

func (g GithubRelease) copyToDestination(src string, dest string) {
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

func (g GithubRelease) getDestination(conf UserConfig) string {
	return path.Join(conf.BinaryDir, g.Name)
}
