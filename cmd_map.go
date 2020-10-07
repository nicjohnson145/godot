package main

import (
	"flag"
	"fmt"
	"os"
)

type cmdMap map[string]*flag.FlagSet

func (this *cmdMap) makeSubCmd(cmds ...string) {
	for _, cmd := range cmds {
		(*this)[cmd] = flag.NewFlagSet(cmd, flag.ExitOnError)
	}
}

func NewCmdMap() cmdMap {
	return make(cmdMap)
}

func (this *cmdMap) Parse(args []string) {
	if len(args) < 2 {
		fmt.Println("expected subcommand")
		os.Exit(1)
	}

	_, ok := (*this)[args[1]]
	if !ok {
		fmt.Println(fmt.Sprintf("Unknown subcommand %s", args[1]))
		os.Exit(1)
	}

	fmt.Println(fmt.Sprintf("Subcommand %s", args[1]))
}
