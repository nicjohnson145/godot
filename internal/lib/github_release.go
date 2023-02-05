package lib

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"runtime"

	"github.com/carlmjohnson/requests"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
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
	Latest = "LATEST"
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
	Name           string         `yaml:"-"`
	Repo           string         `yaml:"repo" mapstructure:"repo"`
	Tag            string         `yaml:"tag" mapstructure:"tag"`
	IsArchive      bool           `yaml:"is-archive" mapstructure:"is-archive"`
	Regex          string         `yaml:"regex" mapstructure:"regex"`
	MacPattern     string         `yaml:"mac-pattern" mapstructure:"mac-pattern"`
	LinuxPattern   string         `yaml:"linux-pattern" mapstructure:"linux-pattern"`
	WindowsPattern string         `yaml:"windows-pattern" mapstructure:"windows-pattern"`
	log            zerolog.Logger `yaml:"-"`
}

func (g *GithubRelease) SetLogger(log zerolog.Logger) {
	g.log = log
}

func (g *GithubRelease) GetName() string {
	return g.Name
}

func (g *GithubRelease) SetName(n string) {
	g.Name = n
}

func (g *GithubRelease) Type() ExecutorType {
	return ExecutorTypeGithubRelease
}

func (g *GithubRelease) Validate() error {
	var errs *multierror.Error

	if g.Repo == "" {
		errs = multierror.Append(errs, fmt.Errorf("repo is required"))
	}
	if g.Tag == "" {
		errs = multierror.Append(errs, fmt.Errorf("tag is required"))
	}
	if g.Regex != "" {
		_, err := regexp.Compile(g.Regex)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("unable to compile regex: %w", err))
		}
	}

	return errs.ErrorOrNil()
}

func (g *GithubRelease) Execute(conf UserConfig, opts SyncOpts, _ GodotConfig) error {
	g.log.Info().Str("release", g.Repo).Msg("ensuring release")
	release, err := g.getRelease(conf)
	if err != nil {
		return fmt.Errorf("error determining release: %w", err)
	}

	searchFunc, err := g.regexFunc()
	if err != nil {
		return err
	}

	dest, err := getDestination(conf, g.Name, g.Tag)
	if err != nil {
		return err
	}

	symlink, err := getSymlinkName(conf, g.Name, g.Tag)
	if err != nil {
		return err
	}

	err = downloadAndSymlinkBinary(downloadOpts{
		Name:         g.Name,
		DownloadName: path.Base(release.DownloadUrl),
		FinalDest:    dest,
		Url:          release.DownloadUrl,
		RequestFunc: func(req *requests.Builder) {
			if conf.GithubAuth != "" {
				req.Header("Authorization", conf.GithubAuth)
			}
		},
		SearchFunc:  searchFunc,
		SymlinkName: symlink,
	}, g.log)
	if err != nil {
		return fmt.Errorf("error during download/symlink: %w", err)
	}
	return nil
}

func (g *GithubRelease) regexFunc() (searchFunc, error) {
	if g.Regex == "" {
		return nil, nil
	}

	regex, err := regexp.Compile(g.Regex)
	if err != nil {
		return nil, fmt.Errorf("unable to compile executable regex: %v", err)
	}

	return func(path string) (bool, error) {
		return regex.MatchString(path), nil
	}, nil
}

func (g *GithubRelease) getRelease(conf UserConfig) (release, error) {
	if g.Tag == Latest {
		tag, err := g.GetLatestRelease(conf)
		if err != nil {
			return release{}, fmt.Errorf("error determining latest release: %w", err)
		}
		g.Tag = tag
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
		return release{}, fmt.Errorf("error getting release %v for %v: %v", g.Tag, g.Repo, err)
	}

	asset, err := g.getAsset(resp, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return release{}, fmt.Errorf("error determing asset: %w", err)
	}
	return asset, nil
}

func (g *GithubRelease) GetLatestRelease(conf UserConfig) (string, error) {
	var resp githubTag
	req := requests.
		URL(fmt.Sprintf("https://api.github.com/repos/%v/releases/latest", g.Repo)).
		ToJSON(&resp)
	if conf.GithubAuth != "" {
		req = req.Header("Authorization", conf.GithubAuth)
	}
	err := req.Fetch(context.TODO())
	if err != nil {
		return "", fmt.Errorf("error getting tag list for %v: %v", g.Repo, err)
	}

	// Sort of assuming that the API returns things in cronological order? A better approach would
	// be to get all tags fully, and then do a semver compare, :shrug:
	return resp.TagName, nil
}

// TODO: reduce the complexity of this function
//
//nolint:gocyclo
func (g *GithubRelease) getAsset(resp releaseResponse, userOs string, userArch string) (release, error) {
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
		g.log.Debug().Str("pattern", pat).Msg("using user specified pattern")
		userRegex, err := regexp.Compile(pat)
		if err != nil {
			return release{}, fmt.Errorf("error compiling user specified regex: %v", err)
		}
		assets := g.filterAssets(resp.Assets, userRegex, true)
		if len(assets) != 1 {
			return release{}, fmt.Errorf("expected 1 matching asset for pattern %v, got %v", pat, len(assets))
		}

		return assets[0], nil
	}

	noMatchErr := func(matchType string) (release, error) {
		return release{}, fmt.Errorf("enable to auto detect release name, no assets match pre-defined patterns for %v", matchType)
	}

	// Otherwise, lets try to detect it
	osPat, ok := osRegexMap[userOs]
	if !ok {
		return release{}, fmt.Errorf("unsupported OS of %v", userOs)
	}
	assets := g.filterAssets(resp.Assets, osPat, true)
	if len(assets) == 0 {
		return noMatchErr("OS")
	}
	if len(assets) == 1 {
		// If there's only one matching asset by OS, then we're done here
		g.log.Debug().Str("asset", assets[0].Name).Msg("reached a single asset after OS matching")
		g.setArchive(assets[0])
		return assets[0], nil
	}

	// If we're got more than 1, then lets try to narrow it down by architecture
	archPat, ok := archRegexMap[userArch]
	if !ok {
		return release{}, fmt.Errorf("unsupported architecture of %v", userArch)
	}
	assets = g.filterAssets(assets, archPat, true)
	if len(assets) == 0 {
		return noMatchErr("architecture")
	}
	if len(assets) == 1 {
		// If there's only one, then we're done here
		g.log.Debug().Str("asset", assets[0].Name).Msg("reached a single asset after architecture matching")
		g.setArchive(assets[0])
		return assets[0], nil
	}

	// If we're not linux, then I don't have any more tricks up my sleeve
	if userOs != "linux" {
		return release{}, fmt.Errorf("unable to auto-detect release asset, please specify a per-OS pattern")
	}

	// But if we are, lets filter off any non-MUSL or deb/rpm
	assets = g.filterAssets(assets, regexLinuxPkg, false)
	if len(assets) == 0 {
		return noMatchErr("non linux packages")
	}
	if len(assets) == 1 {
		g.log.Debug().Str("asset", assets[0].Name).Msg("reached a single asset after package filtering")
		g.setArchive(assets[0])
		return assets[0], nil
	}

	// If we've still got multiples, prefer statically linked binaries
	assets = g.filterAssets(assets, regexMusl, true)
	if len(assets) == 0 {
		return noMatchErr("musl static linking")
	}
	if len(assets) == 1 {
		// If there's only one, then we're done here
		g.log.Debug().Str("asset", assets[0].Name).Msg("reached a single asset after musl filtering")
		g.setArchive(assets[0])
		return assets[0], nil
	}

	return release{}, fmt.Errorf("enable to auto-detect release asset, please specify a per-OS pattern")
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
