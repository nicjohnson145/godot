package lib

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

var _ Executor = (*ConfigDir)(nil)

type ConfigDir struct {
	Name        string         `yaml:"-"`
	DirName     string         `yaml:"dir-name" mapstructure:"dir-name"`
	Destination string         `yaml:"destination" mapstructure:"destination"`
	log         zerolog.Logger `yaml:"-"`
}

func (c *ConfigDir) Execute(conf UserConfig, opts SyncOpts, godotConf GodotConfig) error {
	c.log.Info().Str("config-dir", c.DirName).Msg("ensuring config-dir")
	files, err := c.getFiles(conf)
	if err != nil {
		return err
	}

	for _, file := range files {
		configFile := ConfigFile{
			TemplateName: file,
			Destination:  filepath.Join(c.Destination, strings.TrimPrefix(file, c.DirName+"/")),
			NoTemplate:   true,
		}
		// Quiet the logging down so we dont get wierd spam from using a nested executor
		configFile.SetLogger(LoggerWithLevel(zerolog.WarnLevel))
		if err := configFile.Execute(conf, opts, godotConf); err != nil {
			return fmt.Errorf("error handling %v: %w", file, err)
		}
	}

	return nil
}

func (c *ConfigDir) getFiles(conf UserConfig) ([]string, error) {
	templatePath := filepath.Join(conf.CloneLocation, "templates")
	dirPath := filepath.Join(templatePath, c.DirName)
	paths := []string{}
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}
		paths = append(paths, relPath)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}
	return paths, nil
}

func (c *ConfigDir) Type() ExecutorType {
	return ExecutorTypeConfigDir
}

func (c *ConfigDir) Validate() error {
	var errs *multierror.Error

	if c.DirName == "" {
		errs = multierror.Append(errs, fmt.Errorf("dir-name is required"))
	}
	if c.Destination == "" {
		errs = multierror.Append(errs, fmt.Errorf("destination is required"))
	}

	return errs.ErrorOrNil()
}

func (c *ConfigDir) GetName() string {
	return c.Name
}

func (c *ConfigDir) SetName(name string) {
	c.Name = name
}

func (c *ConfigDir) SetLogger(log zerolog.Logger) {
	c.log = log
}
