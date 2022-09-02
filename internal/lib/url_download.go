package lib

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"path"
	"runtime"
	"text/template"
)

var _ Executor = (*UrlDownload)(nil)

type UrlDownload struct {
	Name       string `yaml:"name"`
	Tag        string `yaml:"tag"`
	MacUrl     string `yaml:"mac-url"`
	LinuxUrl   string `yaml:"linux-url"`
	WindowsUrl string `yaml:"windows-url"`
}

type urlVars struct {
	Tag string
}

func (u *UrlDownload) GetName() string {
	return u.Name
}

func (u *UrlDownload) Type() ExecutorType {
	return ExecutorTypeUrlDownloads
}

func (u *UrlDownload) Execute(conf UserConfig, opts SyncOpts, _ TargetConfig) {
	log.Infof("Ensuring %v", u.Name)
	url := u.getDownloadUrl()
	err := downloadAndSymlinkBinary(downloadOpts{
		Name:         u.Name,
		DownloadName: path.Base(url),
		FinalDest:    getDestination(conf, u.Name, u.Tag),
		Url:          url,
		SymlinkName:  getSymlinkName(conf, u.Name, u.Tag),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (u *UrlDownload) getDownloadUrl() string {
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
		log.Fatalf("No download url specified for %v", runtime.GOOS)
	}

	// It could be a template, so parse it as such
	tmpl, err := template.New(u.Name).Parse(url)
	if err != nil {
		log.Fatalf("Error parsing download url as template: %v", err)
	}
	value := new(bytes.Buffer)
	err = tmpl.Execute(value, urlVars{
		Tag: u.Tag,
	})
	if err != nil {
		log.Fatalf("Error rendering URL template: %v", err)
	}

	return value.String()
}
