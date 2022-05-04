package lib

import (
	"strings"
)

func replaceTilde(s string, replacement string) string {
	if !strings.Contains(s, "~") {
		return s
	}
	return strings.ReplaceAll(s, "~", replacement)
}
