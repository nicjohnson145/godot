package lib

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

var funcs = template.FuncMap{
	"oneOf": func(vars TemplateVars, options ...string) bool {
		return lo.Contains(options, vars.Target)
	},
	"notOneOf": func(vars TemplateVars, options ...string) bool {
		return !lo.Contains(options, vars.Target)
	},
}

const (
	funcNameVaultLookup = "VaultLookup"
	funcNameIsInstalled = "IsInstalled"
)

type TemplateVars struct {
	Target     string
	Submodules string
	Home       string
}

var _ Executor = (*ConfigFile)(nil)

type ConfigFile struct {
	Name         string `yaml:"-"`
	TemplateName string `yaml:"template-name" mapstructure:"template-name"`
	Destination  string `yaml:"destination" mapstructure:"destination"`
}

func (c *ConfigFile) GetName() string {
	return c.Name
}

func (c *ConfigFile) SetName(n string) {
	c.Name = n
}

func (c *ConfigFile) Type() ExecutorType {
	return ExecutorTypeConfigFile
}

func (c *ConfigFile) Execute(conf UserConfig, opts SyncOpts, godotConf GodotConfig) error {
	c.createVaultClosure(conf, opts)
	if err := c.createIsInstalledClosure(conf, godotConf); err != nil {
		return fmt.Errorf("error creating IsInstalled closure: %w", err)
	}

	log.Infof("Executing config file %v", c.TemplateName)
	tmpl, err := c.parseTemplate(conf.CloneLocation)
	if err != nil {
		return err
	}

	buildPath := path.Join(conf.BuildLocation, c.TemplateName)
	if err := ensureContainingDir(buildPath); err != nil {
		return err
	}
	f, err := os.OpenFile(buildPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		return fmt.Errorf("error opening destination file %v: %v", buildPath, err)
	}
	defer f.Close()

	err = tmpl.Execute(f, TemplateVars{
		Target:     conf.Target,
		Submodules: path.Join(conf.CloneLocation, "submodules"),
		Home:       conf.HomeDir,
	})
	if err != nil {
		return fmt.Errorf("error rendering template: %v", err)
	}

	dest := replaceTilde(c.Destination, conf.HomeDir)
	exists, err := c.pathExists(dest)
	if err != nil {
		return err
	}
	if exists {
		if err := c.removePath(dest); err != nil {
			return fmt.Errorf("error removing path: %w", err)
		}
	}

	if err := c.symlink(buildPath, dest); err != nil {
		return fmt.Errorf("error symlinking: %w", err)
	}

	return nil
}

func (c *ConfigFile) createVaultClosure(conf UserConfig, opts SyncOpts) {
	if _, ok := funcs[funcNameVaultLookup]; ok {
		return
	}

	if opts.NoVault {
		log.Debug("Creating no-op vault template func")
		funcs[funcNameVaultLookup] = func(path string, key string) (string, error) {
			return "NOT_USING_VAULT", nil
		}
		return
	}

	log.Debug("Creating vault client template func")
	funcs[funcNameVaultLookup] = func(path string, key string) (string, error) {
		if !conf.VaultConfig.Client.Initialized() {
			return "", fmt.Errorf("Template requires Valut to be set up")
		}

		val, err := conf.VaultConfig.Client.ReadKey(path, key)
		if err != nil {
			return "", err
		}

		return val, nil
	}
}

func (c *ConfigFile) createIsInstalledClosure(userConf UserConfig, godotConf GodotConfig) error {
	if _, ok := funcs[funcNameIsInstalled]; ok {
		return nil
	}

	executors, err := godotConf.ExecutorsForTarget(userConf.Target)
	if err != nil {
		return fmt.Errorf("error getting executors list: %w", err)
	}

	log.Debug("Creating IsInstalled template func")
	funcs[funcNameIsInstalled] = func(item string) bool {
		return lo.ContainsBy(executors, func(t Executor) bool {
			return t.GetName() == item
		})
	}

	return nil
}

func (c *ConfigFile) parseTemplate(dotfiles string) (*template.Template, error) {
	t := template.New(c.TemplateName).Funcs(funcs)
	t, err := t.ParseFiles(path.Join(dotfiles, "templates", c.TemplateName))
	if err != nil {
		return nil, fmt.Errorf("error parsing template file: %w", err)
	}
	return t, nil
}

func (c *ConfigFile) pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("error checking existance of %v: %v", path, err)
		}
	}
	return true, nil
}

func (c *ConfigFile) removePath(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("error deleting path %v: %v", path, err)
	}
	return nil
}

func (c *ConfigFile) symlink(source string, dest string) error {
	if err := ensureContainingDir(dest); err != nil {
		return err
	}

	err := os.Symlink(source, dest)
	if err != nil {
		return fmt.Errorf("error creating symlink: %v", err)
	}
	return nil
}
