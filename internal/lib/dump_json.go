package lib

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/tidwall/pretty"
)

func dumpJson(obj interface{}, path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	prettyContents := pretty.PrettyOptions(bytes, &pretty.Options{Indent: "    "})
	return ioutil.WriteFile(path, prettyContents, 0644)
}
