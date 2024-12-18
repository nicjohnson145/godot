package lib

import (
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

func NewGodotConfig(location string) (GodotConfig, error) {
	fBytes, err := os.ReadFile(location)
	if err != nil {
		return GodotConfig{}, err
	}

	var conf GodotConfig
	err = yaml.Unmarshal(fBytes, &conf)
	if err != nil {
		return GodotConfig{}, err
	}

	if err := conf.Validate(); err != nil {
		return GodotConfig{}, err
	}
	conf.SetExecutorNames()

	return conf, nil
}

func NewGodotConfigFromUserConfig(conf UserConfig) (GodotConfig, error) {
	return NewGodotConfig(path.Join(conf.CloneLocation, "config.yaml"))
}

type GodotExecutor struct {
	Name string
	Type ExecutorType   `json:"type"`
	Spec map[string]any `json:"spec"`
}

//nolint:ireturn
func decodeStructure[T any](x T, spec map[string]any, typeName string) (T, error) {
	if err := mapstructure.Decode(spec, &x); err != nil {
		return *new(T), fmt.Errorf("error decoding as %v: %w", typeName, err)
	}
	return x, nil
}

//nolint:ireturn,gocyclo
func (r *GodotExecutor) AsExecutor() (Executor, error) {
	var executor Executor
	var err error

	switch r.Type {
	case ExecutorTypeConfigFile:
		var x ConfigFile
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeGitRepo:
		var x GitRepo
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeGithubRelease:
		var x GithubRelease
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeSysPackage:
		var x SystemPackage
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeUrlDownload:
		var x UrlDownload
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeBundle:
		var x Bundle
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeGolang:
		var x Golang
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeGoInstall:
		var x GoInstall
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeConfigDir:
		var x ConfigDir
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	case ExecutorTypeNeovim:
		var x Neovim
		executor, err = decodeStructure(&x, r.Spec, r.Type.String())
	default:
		return nil, fmt.Errorf("programming error: unhandled executor type of '%v' with name '%v'", r.Type, r.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("error decoding spec: %w", err)
	}
	executor.SetName(r.Name)
	return executor, nil
}

type GodotConfig struct {
	Executors map[string]GodotExecutor `json:"executors"`
	Targets   map[string][]string      `json:"targets"`
}

func (r *GodotConfig) Validate() error {
	var errors *multierror.Error

	for target, selected := range r.Targets {
		for _, s := range selected {
			if _, ok := r.Executors[s]; !ok {
				errors = multierror.Append(
					errors,
					fmt.Errorf("error with target %v: unknown executor %v", target, s),
				)
			}
		}
	}

	for name, rawEx := range r.Executors {
		ex, err := rawEx.AsExecutor()
		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("error with executor %v: %w", name, err))
			continue
		}
		if err := ex.Validate(); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("executor %v is invalid: %w", name, err))
		}
		// Bundles work a little differently
		if rawEx.Type == ExecutorTypeBundle {
			if err := r.validateBundle(name, ex); err != nil {
				errors = multierror.Append(errors, err)
			}
		}
	}

	return errors.ErrorOrNil()
}

func (r *GodotConfig) validateBundle(name string, ex Executor) error {
	var errors *multierror.Error

	bndl, ok := ex.(*Bundle)
	if !ok {
		errors = multierror.Append(errors, fmt.Errorf("executor %v has type bundle, but cannot cast", name))
	}
	for _, item := range bndl.Items {
		if _, ok := r.Executors[item]; !ok {
			errors = multierror.Append(errors, fmt.Errorf("bundle %v has unknown item %v", name, item))
		}
	}

	return errors.ErrorOrNil()
}

func (r *GodotConfig) SetExecutorNames() {
	for name := range r.Executors {
		ex := r.Executors[name]
		ex.Name = name
		r.Executors[name] = ex
	}
}

func (r *GodotConfig) ExecutorsForTarget(name string) ([]Executor, error) {
	return r.fetchExecutorsForSlice(r.Targets[name])
}

func (r *GodotConfig) fetchExecutorsForSlice(selection []string) ([]Executor, error) {
	executors := []Executor{}
	for _, name := range selection {
		rawEx := r.Executors[name]
		ex, err := rawEx.AsExecutor()
		if err != nil {
			return nil, err
		}
		executors = append(executors, ex)
		if rawEx.Type == ExecutorTypeBundle {
			bEx := ex.(*Bundle)
			subExecs, err := r.fetchExecutorsForSlice(bEx.Items)
			if err != nil {
				return nil, err
			}
			executors = append(executors, subExecs...)
		}
	}
	return applyOrdering(deduplicate(executors)), nil
}
