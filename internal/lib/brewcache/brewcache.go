package brewcache

import (
	"fmt"
	"strings"

	"github.com/nicjohnson145/godot/internal/help"
)

var instance *BrewCache

type BrewCache struct {
	Installed map[string]struct{}
}

func (b *BrewCache) populate() {
	out, _, err := help.RunCmd("brew", "list", "--formula", "-1")
	if err != nil {
		panic(fmt.Sprintf("Error initializing brew cache: %v", err))
	}

	for _, prog := range strings.Split(out, "\n") {
		b.Installed[prog] = struct{}{}
	}
}

func (b *BrewCache) Reset() {
	b.populate()
}

func (b *BrewCache) IsInstalled(prog string) bool {
	_, ok := b.Installed[prog]
	return ok
}

func newBrewCache() *BrewCache {
	b := &BrewCache{
		Installed: make(map[string]struct{}),
	}
	b.populate()
	return b
}

func GetInstance() *BrewCache {
	if instance == nil {
		instance = newBrewCache()
	}
	return instance
}
