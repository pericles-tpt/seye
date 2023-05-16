package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Fiye/tree"
	"github.com/boynton/repl"
)

type TestHandler struct {
	value string
}

//test incomplete lines by counting parens -- they must match.
func (th *TestHandler) Eval(expr string) (string, bool, error) {
	var (
		err    error
		args   = strings.Split(expr, " ")
		cmd    = args[0]
		params []string
	)

	if len(args) > 1 {
		params = args[1:]
	}

	switch cmd {
	case "scan":
		if len(params) < 1 {
			return "", false, errors.New("Not enough arguments in call to `scan`")
		}

		var numFiles int64 = 10
		if len(params) > 1 && strings.HasPrefix(args[2], "-n") {
			numFilesString := strings.Split(args[2], "=")[1]

			numFiles, err = strconv.ParseInt(numFilesString, 10, 64)
			if err != nil {
				return "", false, err
			}
		}

		output := runScan(args[1], numFiles)
		return output, false, nil
	case "schedule":
		return "unimplemented", true, nil
	case "view":
		return "unimplemented", true, nil
	case "help":
		return getHelpString(), false, nil
	default:
		return "command not recognised", true, nil
	}
}

func (th *TestHandler) Reset() {
	th.value = ""
}

func (th *TestHandler) Prompt() string {
	return "> "
}

func (th *TestHandler) Complete(expr string) (string, []string) {
	return "", []string{}
}

func (th *TestHandler) Start() []string {
	return []string{}
}

func (th *TestHandler) Stop(history []string) {
}

func main() {
	repl.REPL(new(TestHandler))
}

//
func getHelpString() string {
	help := ` Available commands:
	* scan [PATH] -n=[NUM_LARGEST_FILES]: scans a directory, prints out the n largest files (default 10)
	* schedule [PATH] [INTERVAL]: schedules a for scanning by the 'dirt' daemon, on interval (e.g. 50m, 6h, etc)
	* view: view scans for a path, the user will be prompted further to select the PATH and scans to view/compare`

	return help
}

func runScan(path string, numLargestFiles int64) string {
	timer := time.Now()

	largestFiles := []tree.LargeFile{}
	_, _, _, _, _ = tree.Walk(path, 0, false, &largestFiles)

	output := fmt.Sprintf("# SHOWING THE TOP %d LARGEST FILES\n", numLargestFiles)
	for i, f := range largestFiles {
		if i > int(numLargestFiles) {
			break
		}

		output += fmt.Sprintf("%s %db\n", f.FullName, f.Size)
	}

	return output + fmt.Sprintf("Took: %dms", time.Since(timer).Milliseconds())
}
