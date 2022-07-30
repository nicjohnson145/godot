package lib

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"runtime"

	"github.com/carlmjohnson/requests"
	log "github.com/sirupsen/logrus"
)

var (
	regexMusl     = regexp.MustCompile("(?i)musl")
	regexLinuxPkg = regexp.MustCompile(`(?i)(\.deb|\.rpm|\.apk)$`)

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

const (
	TypeGithubRelease = "github-release"
	Latest            = "LATEST"
)

type releaseResponse struct {
	Assets []release `json:"assets"`
}

type release struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

type githubTag struct {
	TagName string `json:"tag_name"`
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

func (g *GithubRelease) Type() string {
	return TypeGithubRelease
}

func (g *GithubRelease) Execute(conf UserConfig, opts SyncOpts) {
	log.Infof("Ensuring %v", g.Repo)
	release := g.getRelease(conf)

	err := downloadAndSymlinkBinary(downloadOpts{
		Name:         g.Name,
		DownloadName: path.Base(release.DownloadUrl),
		FinalDest:    getDestination(conf, g.Name, g.Tag),
		Url:          release.DownloadUrl,
		RequestFunc: func(req *requests.Builder) {
			if conf.GithubAuth != "" {
				req.Header("Authorization", conf.GithubAuth)
			}
		},
		SearchFunc:  g.regexFunc(),
		SymlinkName: getSymlinkName(conf, g.Name, g.Tag),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (g *GithubRelease) regexFunc() searchFunc {
	if g.Regex == "" {
		return nil
	}

	regex, err := regexp.Compile(g.Regex)
	if err != nil {
		log.Fatalf("Unable to compile executable regex: %v", err)
	}

	return func(path string) (bool, error) {
		return regex.MatchString(path), nil
	}
}

func (g *GithubRelease) getRelease(conf UserConfig) release {
	if g.Tag == Latest {
		g.Tag = g.GetLatestRelease(conf)
	}
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
}

func (g *GithubRelease) GetLatestRelease(conf UserConfig) string {
	var resp githubTag
	req := requests.
		URL(fmt.Sprintf("https://api.github.com/repos/%v/releases/latest", g.Repo)).
		ToJSON(&resp)
	if conf.GithubAuth != "" {
		req = req.Header("Authorization", conf.GithubAuth)
	}
	err := req.Fetch(context.TODO())
	if err != nil {
		log.Fatalf("Error getting tag list for %v: %v", g.Repo, err)
	}

	// Sort of assuming that the API returns things in cronological order? A better approach would
	// be to get all tags fully, and then do a semver compare, :shrug:
	return resp.TagName
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
		g.setArchive(assets[0])
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
		g.setArchive(assets[0])
		return assets[0]
	}

	// If we're not linux, then I don't have any more tricks up my sleeve
	if userOs != "linux" {
		log.Fatalf("Unable to auto-detect release asset, please specify a per-OS pattern")
	}

	// But if we are, lets filter off any non-MUSL or deb/rpm
	assets = g.filterAssets(assets, regexLinuxPkg, false)
	checkNoMatches(assets, "non linux packages")
	if len(assets) == 1 {
		log.Debugf("Reached a single asset after package filtering, found %v", assets[0].Name)
		g.setArchive(assets[0])
		return assets[0]
	}

	// If we've still got multiples, prefere statically linked binaries
	assets = g.filterAssets(assets, regexMusl, true)
	checkNoMatches(assets, "musl static linking")
	if len(assets) == 1 {
		// If there's only one, then we're done here
		log.Debugf("Reached a single asset after musl filtering, found %v", assets[0].Name)
		g.setArchive(assets[0])
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

func (g *GithubRelease) setArchive(asset release) {
	g.IsArchive = isArchiveFile(path.Base(asset.DownloadUrl))
}
