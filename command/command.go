package command

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Fiye/config"
	"github.com/Fiye/diff"
	"github.com/Fiye/records"
	"github.com/Fiye/stats"
	"github.com/Fiye/tree"
	"github.com/joomcode/errorx"
)

func Help() {
	fmt.Print(`usage: seye [-s | --scan] [-r | --report] [-d | --diff] [-h | --help]
Parameters for the commands above:
	scan [PATH]: Runs a manual scan of a directory (storing the resulting tree in a file)
		'-c=false'     : forces either a "comprehensive" (true) or "shallow" (false) scan
		'-s'           : forces a "shallow" scan
		'-n=2'         : lets you specify number of processing threads to run
		'-label=setup' : lets you assign a label for the scan (can't be a whole number)
		'-p'           : prints out additional performance information (files found, etc)

	* NOTE_1: Scans between the initial and last scan for a directory are stored as "file-
		tree diffs", to reduce disk usage	
	* NOTE_2: Comprehensive scans can take 2-3x the time (or longer) as "shallow" scans. Scan
		duration depends on multiple factors: num files, avg file size, disk speed, etc

	report: Reports on the data from the LAST records. Additional args are
		'-l=10'        : get the n largest files
		'-d=10'        : get the n largest duplicates

	diff [PATH]: Gets the difference of two prior scans (currently only supports "diff"ing the first and last scan)

	* NOTE: Can only report on duplicates if the last two scans are BOTH comprehensive

	help: Prints this help text
`)
}

func Scan(args []string, runPreviously bool) error {
	// First execution setup, ask for output directory for tree scans
	if !runPreviously {
		fmt.Println("Detected first execution of `Fiye`")
		newOutput, err := promptNewOutputDir()
		if err != nil {
			return err
		}
		fmt.Println("New output is: ", newOutput)
		config.SetScansOutputDir(newOutput)

		outDir := config.GetScansOutputDir()
		_, err = os.Stat(outDir)
		if os.IsNotExist(err) {
			err = os.Mkdir(outDir, 0600)
			if err != nil {
				return errorx.Decorate(err, "failed to create directory '%s' for adding FileTree scans", outDir)
			}
		} else {
			return errorx.Decorate(err, "unexpected error when accessing scan directory '%s'", outDir)
		}

		config.SetRunPreviously(true)
	}

	// Check provided directory is readable
	targetDir := args[0]
	_, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	// Should set "Comprehensive" ON when: it's the first scan for a dir OR requested by user
	previousFullScans := records.GetScansFull(targetDir)
	// isComprehensive := (previousFullScans == nil || len((*previousFullScans).Records) == 0)
	isComprehensive := false
	if len(args) > 1 {
		if args[1] == "-c" {
			isComprehensive = true
		} else {
			return errors.New("invalid third argument provided, must be '-c' (fi provided)")
		}
	}

	// Walk the tree, write the scan to `ScansRecord` and disk
	fmt.Printf("Started traversing tree '%s'... ", targetDir)
	timer := time.Now()
	newTree := tree.WalkTreeIterativeFile(targetDir, 0, isComprehensive, nil)
	fmt.Printf("Took %d ms to traverse the tree\n", time.Since(timer).Milliseconds())

	// Diff this scan with the previous full scan (if one exists)
	if previousFullScans != nil && len((*previousFullScans).Records) > 0 {
		lastScanTime := ((*previousFullScans).Records)[len((*previousFullScans).Records)-1].TimeCompleted
		fmt.Printf("Detected an existing full scan, performed at: %s, running 'diff'... ", lastScanTime.String())

		// 1. Read previous scan into memory
		lastTree, err := tree.ReadBinary(config.GetScansOutputDir() + records.GetLastScanFilename(newTree.BasePath, false))
		if err != nil {
			fmt.Println("WARNING: Failed to read last local scan for 'diff'ing, may be corrupt or inaccessible")
		} else {
			// 2. Diff with new scan
			timer = time.Now()
			tDiff := diff.CompareTrees(&lastTree, newTree)
			fmt.Printf("Took %d ms to run diff comparing this tree with the last one\n", time.Since(timer).Milliseconds())

			// 3. Add new diff record (writing to disk in the process)
			err = records.AddDiffScanRecord(lastTree.BasePath, lastTree.Comprehensive && newTree.Comprehensive, tDiff)
			if err != nil {
				// TODO: Do something with error here
				records.RevertDiffScanRecord(lastTree.BasePath, lastTree.Comprehensive && newTree.Comprehensive, tDiff)
				fmt.Println("WARNING: Failed to record new diff to disk")
			}
		}
	}
	fmt.Println("Writing tree data to disk...")
	err = records.AddFullScanRecord(*newTree)
	if err != nil {
		// TODO: Do something with error here
		records.RevertFullScanRecord(*newTree)
		return errorx.Decorate(err, "failed to add scan information to record and/or local file")
	}

	return nil
}

func Report(args []string, runPreviously bool) error {
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
			fmt.Println("NOTE: Duplicate finding requested, performing a 'Comprehensive' scan, this may take a while")
			fileMap := make(map[string][]stats.BasicFile, reportDuplicates)
			ws.DuplicateMap = &fileMap
		}
	}

	// Walk the tree, write the scan to `ScansRecord` and disk
	// TODO: Write the resultant `newTree` to disk, and perform diffs if possible
	fmt.Printf("Started traversing tree '%s'...\n", targetDir)
	timer := time.Now()
	newTree := tree.WalkTreeIterativeFile(targetDir, 0, isComprehensive, &ws)
	fmt.Printf(" Took %d ms to traverse the tree", time.Since(timer).Milliseconds())

	fmt.Printf("REPORT GENERATED FOR TREE WITH ROOT '%s'\n", newTree.BasePath)
	fmt.Printf("Tree contains: %d files and is of size %d\n\n", newTree.NumFilesBelow, newTree.SizeBelow)

	if ws.LargestFiles != nil {
		i := 0
		fmt.Printf("\n## The %d largest files are: ##\n", reportLargest)
		for _, v := range *ws.LargestFiles {
			fmt.Printf("'%s': %d bytes\n", v.Path, v.Size)
			i++
			if i >= int(reportDuplicates) {
				break
			}
		}
	}

	if ws.DuplicateMap != nil {
		i := 0
		fmt.Printf("## The %d largest duplicates are (other copies' names may differ): ##\n", reportDuplicates)
		for _, v := range ws.GetLargestDuplicates(int(reportLargest)) {
			fmt.Printf("'%s': %d * %d bytes = %d bytes\n", v[0].Path, len(v), v[0].Size, len(v)*int(v[0].Size))
			i++
			if i >= int(reportLargest) {
				break
			}
		}
	}

	return nil
}

func Diff(args []string) error {
	var err error
	scans := records.GetAllScansFull()
	if scans == nil {
		return errors.New("no diffs available")
	}

	targetDir := args[0]
	if _, ok := (*scans)[targetDir]; !ok {
		return errors.New("cannot perform diff, no prior scans exist to diff")
	}

	diffScans := (*scans)[targetDir]
	if len(diffScans.Records) < 2 {
		return fmt.Errorf("cannot get difference between scans for directory '%s', not enough scan to perform diff, have: %d, need: 2", targetDir, len(diffScans.Records))

	}

	// Take the difference of the first and last scans
	// TODO: Allow this to take the difference of ANY two prior scans
	first, err := tree.ReadBinary(config.GetScansOutputDir() + records.GetScanFilename(targetDir, 0, false))
	if err != nil {
		return err
	}

	last, err := tree.ReadBinary(config.GetScansOutputDir() + records.GetLastScanFilename(targetDir, false))
	if err != nil {
		return err
	}

	// Finally print the largest 10 differences
	// TODO: Allow this parameter to be user specified in the future
	sdiff := diff.CompareTrees(&first, &last)
	for _, t := range sdiff.Trees {
		fmt.Println(t.NewerPath)
	}
	diff.PrintLargestDiffs(10, sdiff)

	return nil
}

func promptNewOutputDir() (string, error) {
	var (
		newOutputDir   string
		outputDirValid bool
	)

	fmt.Printf("Do you want to change your output path for directory scans? Default:\n\t%s\nType a new VALID path to set a different output path OR just press ENTER to use the default", config.GetScansOutputDir())

	for !outputDirValid {
		fmt.Scanln(&newOutputDir)

		if newOutputDir == "" {
			newOutputDir = config.GetScansOutputDir()
			break
		} else {
			fmt.Println("Output dir is NOT empty")
			_, err := os.ReadDir(newOutputDir)
			outputDirValid = (err == nil)
		}

		// 3. If confirmed return
		if outputDirValid {
			fmt.Printf("Provided path '%s', is valid. Are you sure you want to use it for storing directory scans? [Y/n]\n", newOutputDir)
			var confirmResp string
			fmt.Scanln(&confirmResp)
			fmt.Printf("|%s|\n", confirmResp)
			if len(confirmResp) != 0 && !strings.EqualFold(confirmResp, "y") {
				break
			}
		}
	}

	return newOutputDir, nil
}
