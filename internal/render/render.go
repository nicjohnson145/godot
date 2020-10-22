package render

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"github.com/nicjohnson145/godot/internal/util"
	"github.com/nicjohnson145/godot/internal/config"
)

type templateData struct {
	Target string
}

const TEMPLATES = "templates"
const COMPILED = "compiled"

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

func RenderTemplates(config config.GodotConfig) {
	outputRoot := filepath.Join(config.RepoPath, COMPILED)
	srcRoot := filepath.Join(config.RepoPath, TEMPLATES)

	templateData := templateData{Target: config.TargetName}
	templateFuncs := getTemplateFuncs()

	if !util.IsDir(outputRoot) {
		err := os.Mkdir(outputRoot, 0775)
		util.Check(err)
	}
	for _, pair := range config.ManagedFiles {
		var src string
		if pair.Group != "" {
			src = filepath.Join(srcRoot, pair.Group)
		} else {
			src = srcRoot
		}
		pair.AddDestRoot(outputRoot)
		pair.AddSourceRoot(src)
		renderManagedFile(pair, templateData, templateFuncs)
	}
}

func renderManagedFile(pair config.PathPair, templateData templateData, templateFuncs template.FuncMap) {
	log.Println(fmt.Sprintf("Rendering - %+v", pair))
	if util.IsFile(pair.DestPath) {
		e := os.Remove(pair.DestPath)
		util.Check(e)
	}

	tmpl := template.New("thing")
	tmpl.Funcs(templateFuncs)
	data, err := ioutil.ReadFile(pair.SourcePath)
	util.Check(err)
	tmpl.Parse(string(data))

	outFile, err := os.OpenFile(pair.DestPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	util.Check(err)
	defer outFile.Close()

	err = tmpl.Execute(outFile, templateData)
	util.Check(err)
}

