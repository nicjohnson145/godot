package main

import (
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/util"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, "godot.log")
	logFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	util.Check(err)
	defer logFile.Close()
	log.SetOutput(logFile)

	config := config.GetGodotConfig()
	workerFunc := getWorkerFunc(config, os.Args)
	workerFunc()
}
