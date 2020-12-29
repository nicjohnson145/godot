package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type File struct {
	DestinationPath string
	TemplatePath    string
}

type TemplateVars struct {
	Target string
}

var funcs = template.FuncMap{
	"oneOf": func(vars TemplateVars, options ...string) bool {
		for _, opt := range options {
			if opt == vars.Target {
				return true
			}
		}
		return false
	},
}

func (f *File) Render(buildDir string, vars TemplateVars) error {
	tmplName := filepath.Base(f.TemplatePath)

	tpl := template.New(tmplName).Funcs(funcs)
	_, err := tpl.ParseFiles(f.TemplatePath)

	b := bytes.NewBufferString("")
	err = tpl.Execute(b, vars)
	if err != nil {
		err = fmt.Errorf("error executing template, %v", err)
		return err
	}

	destPath := filepath.Join(buildDir, tmplName)

	err = ioutil.WriteFile(destPath, b.Bytes(), 0700)
	if err != nil {
		err = fmt.Errorf("could not open %q for writing, %v", destPath, err)
		return err
	}

	return nil
}

func (f *File) Symlink(buildDir string) error {
	src := filepath.Join(buildDir, filepath.Base(f.TemplatePath))
	destbase := filepath.Dir(f.DestinationPath)
	err := os.MkdirAll(destbase, 0700)
	if err != nil {
		err = fmt.Errorf("unable to create dir %q, %v", destbase, err)
		return err
	}
	err = os.Symlink(src, f.DestinationPath)
	if err != nil {
		err = fmt.Errorf("unable to symlink %q to %q, %v", src, f.DestinationPath, err)
		return err
	}
	return err
}
