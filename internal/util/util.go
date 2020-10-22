package util

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func IsFile(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsDir(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func Check(e error) {
	if e != nil {
		if val, _ := strconv.Atoi(os.Getenv("GODOT_PANIC")); val == 1 {
			log.Panicln(e)
		} else {
			fmt.Println(e)
			log.Fatalln(e)
		}
	}
}
