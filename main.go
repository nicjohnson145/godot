package main

import (
	"os"
)

func main() {
	cmdMap := NewCmdMap()
	cmdMap.makeSubCmd("build", "install", "run")
	cmdMap.Parse(os.Args)
}
