package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/nicjohnson145/godot/internal/config"
	"github.com/nicjohnson145/godot/internal/render"
	"github.com/nicjohnson145/godot/internal/managed_files"
)

type workerFunc func()

func getWorkerFunc(config config.GodotConfig, args []string) workerFunc {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)

	// Add command flags
	addSrc := addCmd.String("src", "", "Path to the source file to be added (required)")
	addAs := addCmd.String(
		"as",
		"",
		"Optional filename for the template (defaults to the same name as the file with '.' replaced with 'dot_')",
	)
	addGroup := addCmd.String(
		"group",
		"",
		"Option group name for template folder, defaults to the name of the file with '.' replaced with ''",
	)

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
			addFunc(config, *addSrc, *addAs, *addGroup)
		}
	default:
		msg := fmt.Sprintf("Unknown command %s", args[1])
		fmt.Println(msg)
		log.Fatalln(msg)
	}
	// Can't get here
	return nil
}

func runFunc(config config.GodotConfig) {
	render.RenderTemplates(config)
}

func addFunc(config config.GodotConfig, newSrc string, addAs string, addGroup string) {
	files := managed_files.NewManagedFiles(config)
	files.AddFile(newSrc, addAs, addGroup)
	for _, fl := range files.Files {
		fmt.Println(fmt.Sprintf("%+v", fl))
	}
	// files.WriteConfig()
}
