package lib

import (
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"text/template"
)

var funcs = template.FuncMap{
	"oneOf": func(vars TemplateVars, options ...string) bool {
		return lo.Contains(options, vars.Target)
	},
	"notOneOf": func(vars TemplateVars, options ...string) bool {
		return !lo.Contains(options, vars.Target)
	},
}

type TemplateVars struct {
	Target     string
	Submodules string
	Home       string
}

var _ Executor = (*ConfigFile)(nil)

type ConfigFile struct {
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
}

func (c ConfigFile) GetName() string {
	return c.Name
}

func (c ConfigFile) Execute(conf UserConfig) {
	tmpl := c.parseTemplate(conf.CloneLocation)

	buildPath := path.Join(conf.BuildLocation, c.Name)
	ensureContainingDir(buildPath)
	f, err := os.OpenFile(buildPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		log.Fatalf("Error opening destination file %v: %v", buildPath, err)
	}
	defer f.Close()

	err = tmpl.Execute(f, TemplateVars{
		Target:     conf.Target,
		Submodules: path.Join(conf.CloneLocation, "submodules"),
		Home:       conf.HomeDir,
	})
	if err != nil {
		log.Fatalf("Error rendering template: %v", err)
	}

	dest := replaceTilde(c.Destination, conf.HomeDir)
	if c.pathExists(dest) {
		c.removePath(dest)
	}

	c.symlink(buildPath, dest)
}

func (c ConfigFile) parseTemplate(dotfiles string) *template.Template {
	t := template.New(c.Name).Funcs(funcs)
	t, err := t.ParseFiles(path.Join(dotfiles, "templates", c.Name))
	if err != nil {
		log.Fatalf("Error parsing template file: %v", err)
	}
	return t
}

func (c ConfigFile) pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Fatalf("Error checking existance of %v: %v", path, err)
		}
	}
	return true
}

func (c ConfigFile) removePath(path string) {
	if err := os.Remove(path); err != nil {
		log.Fatalf("Error deleting path %v: %v", path, err)
	}
}

func (c ConfigFile) symlink(source string, dest string) {
	ensureContainingDir(dest)

	err := os.Symlink(source, dest)
	if err != nil {
		log.Fatalf("Error creating symlink: %v", err)
	}
}