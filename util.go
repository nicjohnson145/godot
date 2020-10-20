package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
		if val, _ := strconv.Atoi(os.Getenv("GODOT_PANIC")); val == 1 {
			log.Panicln(e)
		} else {
			fmt.Println(e)
			log.Fatalln(e)
		}
	}
}
