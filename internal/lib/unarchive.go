package lib

import (
	"github.com/mholt/archiver"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"path/filepath"
	"fmt"
)

type searchFunc func(string) (bool, error)

var archiveExtensions = []string{
	".gz",
	".zip",
}

func isArchiveFile(path string) bool {
	return lo.Contains(archiveExtensions, filepath.Ext(path))
}

func isExecutableFile(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error determining if file is executable: %w", err)
	}

	filePerm := fileInfo.Mode()
	return !fileInfo.IsDir() && filePerm&0111 != 0, nil
}

func extractBinary(downloadPath string, extractPath string, binaryPath string, findFunc searchFunc) (string, error) {
	if isArchiveFile(downloadPath) {
		err := archiver.Unarchive(downloadPath, extractPath)
		if err != nil {
			return "", fmt.Errorf("error extracting archive: %w", err)
		}
		if binaryPath != "" {
			return filepath.Join(extractPath, binaryPath), nil
		} else {
			return findExecutable(extractPath, findFunc)
		}
	}
	return downloadPath, nil
}

func findExecutable(path string, searchFunc searchFunc) (string, error) {
	search := isExecutableFile
	if searchFunc != nil {
		search = searchFunc
	}
	executables := []string{}
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		found, err := search(path)
		if err != nil {
			return err
		}
		if found {
			executables = append(executables, path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking extracted directory tree: %w", err)
	}

	if len(executables) != 1 {
		return "", fmt.Errorf("expected to find 1 executable, instead found %v, please specify the binary path manually", len(executables))
	}

	return executables[0], nil
}
