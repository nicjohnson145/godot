package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type managedFiles struct {
	Files    []_PathPair `json:"managed_files"`
	RepoPath string      `json:"-"`
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
	files.RepoPath = config.RepoPath
	return files
}

func (this *managedFiles) AddFile(path string, addAs string, group string) {
	group = this.getGroupName(path, group)
	sourcePath := filepath.Join(this.RepoPath, TEMPLATES, group, this.getSourcePath(path, addAs))

	this.Files = append(
		this.Files,
		_PathPair{
			DestPath:   path,
			SourcePath: sourcePath,
			Group:      group,
		},
	)
}

func (this *managedFiles) WriteConfig() {
	file, err := json.MarshalIndent(this, "", "    ")
	check(err)
	err = ioutil.WriteFile(this.RepoPath, file, 0664)
	check(err)
}

func (this *managedFiles) getSourcePath(path string, addAs string) string {
	if addAs != "" {
		return addAs
	}
	return strings.ReplaceAll(filepath.Base(path), ".", "dot_")
}

func (this *managedFiles) getGroupName(path string, group string) (outgroup string) {
	if group != "" {
		outgroup = group
	} else {
		outgroup = strings.ReplaceAll(filepath.Base(path), ".", "")
	}
	outPath := filepath.Join(this.RepoPath, TEMPLATES, outgroup)
	if isDir(outPath) || isFile(outPath) {
		msg := fmt.Sprintf("Template folder '%s' already exists", outgroup)
		fmt.Println(msg)
		log.Fatalln(msg)
	}
	return
}
