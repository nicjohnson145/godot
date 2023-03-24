package lib

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

var (
	regexMusl = regexp.MustCompile("(?i)musl")
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

	ErrUnableToAutodetectError = errors.New("unable to auto detect release asset")
)

const (
	Latest = "LATEST"
)

type asset struct {
	Name        string
	DownloadUrl string
}

type assetFilter struct {
	Name string
	Func func(asset) (bool, error)
}

func autoDetectAsset(log zerolog.Logger, assets []asset, userOS string, userArch string) (asset, error) {
	osRegex, ok := osRegexMap[userOS]
	if !ok {
		return asset{}, fmt.Errorf("unable to load regex for os filtering for os %v", userOS)
	}

	archRegex, ok := archRegexMap[userArch]
	if !ok {
		return asset{}, fmt.Errorf("unable to load regex for arch filtering for arch %v", userArch)
	}

	filters := []assetFilter{
		{
			Name: "remove checksums",
			Func: func(a asset) (bool, error) {
				return !strings.HasSuffix(a.Name, ".sha256") && !strings.HasSuffix(a.Name, ".checksum"), nil
			},
		},
		{
			Name: "remove linux packages",
			Func: func(a asset) (bool, error) {
				return !regexLinuxPkg.Match([]byte(a.Name)), nil
			},
		},
		{
			Name: "os-filter",
			Func: func(a asset) (bool, error) {
				return osRegex.Match([]byte(a.Name)), nil
			},
		},
		{
			Name: "architecture-filter",
			Func: func(a asset) (bool, error) {
				return archRegex.Match([]byte(a.Name)), nil
			},
		},
	}

	remainingAssets := assets

	log.Debug().Msg("attempting to auto detect release asset")
	for _, filter := range filters {
		log.Debug().Str("name", filter.Name).Msg("applying filter")

		result := []asset{}
		for _, item := range remainingAssets {
			ok, err := filter.Func(item)
			if err != nil {
				return asset{}, fmt.Errorf("error applying filter %v on asset %v", filter.Name, item.Name)
			}
			if ok {
				result = append(result, item)
			}
		}

		numAssets := len(result)
		log.Debug().Msgf("%v remaining assets after filter", numAssets)
		if numAssets == 0 {
			return asset{}, ErrUnableToAutodetectError
		}

		if numAssets == 1 {
			return result[0], nil
		}

		remainingAssets = result
	}

	// If we've exhausted all filters, then we can't auto detect it
	return asset{}, ErrUnableToAutodetectError
}
