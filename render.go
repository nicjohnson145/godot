package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
)

type templateData struct {
	Host string
}

const templates = "templates"
const compiled = "compiled"

func RenderTemplates(config godotConfig) {
	outputRoot := filepath.Join(config.RepoPath, compiled)
	srcRoot := filepath.Join(config.RepoPath, templates)

	if !isDir(outputRoot) {
		err := os.Mkdir(outputRoot, 0775)
		check(err)
	}
	for _, pair := range config.ManagedFiles {
		pair.addDestRoot(outputRoot)
		pair.addSourceRoot(srcRoot)
		renderManagedFile(pair)
	}
}

func renderManagedFile(pair PathPair) {
	fmt.Println(fmt.Sprintf("Rendering - %+v", pair))
}

const myJson = `{"name": {"first": "Nic"}}`

func readMappings() {
	fmt.Println(gjson.Get(myJson, "name.first"))
}
