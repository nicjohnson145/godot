package lib

import (
	"bytes"
	"context"
	"github.com/carlmjohnson/requests"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"text/template"
)

var _ Executor = (*GitRepo)(nil)

const (
	TypeUrlDownload = "url-download"
)

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

func (u *UrlDownload) Type() string {
	return TypeUrlDownload
}

func (u *UrlDownload) Execute(conf UserConfig, opts SyncOpts) {
	destination := getDestination(conf, u.Name, u.Tag)
	exists, err := pathExists(destination)
	if err != nil {
		log.Fatalf("Unable to check existance of %v: %v", destination, err)
	}
	if exists {
		log.Infof("%v already downloaded, skipping", u.Name)
		return
	}

	log.Infof("Downloading %v", u.Name)

	url := u.getDownloadUrl()

	dir, err := ioutil.TempDir("", "godot-")
	if err != nil {
		log.Fatal("Unable to make temp directory")
	}
	defer os.RemoveAll(dir)

	filepath := path.Join(dir, path.Base(url))
	req := requests.
		URL(url).
		ToFile(filepath)
	err = req.Fetch(context.TODO())
	if err != nil {
		log.Fatalf("Error downloading release asset: %v", err)
	}

	extractDir := path.Join(dir, "extract")
	binary := extractBinary(filepath, extractDir, "", nil)
	copyToDestination(binary, destination)
	createSymlink(destination, getSymlinkName(conf, u.Name, u.Tag))
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
