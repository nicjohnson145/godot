package lib

import (
	"github.com/mholt/archiver"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
)

var archiveExtensions = []string{
	".gz",
	".zip",
}

func isArchiveFile(path string) bool {
	return lo.Contains(archiveExtensions, filepath.Ext(path))
}

func isExecutableFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Error determining if file is executable: %v", err)
	}

	filePerm := fileInfo.Mode()
	return !fileInfo.IsDir() && filePerm&0111 != 0
}

func extractBinary(downloadPath string, extractPath string, binaryPath string) string {
	if isArchiveFile(downloadPath) {
		err := archiver.Unarchive(downloadPath, extractPath)
		if err != nil {
			log.Fatalf("Error extracting archive: %v", err)
		}
		if binaryPath != "" {
			return filepath.Join(extractPath, binaryPath)
		} else {
			return findExecutable(extractPath)
		}
	}
	return downloadPath
}

func findExecutable(path string) string {
	executables := []string{}
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if isExecutableFile(path) {
			executables = append(executables, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking extracted directory tree: %v", err)
	}

	if len(executables) != 1 {
		log.Fatalf("Expected to find 1 executable, instead found %v, please specify the binary path manually", len(executables))
	}

	return executables[0]
}
