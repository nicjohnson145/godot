package managed_files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/render"
	"github.com/nicjohnson145/godot/internal/util"
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

func NewManagedFiles(config config.GodotConfig) managedFiles {
	path := filepath.Join(config.RepoPath, "managed_files.json")
	data, _ := ioutil.ReadFile(path)
	var files managedFiles
	err := json.Unmarshal(data, &files)
	util.Check(err)
	files.RepoPath = config.RepoPath
	return files
}

func (this *managedFiles) AddFile(path string, addAs string, group string) {
	group = this.getGroupName(path, group)
	sourcePath := filepath.Join(this.RepoPath, render.TEMPLATES, group, this.getSourcePath(path, addAs))

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
	util.Check(err)
	err = ioutil.WriteFile(this.RepoPath, file, 0664)
	util.Check(err)
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
	outPath := filepath.Join(this.RepoPath, render.TEMPLATES, outgroup)
	if util.IsDir(outPath) || util.IsFile(outPath) {
		msg := fmt.Sprintf("Template folder '%s' already exists", outgroup)
		fmt.Println(msg)
		log.Fatalln(msg)
	}
	return
}
