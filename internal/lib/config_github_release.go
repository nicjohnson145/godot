package lib

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"text/tabwriter"

	"github.com/nicjohnson145/godot/internal/util"
)

func (c *Config) GetAllGithubReleaseNames() []string {
	names := []string{}
	for name := range c.content.GithubReleases {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Config) GetAllGithubReleaseNamesForTarget(target string) ([]string, error) {
	current, ok := c.content.Hosts[target]
	if !ok {
		return []string{}, fmt.Errorf("Target %v: %w", target, NotFoundError)
	}

	names := []string{}
	for _, ghr := range current.GithubRelease {
		names = append(names, ghr.Name)
	}
	return names, nil
}

func (c *Config) TargetUseGithubRelease(target string, name string, location string, trackUpdates bool) error {
	_, ok := c.content.GithubReleases[name]
	if !ok {
		return fmt.Errorf("Release %v: %w", name, NotFoundError)
	}

	current, ok := c.content.Hosts[target]
	if !ok {
		current = Host{}
	}

	current.GithubRelease = append(
		current.GithubRelease,
		GithubReleaseUsage{Name: name, Location: location, TrackUpdates: trackUpdates},
	)
	c.content.Hosts[target] = current
	return nil
}

func (c *Config) TargetCeaseGithubRelease(target string, name string) error {
	current, ok := c.content.Hosts[target]
	if !ok {
		return fmt.Errorf("Target %v: %w", target, NotFoundError)
	}

	newUsage := []GithubReleaseUsage{}
	for _, gru := range current.GithubRelease {
		if gru.Name != name {
			newUsage = append(newUsage, gru)
		}
	}

	current.GithubRelease = newUsage
	c.content.Hosts[target] = current
	return nil
}

func (c *Config) TargetShowGithubRelease(target string, w io.Writer) error {
	current, ok := c.content.Hosts[target]
	if !ok {
		return fmt.Errorf("Target %v: %w", target, NotFoundError)
	}

	tw := tabwriter.NewWriter(w, 0, 1, 0, ' ', tabwriter.AlignRight)

	sorted := []GithubReleaseUsage{}
	for _, gru := range current.GithubRelease {
		sorted = append(sorted, gru)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, gru := range sorted {
		if err := c.writeGithubReleaseUsage(gru, tw); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func (c *Config) ShowAllGithubRelease(w io.Writer) error {
	keys := []string{}
	for name := range c.content.GithubReleases {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(w, 0, 1, 0, ' ', tabwriter.AlignRight)
	for _, key := range keys {
		if err := c.writeGithubRelease(c.content.GithubReleases[key], tw); err != nil {
			return err
		}
	}

	return tw.Flush()
}

func (c *Config) writeGithubReleaseUsage(gru GithubReleaseUsage, tw *tabwriter.Writer) error {
	location := DefaultLocation
	if gru.Location != "" {
		location = gru.Location
	}
	rows := []string{
		fmt.Sprintf("%v\t", gru.Name),
		fmt.Sprintf("\t - Location: %v", location),
	}

	for _, r := range rows {
		_, err := fmt.Fprintln(tw, r)
		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Config) writeGithubRelease(ghr GithubReleaseConfiguration, tw *tabwriter.Writer) error {
	rows := []string{
		fmt.Sprintf("%v\t (https://github.com/%v)", ghr.name, ghr.RepoName),
		fmt.Sprintf("\t - DownloadType: %v", ghr.Download.Type),
	}

	if ghr.Download.Type == TarGz {
		rows = append(rows, fmt.Sprintf("\t - DownloadPath: %v", ghr.Download.Path))
	}

	oses := []string{}
	for os := range ghr.Patterns {
		oses = append(oses, os)
	}
	sort.Strings(oses)

	for _, os := range oses {
		rows = append(rows, fmt.Sprintf("\t - %v Pattern: %v", os, ghr.Patterns[os]))
	}

	for _, r := range rows {
		_, err := fmt.Fprintln(tw, r)
		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Config) GetGithubReleaseImplForTarget(target string) ([]Item, error) {
	impls := []Item{}

	current, ok := c.content.Hosts[target]
	if !ok {
		return impls, fmt.Errorf("Target %v: %w", target, NotFoundError)
	}

	for _, usage := range current.GithubRelease {
		obj := c.content.GithubReleases[usage.Name]
		pattern, ok := obj.Patterns[runtime.GOOS]
		if !ok {
			return impls, fmt.Errorf("%v does not have a pattern for %v: %w", usage.Name, runtime.GOOS, NotFoundError)
		}
		ghr := &GithubReleaseItem{
			Name:         usage.Name,
			Location:     usage.Location,
			RepoName:     obj.RepoName,
			DownloadType: obj.Download.Type,
			DownloadPath: obj.Download.Path,
			Pattern:      pattern,
		}
		if ghr.Location == "" {
			ghr.Location = util.ReplacePrefix(DefaultLocation, "~/", c.Home)
		}
		if c.GithubUser != "" {
			ghr.GithubUser = c.GithubUser
		}

		impls = append(impls, ghr)
	}

	return impls, nil
}
