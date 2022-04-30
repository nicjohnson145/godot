package binary

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/lib"
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
	TarGz  = "targz"
)

type releaseResponse struct {
	Assets []release `json:"assets"`
}

type release struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

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

func (g GithubRelease) Execute(conf config.UserConfig) {
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
		log.Fatal("Error downloading release asset: ", err)
	}

	switch g.Type {
	case TarGz:
		g.handleTarGz(conf, dir, filepath, release)
	case Binary:
		g.handleBinary(conf, filepath)
	}
}

func (g GithubRelease) getRelease(conf config.UserConfig) release {
	var resp releaseResponse
	req := requests.
		URL(fmt.Sprintf("https://api.github.com/repos/%v/releases/tags/%v", g.Repo, g.Tag)).
		ToJSON(&resp)
	if conf.GithubAuth != "" {
		req = req.Header("Authorization", conf.GithubAuth)
	}
	err := req.Fetch(context.TODO())
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting release %v for %v: %v", g.Tag, g.Repo, err))
	}

	pattern := g.getDownloadPattern()

	for _, r := range resp.Assets {
		if pattern.MatchString(r.Name) {
			return r
		}
	}

	log.Fatal(fmt.Sprintf("No assets in %v:%v match the pattern %v", g.Tag, g.Repo, pattern))
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

func (g GithubRelease) handleTarGz(conf config.UserConfig, tempdir string, downloadpath string, release release) {
	file, err := os.Open(downloadpath)
	if err != nil {
		log.Fatal("Error opening downloaded release: ", err)
	}
	defer file.Close()

	outpath := path.Join(tempdir, "release-unpacked")
	if err := os.Mkdir(outpath, os.ModePerm); err != nil {
		log.Fatal("Error creating temp directory: ", err)
	}

	if err := lib.Untar(file, outpath); err != nil {
		log.Fatal("Error unpacking tarball: ", err)
	}

	// Strip the .targz off the release name, since that's what it will un-tar to
	minusExt := release.Name[0 : len(release.Name)-7]

	g.copyToDestination(path.Join(outpath, minusExt, g.Path), path.Join(conf.BinaryDir, g.Name))
}

func (g GithubRelease) handleBinary(conf config.UserConfig, downloadpath string) {
	g.copyToDestination(downloadpath, path.Join(conf.BinaryDir, g.Name))
}

func (g GithubRelease) copyToDestination(src string, dest string) {
	sfile, err := os.Open(src)
	if err != nil {
		log.Fatal("Error opening binary file: ", err)
	}
	defer sfile.Close()

	dfile, err := os.Create(dest)
	if err != nil {
		log.Fatal("Error creating destination file: ", err)
	}
	defer dfile.Close()

	_, err = io.Copy(dfile, sfile)
	if err != nil {
		log.Fatal("Error copying binary to destination: ", err)
	}

	if err := os.Chmod(dest, 0755); err != nil {
		log.Fatal("Error chmoding destination file: ", err)
	}
}
