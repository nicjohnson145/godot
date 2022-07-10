package lib

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/flytam/filenamify"
	"github.com/mholt/archiver"
	log "github.com/sirupsen/logrus"
)

var (
	regexMusl = regexp.MustCompile("(?i)musl")
	regexLinuxPkg = regexp.MustCompile(`(?i)(\.deb|\.rpm)$`)

	osRegexMap = map[string]*regexp.Regexp{
		"windows": regexp.MustCompile(`(?i)(windows|win)`),
		"linux":   regexp.MustCompile("(?i)linux"),
		"darwin":  regexp.MustCompile(`(?i)(darwin|mac(os)?|apple|osx)`),
	}

	archRegexMap = map[string]*regexp.Regexp{
		"386":   regexp.MustCompile(`(?i)(i?386|x86_32|amd32|x32)`),
		"amd64": regexp.MustCompile(`(?i)(x86_64|amd64|x64)`),
		"arm64": regexp.MustCompile(`(?i)(arm64|aarch64)`),
	}
)

type releaseResponse struct {
	Assets []release `json:"assets"`
}

type release struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

type githubTag struct {
	Name string `json:"name"`
}

var _ Executor = (*GithubRelease)(nil)

type GithubRelease struct {
	Name           string `yaml:"name"`
	Repo           string `yaml:"repo"`
	Tag            string `yaml:"tag"`
	IsArchive      bool   `yaml:"is-archive"`
	Regex          string `yaml:"regex"`
	MacPattern     string `yaml:"mac-pattern"`
	LinuxPattern   string `yaml:"linux-pattern"`
	WindowsPattern string `yaml:"windows-pattern"`
}

func (g *GithubRelease) GetName() string {
	return g.Name
}

func (g *GithubRelease) Execute(conf UserConfig, opts SyncOpts) {
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

	release := g.getRelease(conf)
	filepath := path.Join(dir, release.Name)

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

	extractDir := path.Join(dir, "extract")
	binaryPath := g.extractBinary(filepath, extractDir)
	g.copyToDestination(binaryPath, g.getDestination(conf))
	g.createSymlink(g.getDestination(conf), g.getSymlinkName(conf))
}

func (g *GithubRelease) getRelease(conf UserConfig) release {
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

	return g.getAsset(resp, runtime.GOOS, runtime.GOARCH)

	//for _, r := range resp.Assets {
	//    if pattern.MatchString(r.Name) {
	//        return r
	//    }
	//}
}

func (g *GithubRelease) GetLatestTag(conf UserConfig) string {
	var resp []githubTag
	req := requests.
		URL(fmt.Sprintf("https://api.github.com/repos/%v/tags", g.Repo)).
		ToJSON(&resp)
	if conf.GithubAuth != "" {
		req = req.Header("Authorization", conf.GithubAuth)
	}
	err := req.Fetch(context.TODO())
	if err != nil {
		log.Fatalf("Error getting tag list for %v", g.Repo)
	}

	// Sort of assuming that the API returns things in cronological order? A better approach would
	// be to get all tags fully, and then do a semver compare, :shrug:
	return resp[0].Name
}

func (g *GithubRelease) getAsset(resp releaseResponse, userOs string, userArch string) release {
	var pat string
	switch userOs {
	case "windows":
		pat = g.WindowsPattern
	case "linux":
		pat = g.LinuxPattern
	case "darwin":
		pat = g.MacPattern
	}

	if pat != "" {
		log.Debugf("Using user specified pattern of %v", pat)
		userRegex, err := regexp.Compile(pat)
		if err != nil {
			log.Fatalf("Error compiling user specified regex: %v", err)
		}
		assets := g.filterAssets(resp.Assets, userRegex, true)
		if len(assets) != 1 {
			log.Fatalf("Expected 1 matching asset for pattern %v, got %v", pat, len(assets))
		}

		return assets[0]
	}

	checkNoMatches := func(assets []release, matchType string) {
		if len(assets) == 0 {
			log.Fatalf("Unable to auto detect release name, no assets match pre-defined patterns for %v", matchType)
		}
	}

	// Otherwise, lets try to detect it
	osPat, ok := osRegexMap[userOs]
	if !ok {
		log.Fatalf("Unsupported OS of %v", userOs)
	}
	assets := g.filterAssets(resp.Assets, osPat, true)
	checkNoMatches(assets, "OS")
	if len(assets) == 1 {
		// If there's only one matching asset by OS, then we're done here
		log.Debugf("Reached a single asset after OS matching, found %v", assets[0].Name)
		return assets[0]
	}

	// If we're got more than 1, then lets try to narrow it down by architecture
	archPat, ok := archRegexMap[userArch]
	if !ok {
		log.Fatalf("Unsupported architecture of %v", userArch)
	}
	assets = g.filterAssets(assets, archPat, true)
	checkNoMatches(assets, "architecture")
	if len(assets) == 1 {
		// If there's only one, then we're done here
		log.Debugf("Reached a single asset after architecture matching, found %v", assets[0].Name)
		return assets[0]
	}

	// If we're not linux, then I don't have any more tricks up my sleeve
	if userOs != "linux" {
		log.Fatalf("Unable to auto-detect release asset, please specify a per-OS pattern")
	}

	// But if we are, lets filter off any non-MUSL or deb/rpm 
	assets = g.filterAssets(assets, regexLinuxPkg, false)
	assets = g.filterAssets(assets, regexMusl, true)
	checkNoMatches(assets, "linux packages/musl")

	if len(assets) == 1 {
		// If there's only one, then we're done here
		log.Debugf("Reached a single asset after linux package and musl filtering, found %v", assets[0].Name)
		return assets[0]
	}

	log.Fatalf("Unable to auto-detect release asset, please specify a per-OS pattern")
	return release{}
}

func (g *GithubRelease) filterAssets(assets []release, pat *regexp.Regexp, match bool) []release {
	matches := []release{}
	for _, r := range assets {
		if pat.MatchString(r.Name) == match {
			matches = append(matches, r)
		}
	}

	return matches
}

func (g *GithubRelease) getDownloadPattern() *regexp.Regexp {
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

func (g *GithubRelease) copyToDestination(src string, dest string) {
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

func (g *GithubRelease) createSymlink(src string, dest string) {
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

func (g *GithubRelease) getDestination(conf UserConfig) string {
	return path.Join(conf.BinaryDir, g.Name+"-"+g.normalizeTag())
}

func (g *GithubRelease) normalizeTag() string {
	out, err := filenamify.Filenamify(g.Tag, filenamify.Options{
		Replacement: "-",
	})
	if err != nil {
		log.Fatalf("Error converting tag to filesystem-safe name: %v", err)
	}
	return out
}

func (g *GithubRelease) getSymlinkName(conf UserConfig) string {
	return strings.TrimSuffix(g.getDestination(conf), "-"+g.normalizeTag())
}

func (g *GithubRelease) isExecutableFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Error determining if file is executable: %v", err)
	}

	filePerm := fileInfo.Mode()
	return !fileInfo.IsDir() && filePerm&0111 != 0
}

func (g *GithubRelease) extractBinary(downloadPath string, extractPath string) string {
	if g.IsArchive {
		err := archiver.Unarchive(downloadPath, extractPath)
		if err != nil {
			log.Fatalf("Error extracting archive: %v", err)
		}
		return g.findExecutable(extractPath)
	}
	return downloadPath
}

func (g *GithubRelease) findExecutable(path string) string {
	executables := []string{}

	var validFile func(path string) bool
	if g.Regex != "" {
		regex, err := regexp.Compile(g.Regex)
		if err != nil {
			log.Fatalf("Unable to compile executable regex: %v", err)
		}
		validFile = func(path string) bool {
			return regex.MatchString(path)
		}
	} else {
		validFile = g.isExecutableFile
	}
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if validFile(path) {
			executables = append(executables, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking extracted directory tree: %v", err)
	}

	if len(executables) != 1 {
		log.Fatalf("Expected to find 1 executable, instead found %v, please specify a regex to the binary", len(executables))
	}

	return executables[0]
}
