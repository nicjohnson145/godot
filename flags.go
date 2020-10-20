package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type subCmd struct {

}
type workerFunc func()

func getWorkerFunc(config godotConfig, args []string) workerFunc {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)

	// Add command flags
	addSrc := addCmd.String("src", "", "Path to the source file to be added (required)")

	runWrapper := func() {
		runFunc(config)
	}

	if len(args) < 2 {
		return runWrapper
	}

	switch args[1] {
		case "run":
			runCmd.Parse(args[2:])
			return runWrapper
		case "add":
			addCmd.Parse(args[2:])

			if *addSrc == "" {
				addCmd.PrintDefaults()
				os.Exit(1)
			}

			return func() {
				addFunc(config, *addSrc)
			}
		default:
			msg := fmt.Sprintf("Unknown command %s", args[1])
			fmt.Println(msg)
			log.Fatalln(msg)
	}
	// Can't get here
	return nil
}

func runFunc(config godotConfig) {
	renderTemplates(config)
}

func addFunc(config godotConfig, newSrc string) {
	files := newManagedFiles(config)
	files.AddFile(newSrc)
	for _, fl := range files.Files {
		fmt.Println(fmt.Sprintf("%+v", fl))
	}
	files.WriteFile()
}
