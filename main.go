package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Fiye/config"
	"github.com/Fiye/diff"
	"github.com/Fiye/scan"
	"github.com/Fiye/stats"
	"github.com/Fiye/tree"
	u "github.com/bcicen/go-units"
	"github.com/davecgh/go-spew/spew"
	"github.com/joomcode/errorx"
	term "github.com/nsf/termbox-go"
)

var (
	validCommands = []string{"scan", "report", "changes", "help", "diffTest", "rwBenchmark"}
)

func main() {
	// Setup
	err := config.Load()
	if err != nil {
		log.Fatal("[Fiye] failed to load config", err)
	}
	runPreviously := config.GetRunPreviously()

	err = scan.Load()
	if err != nil {
		log.Fatal("[Fiye] failed to load scans data", err)
	}

	err = term.Init()
	if err != nil {
		log.Fatal("[Fiye] failed to initialise module for reading STDIN", err)
	}
	defer term.Close()

	// Commands
	if len(os.Args) < 2 {
		log.Fatal("[Fiye] You must provide at least 1 argument to run a command")
	} else if os.Args[1] == "scan" && len(os.Args) < 3 {
		log.Fatal("[Fiye] You must provide at least 2 arguments to run the `scan` command")
	}

	var (
		command = os.Args[1]
		params  = os.Args[2:]
	)

	switch command {
	case "scan":
		err = scanDir(params, runPreviously)
		if err != nil {
			log.Fatal("[Fiye] failed to run scan", err)
		}
	case "report":
		err = report(params, runPreviously)
		if err != nil {
			log.Fatal("[Fiye] failed to run report", err)
		}
	case "changes":
		err = changes(runPreviously)
		if err != nil {
			log.Fatal("[Fiye] failed to run changes", err)
		}
	case "help":
		fmt.Print(getHelpString())
	case "diffTest":
		runDiffTest("/Users/ptelemachou/Library")
	case "rwBenchmark":
		runBinaryReadWriteTest()
	default:
		log.Fatal("[Fiye] Invalid argument provided must be one of: ", strings.Join(validCommands, ","))
	}

	// TODO: Remove this, just to deal with termbox wiping output atm
	time.Sleep(time.Minute * 1000)
}

func getHelpString() string {
	help := `# Available commands #
* scan [PATH] -c -d: 
	Runs a manual scan of a directory (storing the resulting tree in a file), additional
	args are:
		'-c': run a "comprehensive" scan

* report -l=[NUM_LARGEST_FILES] -d=[NUM_LARGEST_DUPLICATES]: 
	Reports on the data from the LAST scan. Additional args are:
		'-l': get the n largest files
		'-d': get the n largest duplicates

	NOTE: Can only report on duplicates if the last two scans are BOTH comprehensive

* help:
	Prints out this message
`
	return help
}

func scanDir(args []string, runPreviously bool) error {
	// First execution setup
	if !runPreviously {
		fmt.Println("Detected first execution of `Fiye`")
		newOutput, err := promptNewOutputDir()
		if err != nil {
			return err
		}
		config.SetScansOutputDir(newOutput)
		config.SetRunPreviously(true)
	}

	// Check dir is readable
	targetDir := args[0]
	_, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	// Should be comprehensive on: first scan for a dir AND if requested by user
	previousFullScans := scan.GetScansFull(targetDir)
	isComprehensive := (previousFullScans == nil || len((*previousFullScans).Records) == 0)
	if len(args) > 1 {
		if args[1] == "-c" {
			isComprehensive = true
		} else {
			return errors.New("invalid third argument provided, must be '-c' (fi provided)")
		}
	}

	// Walk the tree, write the scan to `ScansRecord` and disk
	fmt.Printf("Started traversing tree '%s'...\n", targetDir)
	timer := time.Now()
	newTree := tree.WalkGenerateTree(targetDir, 0, isComprehensive, nil)
	fmt.Printf("	Took %d ms to traverse the tree\n", time.Since(timer).Milliseconds())

	// Diff this scan with the previous full scan (if exists)
	if previousFullScans != nil && len((*previousFullScans).Records) > 0 {
		lastScanTime := ((*previousFullScans).Records)[len((*previousFullScans).Records)-1].TimeCompleted
		fmt.Printf("Detected an existing full scan, performed at: %s, running 'diff'...\n", lastScanTime.String())

		// 1. Read into memory
		lastTree, err := tree.ReadBinary(config.GetScansOutputDir() + scan.GetLastScanFilename(newTree.BasePath, false))

		fmt.Println(config.GetScansOutputDir() + scan.GetLastScanFilename(newTree.BasePath, false))
		if err != nil {
			fmt.Println("WARNING: Failed to read last local scan for 'diff'ing, may be corrupt or inaccessible")
		} else {
			// 2. Diff with new scan
			timer = time.Now()
			diff := diff.CompareTrees(&lastTree, &newTree)
			spew.Dump(diff)
			spew.Dump(len(lastTree.AllHash))
			spew.Dump(len(newTree.AllHash))
			fmt.Println(len(diff.Files), len(diff.Trees))
			scan.PrintLargestDiffs(10, diff)
			fmt.Printf("	Took %d ms to compare this tree with the last one\n", time.Since(timer).Milliseconds())

			// 3. Add new diff record (writing to disk in the process)
			err = scan.AddScansDiff(lastTree.BasePath, lastTree.Comprehensive && newTree.Comprehensive, diff)
			if err != nil {
				// TODO: Do something with error here
				scan.RevertScansDiff(lastTree.BasePath, lastTree.Comprehensive && newTree.Comprehensive, diff)
				fmt.Println("WARNING: Failed to record new diff on disk")
			}
		}
	}

	fmt.Println("Writing tree data to disk...")
	err = scan.AddScansFull(newTree)
	if err != nil {
		// TODO: Do something with error here
		scan.RevertScansFull(newTree)
		return errorx.Decorate(err, "failed to add scan information to record and/or local file")
	}

	fmt.Printf("Completed tree walk for '%s'!\n", newTree.BasePath)

	return nil
}

func report(args []string, runPreviously bool) error {
	// Check dir is readable
	targetDir := args[0]
	_, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	// Always does COMPREHENSIVE atm
	// TODO: Change this so we can read existing diffs to get data
	var (
		ws                     = stats.WalkStats{}
		isComprehensive        = false
		reportLargest    int64 = -1
		reportDuplicates int64 = -1
	)
	for _, v := range args {
		if strings.HasPrefix(v, "-l=") {
			parts := strings.Split(v, "=")
			reportLargest, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				panic(err)
			}
			fileList := make([]stats.BasicFile, reportLargest)
			ws.LargestFiles = &fileList
		} else if strings.HasPrefix(v, "-d=") {
			parts := strings.Split(v, "=")
			reportDuplicates, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				panic(err)
			}
			isComprehensive = true
			fileMap := make(map[string][]stats.BasicFile, reportDuplicates)
			ws.DuplicateMap = &fileMap
		}
	}

	// Walk the tree, write the scan to `ScansRecord` and disk
	fmt.Printf("Started traversing tree '%s'...\n", targetDir)
	timer := time.Now()
	newTree := tree.WalkGenerateTree(targetDir, 0, isComprehensive, &ws)
	fmt.Printf("	Took %d ms to traverse the tree\n", time.Since(timer).Milliseconds())

	fmt.Printf("REPORT GENERATED FOR TREE WITH ROOT '%s'\n", newTree.BasePath)
	fmt.Printf("Tree contains: %d files and is of size %d\n\n", newTree.NumFilesTotal, newTree.Size)

	limit := 10

	if ws.LargestFiles != nil {
		i := 0
		fmt.Printf("\n## The 10 largest files are: ##\n")
		for _, v := range *ws.LargestFiles {
			fmt.Printf("'%s': %d bytes\n", v.Path, v.Size)
			i++
			if i >= limit {
				break
			}
		}
	}

	if ws.DuplicateMap != nil {
		i := 0
		fmt.Printf("## The 10 largest duplicates are (other copies' names may differ): ##\n")
		for _, v := range ws.GetLargestDuplicates(limit) {
			fmt.Printf("'%s': %d * %d bytes = %d bytes\n", v[0].Path, len(v), v[0].Size, len(v)*int(v[0].Size))
			i++
			if i >= limit {
				break
			}
		}
	}

	return nil
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

func getHelpString() string {
	help := ` Available commands:
	* scan [PATH] -n=[NUM_LARGEST_FILES]: scans a directory, prints out the n largest files (default 10)
	* schedule [PATH] [INTERVAL]: schedules a for scanning by the 'dirt' daemon, on interval (e.g. 50m, 6h, etc)
	* view: view scans for a path, the user will be prompted further to select the PATH and scans to view/compare`

	return help
}

func runScan(path string, numLargestFiles int64) string {
	timer := time.Now()

	ws := stats.WalkStats{
		LargestFiles: &[]stats.BasicFile{},
	}
	_ = tree.WalkGenerateTree(path, 0, false, &ws)

	output := fmt.Sprintf("# SHOWING THE TOP %d LARGEST FILES\n", numLargestFiles)
	for i, f := range *ws.LargestFiles {
		if i > int(numLargestFiles) {
			break
		}

		val := u.NewValue(float64(f.Size), u.Byte)
		output += fmt.Sprintf("%s %f %s\n", f.Path, val.MustConvert(u.MegaByte).Float(), u.MegaByte.Name)
	}

	return output + fmt.Sprintf("Took: %dms", time.Since(timer).Milliseconds())
}

func runDiffTest(path string) {
	os.Remove("./s1_s2.diff")
	os.Remove("./s2.tree")

	fmt.Println("Started walk 1")
	a := time.Now()
	s1 := tree.WalkGenerateTree(path, 0, true, nil)
	fmt.Printf("Took %d ms to generate first tree\n", time.Since(a).Milliseconds())
	fmt.Printf("Finished walk 1, num files: %d\n", s1.NumFilesTotal)

	numMins := 5
	fmt.Printf("Sleeping for %d minute(s)...\n", numMins)
	time.Sleep(time.Duration(numMins) * time.Minute)

	fmt.Println("Started walk 2")
	s2 := tree.WalkGenerateTree(path, 0, true, nil)
	fmt.Printf("Finished walk 2, num files: %d\n", s2.NumFilesTotal)
	fmt.Printf("The size of the s2Hash is: %d\n", len(s2.AllHash))

	fmt.Println("Started diff")
	a = time.Now()
	d := diff.CompareTrees(&s1, &s2)
	fmt.Printf("Took %d ms to generate diff\n", time.Since(a).Milliseconds())
	fmt.Printf("Finished diff, diff num files: %d\n", len(d.Files))
	fmt.Printf("The size of the diffHash is: %d\n", len(d.AllHash))

	err := d.WriteBinary("./s1_s2.diff")
	if err != nil {
		panic(err)
	}

	s1_s2 := s1.DeepCopy()

	err = s1_s2.WriteBinary("./s2.tree")
	if err != nil {
		panic(err)
	}

	fmt.Println(len(d.Files))
	fmt.Println(len(d.Trees))
	a = time.Now()
	removeTree := diff.WalkAddDiff(&s1_s2, &d, &(s1_s2.AllHash), []diff.TreeDiff{}, []diff.FileDiff{})
	fmt.Printf("Took %d ms to add diff to tree\n", time.Since(a).Milliseconds())
	var smth diff.ScanDiff
	if removeTree {
		smth = diff.CompareTrees(nil, &s2)
	}
	smth = diff.CompareTrees(&s1_s2, &s2)
	smth1 := diff.CompareTrees(&s1, &s2)
	fmt.Println(len(smth.Files))
	fmt.Println(len(smth.Trees))
	fmt.Printf("The result of s1 + diff(s1, s2) == s2 is: %v\n", smth.Empty())
	fmt.Printf("The result of s1 == s2 is: %v\n", smth1.Empty())
	fmt.Printf("Num files in the added diff is: %d\n", s1_s2.NumFilesTotal)
}

func runBinaryReadWriteTest() {
	os.Remove("./s1.gob")

	path := "/Users/ptelemachou"
	fmt.Printf("BENCHMARK PERFORMED ON PATH '%s'\n\n", path)

	fmt.Println("READ/WRITE TEST OF 'SHALLOW' SCAN (NO SHA256 HASHES)")
	fmt.Println("started creating big tree")
	s1 := tree.WalkGenerateTree(path, 0, false, nil)
	fmt.Printf("AllHash for shallow tree size is: %d\n", len(s1.AllHash))
	fmt.Printf("walk took: %dms\n\n", s1.TimeTaken.Milliseconds())

	var fileSize int64 = 0

	fmt.Println("doing test write to get size")
	err := s1.WriteBinary("./s1.gob")
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
	err = s1.WriteBinary("./s1.gob")
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
	s1 = tree.WalkGenerateTree(path, 0, true, nil)
	fmt.Printf("There are %d files in this tree\n", s1.NumFilesTotal)
	fmt.Printf("AllHash for deep tree size is: %d\n", len(s1.AllHash))
	fmt.Printf("walk took: %dms\n\n", s1.TimeTaken.Milliseconds())

	fmt.Printf("Size in tree is: %d\n", s1.Size)

	fmt.Println("doing test write to get size")
	err = s1.WriteBinary("./s1.gob")
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
	err = s1.WriteBinary("./s1.gob")
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
