package bootstrap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var _ Item = (*GithubRelease)(nil)

type GithubRelease struct {
	Name         string
	Location     string
	RepoName     string
	Pattern      string
	DownloadType string
	DownloadPath string
}

type releaseResponse struct {
	Assets []release `json:"assets"`
}

type releaseErrorResponse struct {
	Message string `json:"message"`
}

type release struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

func (g *GithubRelease) Check() (bool, error) {
	_, err := os.Stat(filepath.Join(g.Location, g.Name))

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (g *GithubRelease) Install() error {
	release, err := g.getRelease()
	if err != nil {
		return err
	}

	resp, err := http.Get(release.DownloadUrl)
	if err != nil {
		return err
	}

	dest, err := os.CreateTemp("", "godot-ghr-download-*")
	if err != nil {
		return err
	}
	defer os.Remove(dest.Name())

	if err != nil {
		return err
	}

	_, err = io.Copy(dest, resp.Body)
	if err != nil {
		return err
	}

	switch g.DownloadType {
	case "tar_gz":
		err = g.handleTarGz(dest.Name(), release.Name)
	case "binary_only":
		err = g.handleBinaryOnly(dest.Name())
	}

	return err
}

func (g *GithubRelease) getRelease() (release, error) {
	var release release
	resp, err := g.makeRequest()
	if err != nil {
		return release, err
	}

	if resp.StatusCode != 200 {
		var respJson releaseErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&respJson)
		if err != nil {
			return release, fmt.Errorf("Error parsing non-200 API response: %v", err)
		}
		return release, fmt.Errorf("API Error: %v", respJson.Message)
	}

	var respJson releaseResponse
	err = json.NewDecoder(resp.Body).Decode(&respJson)
	if err != nil {
		return release, err
	}

	reg, err := regexp.Compile(g.Pattern)
	if err != nil {
		return release, err
	}

	found := false
	for _, r := range respJson.Assets {
		if reg.MatchString(r.Name) {
			release = r
			found = true
		}
	}

	if !found {
		return release, fmt.Errorf("No %v releases matching %v", g.Name, g.Pattern)
	}

	return release, nil
}

func (g *GithubRelease) makeRequest() (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%v/releases/latest", g.RepoName), nil)
	if err != nil {
		return nil, err
	}

	val, found := os.LookupEnv("GITHUB_PAT")
	if found {
		// TODO: this needs to be dynamic
		req.SetBasicAuth("nicjohnson145", val)
	}

	return client.Do(req)
}

func (g *GithubRelease) handleTarGz(tempPath string, releaseName string) error {
	file, err := os.Open(tempPath)
	if err != nil {
		return err
	}

	dest, err := ioutil.TempDir("", "godot-ghr-unpacked-*")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(dest)
	}()

	err = Untar(file, dest)
	if err != nil {
		return err
	}

	// Remove ".tar.gz" from filename
	minusExt := releaseName[0:len(releaseName) - 7]
	binary := filepath.Join(dest, minusExt, g.DownloadPath)

	return g.copyBinaryToDestination(binary)
}

func (g *GithubRelease) copyBinaryToDestination(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}

	dest, err := os.Create(filepath.Join(g.Location, g.Name))
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return os.Chmod(dest.Name(), 0755)
}

func (g *GithubRelease) handleBinaryOnly(tempPath string) error {
	return g.copyBinaryToDestination(tempPath)
}
