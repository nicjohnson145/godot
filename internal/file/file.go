package file

import (
	"bytes"
	"errors"
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
	Target     string
	Submodules string
	Home       string
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
	"notOneOf": func(vars TemplateVars, options ...string) bool {
		for _, opt := range options {
			if opt == vars.Target {
				return false
			}
		}
		return true
	},
}

func (f *File) getTemplateName() string {
	return filepath.Base(f.TemplatePath)
}

func (f *File) Render(buildDir string, vars TemplateVars, force bool) error {
	err := f.checkFile(force)
	if err != nil {
		return err
	}
	b, err := f.Execute(vars)
	if err != nil {
		return err
	}
	return f.Write(buildDir, b)
}

func (f *File) Execute(vars TemplateVars) (*bytes.Buffer, error) {
	tmplName := f.getTemplateName()
	tpl := template.New(tmplName).Funcs(funcs)
	_, err := tpl.ParseFiles(f.TemplatePath)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBufferString("")
	err = tpl.Execute(b, vars)
	if err != nil {
		return nil, fmt.Errorf("executing: %w", err)
	}
	return b, nil
}

func (f *File) Write(buildDir string, buf *bytes.Buffer) error {
	destPath := filepath.Join(buildDir, f.getTemplateName())
	return ioutil.WriteFile(destPath, buf.Bytes(), 0600)
}

type fileState string

const (
	Symlink     fileState = "symlink"
	RegularFile fileState = "regular-file"
	NotFound    fileState = "not-found"
)

func (f *File) checkFile(force bool) error {
	state, err := f.getFileState()
	if err != nil {
		return err
	}
	switch state {
	case Symlink:
		return f.maybeRemoveFile(true)
	case RegularFile:
		return f.maybeRemoveFile(force)
	case NotFound:
		return nil
	default:
		return errors.New(fmt.Sprintf("Unknonw file state of %q", state))
	}
}

func (f *File) maybeRemoveFile(force bool) error {
	if force {
		err := os.Remove(f.DestinationPath)
		if err != nil {
			err = fmt.Errorf("unable to remove destination path %q, %v", f.DestinationPath, err)
			return err
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("Destination file %q already exists, use force to override", f.DestinationPath))
	}
}

func (f *File) getFileState() (fileState, error) {
	if info, err := os.Lstat(f.DestinationPath); err == nil {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return Symlink, nil
		} else if info.Mode().IsRegular() {
			return RegularFile, nil
		} else {
			return "", errors.New(fmt.Sprintf("%q is not a symlink or file", f.DestinationPath))
		}

	} else if os.IsNotExist(err) {
		return NotFound, nil
	} else {
		return "", err
	}
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
		return fmt.Errorf("unable to symlink %q to %q, %w", src, f.DestinationPath, err)
	}
	return err
}
