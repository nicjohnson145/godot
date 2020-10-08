package main

import (
	"fmt"
	"github.com/tidwall/gjson"
)

type templateData struct {
	Host string
}

const myJson = `{"name": {"first": "Nic"}}`

func RenderTemplates(config godotConfig) {
	// outputRoot := filepath.Join(config.RepoPath, "compiled")
	// srcRoot := filepath.Join(config.RepoPath, "templates")
}

func readMappings() {
	fmt.Println(gjson.Get(myJson, "name.first"))
}
