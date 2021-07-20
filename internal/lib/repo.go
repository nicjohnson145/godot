package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Repo interface {
	Pull() error
	Push() error
}

type shellGitRepo struct {
	Path   string
	gitDir string
}

func NewShellGitRepo(path string) shellGitRepo {
	return shellGitRepo{
		Path: path,
	}
}

func (r shellGitRepo) runCmd(args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	str := append([]string{"-C", r.Path}, args...)
	cmd := exec.Command("git", str...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (r shellGitRepo) dirtyFiles() ([]string, error) {
	stdout, _, err := r.runCmd("status", "--porcelain")
	if err != nil {
		return []string{}, err
	}

	var files []string
	for _, line := range strings.Split(stdout, "\n") {
		if line == "" {
			continue
		}
		files = append(files, line[3:])
	}
	return files, nil
}

func (r shellGitRepo) isWorkdirClean() (bool, error) {
	files, err := r.dirtyFiles()
	return len(files) == 0, err
}

func (r shellGitRepo) ensureClean() error {
	clean, err := r.isWorkdirClean()
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("Godot requires a clean workdir, stash/reset any manual changes")
	}
	return nil
}

func (r shellGitRepo) Pull() error {
	err := r.ensureClean()
	if err != nil {
		return err
	}

	_, _, err = r.runCmd("pull")
	return err
}

func (r shellGitRepo) Push() error {
	dirtyFiles, err := r.dirtyFiles()
	if err != nil {
		return err
	}
	if len(dirtyFiles) == 0 {
		return nil
	}

	_, _, err = r.runCmd("add", "-A")
	if err != nil {
		return err
	}

	_, _, err = r.runCmd("commit", "-m", "[godot]: update configuration")
	if err != nil {
		return err
	}

	_, _, err = r.runCmd("push")
	return err
}

type NoopRepo struct{}

func (n NoopRepo) Push() error { return nil }
func (n NoopRepo) Pull() error { return nil }
