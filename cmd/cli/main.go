package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Fiye/diff"
	"github.com/Fiye/tree"
	u "github.com/bcicen/go-units"
	"github.com/boynton/repl"
	"github.com/davecgh/go-spew/spew"
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
	case "diffTest":
		spew.Dump([32]byte{})
		runDiffTest("/home/pt")
		return "", false, nil
	case "readWriteBenchmark":
		runBinaryReadWriteTest()
		return "", false, nil
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

		val := u.NewValue(float64(f.Size), u.Byte)
		output += fmt.Sprintf("%s %fgb\n", f.FullName, val.MustConvert(u.GigaByte).Float())
	}

	return output + fmt.Sprintf("Took: %dms", time.Since(timer).Milliseconds())
}

func runDiffTest(path string) {
	fmt.Println("Started walk 1")
	s1, _, _, _, _ := tree.Walk(path, 0, true, nil)
	fmt.Printf("Finished walk 1, num files: %d\n", s1.NumFilesTotal)

	fmt.Println("Sleeping for 1 minute...")
	time.Sleep(1 * time.Minute)

	fmt.Println("Started walk 2")
	s2, _, _, _, _ := tree.Walk(path, 0, true, nil)
	fmt.Printf("Finished walk 2, num files: %d\n", s2.NumFilesTotal)

	fmt.Println("Started diff")
	d := diff.CompareTrees(&s1, &s2)
	fmt.Printf("Finished diff, diff num files: %d\n", d.NumFilesTotalDiff)

	s1_s2 := diff.WalkAddDiff(s1, d)
	fmt.Printf("The result of s1 + diff(s1, s2) == s2 is: %v\n", diff.TreeDiffEmpty(diff.CompareTrees(&s1_s2, &s2)))
	fmt.Printf("Num files in the added diff is: %d\n", s1_s2.NumFilesTotal)
}

func runBinaryReadWriteTest() {
	os.Remove("./s1.gob")

	path := "/home/pt"
	fmt.Printf("BENCHMARK PERFORMED ON PATH '%s'\n\n", path)

	fmt.Println("READ/WRITE TEST OF 'SHALLOW' SCAN (NO SHA256 HASHES)")
	fmt.Println("started creating big tree")
	s1, _, _, _, _ := tree.Walk(path, 0, false, nil)
	fmt.Printf("walk took: %dms\n\n", s1.TimeTaken.Milliseconds())

	var fileSize int64 = 0

	fmt.Println("doing test write to get size")
	err := tree.WriteBinary(s1, "./s1.gob")
	if err != nil {
		panic(err)
	}
	f, _ := os.Stat("./s1.gob")
	fileSize = f.Size()
	os.Remove("./s1.gob")
	units, _ := u.ConvertFloat(float64(fileSize), u.Byte, u.MegaByte)
	fmt.Printf("tree is of size %.2f %s\n\n", units.Float(), units.Unit().PluralName())

	fmt.Println("started writing big tree (actual)")
	timer := time.Now()
	err = tree.WriteBinary(s1, "./s1.gob")
	if err != nil {
		panic(err)
	}
	endTimer := time.Since(timer)
	fmt.Printf("write took: %dms, at a speed of %.2f %s/s\n\n", endTimer.Milliseconds(), units.Float()/(float64(endTimer.Nanoseconds())/math.Pow(10, 9)), units.Unit().Name)

	fmt.Println("started reading big tree into memory")
	timer = time.Now()
	s2, err := tree.ReadBinary("./s1.gob")
	if err != nil {
		panic(err)
	}
	endTimer = time.Since(timer)
	fmt.Printf("read took: %dms, at a speed of %.2f %s/s\n", endTimer.Milliseconds(), (units.Float() / (float64(endTimer.Nanoseconds()) / math.Pow(10, 9))), units.Unit().Name)
	s2.Depth += 1
	os.Remove("./s1.gob")

	fmt.Println("\nREAD/WRITE TEST OF 'FULL' SCAN (INCLUDES SHA256 HASHES)")
	fmt.Println("started creating big tree")
	s1, _, _, _, _ = tree.Walk(path, 0, true, nil)
	fmt.Printf("walk took: %dms\n\n", s1.TimeTaken.Milliseconds())

	fmt.Println("doing test write to get size")
	err = tree.WriteBinary(s1, "./s1.gob")
	if err != nil {
		panic(err)
	}
	f, _ = os.Stat("./s1.gob")
	fileSize = f.Size()
	os.Remove("./s1.gob")
	units, _ = u.ConvertFloat(float64(fileSize), u.Byte, u.MegaByte)
	fmt.Printf("tree is of size %.2f %s\n\n", units.Float(), units.Unit().PluralName())

	fmt.Println("started writing big tree (actual)")
	timer = time.Now()
	err = tree.WriteBinary(s1, "./s1.gob")
	if err != nil {
		panic(err)
	}
	endTimer = time.Since(timer)
	fmt.Printf("write took: %dms, at a speed of %.2f %s/s\n\n", endTimer.Milliseconds(), units.Float()/(float64(endTimer.Nanoseconds())/math.Pow(10, 9)), units.Unit().Name)

	fmt.Println("started reading big tree into memory")
	timer = time.Now()
	s2, err = tree.ReadBinary("./s1.gob")
	if err != nil {
		panic(err)
	}
	endTimer = time.Since(timer)
	fmt.Printf("read took: %dms, at a speed of %.2f %s/s\n", endTimer.Milliseconds(), (units.Float() / (float64(endTimer.Nanoseconds()) / math.Pow(10, 9))), units.Unit().Name)
	s2.Depth += 1
	os.Remove("./s1.gob")
}
