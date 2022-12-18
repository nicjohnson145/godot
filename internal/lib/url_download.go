package lib

import (
	"bytes"
	"fmt"
	"path"
	"runtime"
	"text/template"

	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

var _ Executor = (*UrlDownload)(nil)

type UrlDownload struct {
	Name       string `yaml:"-"`
	Tag        string `yaml:"tag" mapstructure:"tag"`
	MacUrl     string `yaml:"mac-url" mapstructure:"mac-url"`
	LinuxUrl   string `yaml:"linux-url" mapstructure:"linux-url"`
	WindowsUrl string `yaml:"windows-url" mapstructure:"windows-url"`
}

type urlVars struct {
	Tag string
}

func (u *UrlDownload) GetName() string {
	return u.Name
}

func (u *UrlDownload) SetName(n string) {
	u.Name = n
}

func (u *UrlDownload) Type() ExecutorType {
	return ExecutorTypeUrlDownload
}

func (u *UrlDownload) Validate() error {
	var errs *multierror.Error

	if u.MacUrl == "" && u.LinuxUrl == "" && u.WindowsUrl == "" {
		errs = multierror.Append(errs, fmt.Errorf("one of mac-url, linux-url, or windows-url is required"))
	}

	return errs.ErrorOrNil()
}

func (u *UrlDownload) Execute(conf UserConfig, opts SyncOpts, _ GodotConfig) error {
	log.Infof("Ensuring %v", u.Name)
	url, err := u.getDownloadUrl()
	if err != nil {
		return fmt.Errorf("error getting url: %w", err)
	}

	dest, err := getDestination(conf, u.Name, u.Tag)
	if err != nil {
		return err
	}

	symlink, err := getSymlinkName(conf, u.Name, u.Tag)
	if err != nil {
		return err
	}

	err = downloadAndSymlinkBinary(downloadOpts{
		Name:         u.Name,
		DownloadName: path.Base(url),
		FinalDest:    dest,
		Url:          url,
		SymlinkName:  symlink,
	})
	if err != nil {
		return fmt.Errorf("error during download/symlink: %w", err)
	}
	return nil
}

func (u *UrlDownload) getDownloadUrl() (string, error) {
	var url string
	switch runtime.GOOS {
	case "linux":
		url = u.LinuxUrl
	case "windows":
		url = u.WindowsUrl
	case "darwin":
		url = u.MacUrl
	}

	if url == "" {
		return "", fmt.Errorf("eo download url specified for %v", runtime.GOOS)
	}

	// It could be a template, so parse it as such
	tmpl, err := template.New(u.Name).Parse(url)
	if err != nil {
		return "", fmt.Errorf("error parsing download url as template: %v", err)
	}
	value := new(bytes.Buffer)
	err = tmpl.Execute(value, urlVars{
		Tag: u.Tag,
	})
	if err != nil {
		return "", fmt.Errorf("error rendering URL template: %v", err)
	}

	return value.String(), nil
}
