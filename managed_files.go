package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type managedFiles struct {
	Files    []_PathPair `json:"managed_files"`
	FilePath string      `json:"-"`
}

type _PathPair struct {
	DestPath   string `json:"dest_path"`
	SourcePath string `json:"src_path"`
	Group      string `json:"group"`
}

func newManagedFiles(config godotConfig) managedFiles {
	path := filepath.Join(config.RepoPath, "managed_files.json")
	data, _ := ioutil.ReadFile(path)
	var files managedFiles
	err := json.Unmarshal(data, &files)
	check(err)
	files.FilePath = path
	return files
}

func (this *managedFiles) AddFile(path string) {
	this.Files = append(
		this.Files,
		_PathPair{
			DestPath:   path,
			SourcePath: this.getSourcePath(path),
			Group:      "",
		},
	)
}

func (this *managedFiles) WriteFile() {
	file, err := json.MarshalIndent(this, "", "    ")
	check(err)
	err = ioutil.WriteFile(this.FilePath, file, 0664)
	check(err)
}

func (this *managedFiles) getSourcePath(path string) string {
	return ""
}
