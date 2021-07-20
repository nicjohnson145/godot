package lib

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/nicjohnson145/godot/internal/util"
)

func (c *Config) GetAllBootstraps() map[string]Bootstrap {
	return c.content.Bootstraps
}

func (c *Config) GetBootstrapsForTarget(target string) map[string]Bootstrap {
	if target == "" {
		target = c.Target
	}

	all := c.GetAllBootstraps()
	ret := make(map[string]Bootstrap)

	host, ok := c.content.Hosts[target]
	if !ok {
		return ret
	}

	for _, key := range host.Bootstraps {
		ret[key] = all[key]
	}

	return ret
}

func (c *Config) GetBootstrapTargetsForTarget(target string) []string {
	if target == "" {
		target = c.Target
	}
	return c.content.Hosts[target].Bootstraps
}

func (c *Config) GetRelevantBootstrapImpls(target string) ([]Item, error) {
	impls := []Item{}
	var errs *multierror.Error

	for _, t := range c.GetBootstrapTargetsForTarget(target) {
		found := false
		for _, mgr := range c.PackageManagers {
			if item, ok := c.content.Bootstraps[t][mgr]; ok {
				switch mgr {
				case APT:
					impls = append(impls, NewAptItem(item.Name))
				case BREW:
					impls = append(impls, NewBrewItem(item.Name))
				case GIT:
					i := c.translateItemLocation(item)
					impls = append(impls, NewRepoItem(i.Name, i.Location))
				}
				found = true
				break
			}
		}
		if !found {
			errs = multierror.Append(
				errs,
				fmt.Errorf(
					"No suitable manager found for %v, %v's available managers are %v",
					t,
					t,
					strings.Join(c.getManagersForBootstrapItem(t), ", "),
				),
			)
		}
	}

	return impls, errs.ErrorOrNil()
}

func (c *Config) translateItemLocation(i BootstrapItem) BootstrapItem {
	location := i.Location
	if location != "" {
		location = util.ReplacePrefix(location, "~", c.Home)
	}
	return BootstrapItem{
		Name:     i.Name,
		Location: location,
	}
}

func (c *Config) getManagersForBootstrapItem(item string) []string {
	keys := make([]string, 0, len(c.content.Bootstraps[item]))
	for key := range c.content.Bootstraps[item] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *Config) ListAllBootstraps(w io.Writer) error {
	m := c.bootstrapToStringMap(c.GetAllBootstraps())
	return c.writeStringMap(w, m)
}

func (c *Config) ListBootstrapsForTarget(w io.Writer, target string) error {
	m := c.bootstrapToStringMap(c.GetBootstrapsForTarget(target))
	return c.writeStringMap(w, m)
}

func (c *Config) bootstrapToStringMap(m map[string]Bootstrap) StringMap {
	new := make(StringMap)
	for key, value := range m {
		new[key] = value.MethodsString()
	}
	return new
}

func (c *Config) AddBootstrapItem(item, manager, pkg, location string) {
	itemMap, ok := c.content.Bootstraps[item]
	if !ok {
		itemMap = make(map[string]BootstrapItem)
	}
	itemMap[manager] = BootstrapItem{Name: pkg, Location: location}
	c.content.Bootstraps[item] = itemMap
}

func (c *Config) isValidBootstrap(name string) bool {
	_, ok := c.content.Bootstraps[name]
	return ok
}

func (c *Config) AddTargetBootstrap(target string, name string) error {
	if !c.isValidBootstrap(name) {
		return fmt.Errorf("Unknown bootstrap item of %q", name)
	}

	current, ok := c.content.Hosts[target]
	if !ok {
		current = Host{}
	}
	current.Bootstraps = append(current.Bootstraps, name)
	c.content.Hosts[target] = current
	return nil
}

func (c *Config) RemoveTargetBootstrap(target string, name string) error {
	if target == "" {
		target = c.Target
	}

	host, ok := c.content.Hosts[target]
	if !ok {
		return fmt.Errorf("Unknown target of %q", target)
	}

	new, err := c.removeItem(host.Bootstraps, name)
	if err != nil {
		return err
	}
	host.Bootstraps = new
	c.content.Hosts[target] = host

	return nil
}
