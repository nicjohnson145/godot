package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, "godot.log")
	logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	check(err)
	defer logFile.Close()
	log.SetOutput(logFile)

	config := getGodotConfig()
	workerFunc := getWorkerFunc(config, os.Args)
	workerFunc()
}
