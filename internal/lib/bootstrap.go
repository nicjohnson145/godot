package lib

import (
	"sort"
	"strings"
)

type Bootstrap map[string]BootstrapItem

func (b Bootstrap) MethodsString() string {
	keys := make([]string, 0, len(b))
	for key := range b {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return strings.Join(keys, ", ")
}
