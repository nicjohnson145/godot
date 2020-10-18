package main

import (
	"log"
	"os"
)

func isFile(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func isDir(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func check(e error) {
	if e != nil {
		log.Panicln(e)
	}
}
