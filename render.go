package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/tidwall/gjson"
)

type templateData struct {
	Target string
}

const templates = "templates"
const compiled = "compiled"

func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"oneOf": func(target string, options ...string) bool {
			for _, opt := range options {
				if target == opt {
					return true
				}
			}
			return false
		},
	}
}

func renderTemplates(config godotConfig) {
	outputRoot := filepath.Join(config.RepoPath, compiled)
	srcRoot := filepath.Join(config.RepoPath, templates)

	templateData := templateData{Target: config.TargetName}
	templateFuncs := getTemplateFuncs()

	if !isDir(outputRoot) {
		err := os.Mkdir(outputRoot, 0775)
		check(err)
	}
	for _, pair := range config.ManagedFiles {
		var src string
		if pair.Group != "" {
			src = filepath.Join(srcRoot, pair.Group)
		} else {
			src = srcRoot
		}
		pair.addDestRoot(outputRoot)
		pair.addSourceRoot(src)
		renderManagedFile(pair, templateData, templateFuncs)
	}
}

func renderManagedFile(pair PathPair, templateData templateData, templateFuncs template.FuncMap) {
	log.Println(fmt.Sprintf("Rendering - %+v", pair))
	if isFile(pair.DestPath) {
		e := os.Remove(pair.DestPath)
		check(e)
	}

	tmpl := template.New("thing")
	tmpl.Funcs(templateFuncs)
	data, err := ioutil.ReadFile(pair.SourcePath)
	check(err)
	tmpl.Parse(string(data))

	outFile, err := os.OpenFile(pair.DestPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	check(err)
	defer outFile.Close()

	err = tmpl.Execute(outFile, templateData)
	check(err)
}

const myJson = `{"name": {"first": "Nic"}}`

func readMappings() {
	fmt.Println(gjson.Get(myJson, "name.first"))
}
