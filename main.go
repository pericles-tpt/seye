package main

import (
	"log"
	"os"
	"strings"

	"github.com/Fiye/command"
	"github.com/Fiye/config"
	"github.com/Fiye/records"
)

var (
	validCommands = []string{"scan", "report", "diff", "help"}
)

func main() {
	// Setup
	err := config.Load()
	if err != nil {
		log.Fatal("[Fiye] failed to load config", err)
	}
	runPreviously := config.GetRunPreviously()

	err = records.Load()
	if err != nil {
		log.Fatal("[Fiye] failed to load scan records", err)
	}

	// Commands
	if len(os.Args) < 2 {
		log.Fatal("[Fiye] You must provide at least 1 argument to run a command")
	} else if os.Args[1] == "scan" && len(os.Args) < 3 {
		log.Fatal("[Fiye] You must provide at least 2 additional arguments to run the `scan` command")
	} else if os.Args[1] == "diff" && len(os.Args) < 2 {
		log.Fatal("[Fiye] You must provide at least 1 additional argument to run the `diff` command")
	}

	var (
		cmd    = os.Args[1]
		params = os.Args[2:]
	)

	switch cmd {
	case "scan":
		err = command.Scan(params, runPreviously)
		if err != nil {
			log.Fatal("[Fiye] failed to run scan", err)
		}
	case "report":
		err = command.Report(params, runPreviously)
		if err != nil {
			log.Fatal("[Fiye] failed to run report", err)
		}
	case "diff":
		err = command.Diff(params)
		if err != nil {
			log.Fatal("[Fiye] failed to run changes", err)
		}
	case "help":
		command.Help()
	default:
		log.Fatal("[Fiye] Invalid argument provided must be one of: ", strings.Join(validCommands, ","))
	}
}
