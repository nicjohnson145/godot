// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.0
// Revision: 919e61c0174b91303753ee3898569a01abb32c97
// Build Date: 2023-12-18T15:54:43Z
// Built By: goreleaser

package lib

import (
	"fmt"
	"strings"
)

const (
	// ExecutorTypeConfigFile is a ExecutorType of type config-file.
	ExecutorTypeConfigFile ExecutorType = "config-file"
	// ExecutorTypeGithubRelease is a ExecutorType of type github-release.
	ExecutorTypeGithubRelease ExecutorType = "github-release"
	// ExecutorTypeGitRepo is a ExecutorType of type git-repo.
	ExecutorTypeGitRepo ExecutorType = "git-repo"
	// ExecutorTypeSysPackage is a ExecutorType of type sys-package.
	ExecutorTypeSysPackage ExecutorType = "sys-package"
	// ExecutorTypeUrlDownload is a ExecutorType of type url-download.
	ExecutorTypeUrlDownload ExecutorType = "url-download"
	// ExecutorTypeBundle is a ExecutorType of type bundle.
	ExecutorTypeBundle ExecutorType = "bundle"
	// ExecutorTypeGolang is a ExecutorType of type golang.
	ExecutorTypeGolang ExecutorType = "golang"
	// ExecutorTypeGoInstall is a ExecutorType of type go-install.
	ExecutorTypeGoInstall ExecutorType = "go-install"
	// ExecutorTypeConfigDir is a ExecutorType of type config-dir.
	ExecutorTypeConfigDir ExecutorType = "config-dir"
	// ExecutorTypeNeovim is a ExecutorType of type neovim.
	ExecutorTypeNeovim ExecutorType = "neovim"
)

var ErrInvalidExecutorType = fmt.Errorf("not a valid ExecutorType, try [%s]", strings.Join(_ExecutorTypeNames, ", "))

var _ExecutorTypeNames = []string{
	string(ExecutorTypeConfigFile),
	string(ExecutorTypeGithubRelease),
	string(ExecutorTypeGitRepo),
	string(ExecutorTypeSysPackage),
	string(ExecutorTypeUrlDownload),
	string(ExecutorTypeBundle),
	string(ExecutorTypeGolang),
	string(ExecutorTypeGoInstall),
	string(ExecutorTypeConfigDir),
	string(ExecutorTypeNeovim),
}

// ExecutorTypeNames returns a list of possible string values of ExecutorType.
func ExecutorTypeNames() []string {
	tmp := make([]string, len(_ExecutorTypeNames))
	copy(tmp, _ExecutorTypeNames)
	return tmp
}

// String implements the Stringer interface.
func (x ExecutorType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ExecutorType) IsValid() bool {
	_, err := ParseExecutorType(string(x))
	return err == nil
}

var _ExecutorTypeValue = map[string]ExecutorType{
	"config-file":    ExecutorTypeConfigFile,
	"github-release": ExecutorTypeGithubRelease,
	"git-repo":       ExecutorTypeGitRepo,
	"sys-package":    ExecutorTypeSysPackage,
	"url-download":   ExecutorTypeUrlDownload,
	"bundle":         ExecutorTypeBundle,
	"golang":         ExecutorTypeGolang,
	"go-install":     ExecutorTypeGoInstall,
	"config-dir":     ExecutorTypeConfigDir,
	"neovim":         ExecutorTypeNeovim,
}

// ParseExecutorType attempts to convert a string to a ExecutorType.
func ParseExecutorType(name string) (ExecutorType, error) {
	if x, ok := _ExecutorTypeValue[name]; ok {
		return x, nil
	}
	return ExecutorType(""), fmt.Errorf("%s is %w", name, ErrInvalidExecutorType)
}

// MarshalText implements the text marshaller method.
func (x ExecutorType) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ExecutorType) UnmarshalText(text []byte) error {
	tmp, err := ParseExecutorType(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

// Set implements the Golang flag.Value interface func.
func (x *ExecutorType) Set(val string) error {
	v, err := ParseExecutorType(val)
	*x = v
	return err
}

// Get implements the Golang flag.Getter interface func.
func (x *ExecutorType) Get() interface{} {
	return *x
}

// Type implements the github.com/spf13/pFlag Value interface.
func (x *ExecutorType) Type() string {
	return "ExecutorType"
}
