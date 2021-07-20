package lib

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/nicjohnson145/godot/internal/util"
)

func (c *Config) GetAllFiles() StringMap {
	return c.content.Files
}

func (c *Config) GetFilesForTarget(target string) StringMap {
	if target == "" {
		target = c.Target
	}

	ret := make(StringMap)
	all := c.GetAllFiles()

	host, ok := c.content.Hosts[target]
	if !ok {
		return ret
	}

	for _, file := range host.Files {
		ret[file] = all[file]
	}

	return ret
}

func (c *Config) AddFile(template string, destination string) (string, error) {
	if template == "" {
		template = util.ToTemplateName(destination)
	}
	if c.IsKnownFile(template) {
		return "", errors.New(fmt.Sprintf("template name %q already exists", template))
	}
	newDest := util.ReplacePrefix(destination, c.Home, "~")
	c.content.Files[template] = newDest
	return template, nil
}

func (c *Config) AddTargetFile(target string, name string) error {
	if !c.IsKnownFile(name) {
		return fmt.Errorf("Unknown template of %q", name)
	}
	host, ok := c.content.Hosts[target]
	if !ok {
		host = Host{}
	}
	host.Files = append(host.Files, name)
	c.content.Hosts[target] = host
	return nil
}

func (c *Config) RemoveTargetFile(target string, name string) error {
	host, ok := c.content.Hosts[target]
	if !ok {
		return fmt.Errorf("unknown target %q", target)
	}

	newFiles, err := c.removeItem(host.Files, name)
	if err != nil {
		return fmt.Errorf("remove target file: %w", err)
	}
	host.Files = newFiles
	c.content.Hosts[target] = host
	return nil
}

func (c *Config) GetAllTemplateNames() []string {
	names := make([]string, 0, len(c.content.Files))
	for name := range c.content.Files {
		names = append(names, name)
	}
	return names
}

func (c *Config) GetAllTemplateNamesForTarget(target string) []string {
	host, ok := c.content.Hosts[target]
	if !ok {
		host = Host{}
	}
	return host.Files
}

func (c *Config) IsKnownFile(name string) bool {
	_, ok := c.content.Files[name]
	return ok
}

func (c *Config) GetTemplateFromFullPath(path string) (string, error) {
	for template, dest := range c.GetAllFiles() {
		fullDest := util.ReplacePrefix(dest, "~/", c.Home)
		if fullDest == path {
			return filepath.Join(c.TemplateRoot, template), nil
		}
	}
	return "", fmt.Errorf("Path %q is not managed by godot", path)
}

func (c *Config) ListAllFiles(w io.Writer) error {
	return c.writeStringMap(w, c.GetAllFiles())
}

func (c *Config) ListTargetFiles(target string, w io.Writer) error {
	return c.writeStringMap(w, c.GetFilesForTarget(target))
}
