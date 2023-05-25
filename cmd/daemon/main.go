package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Fiye/diff"
	"github.com/Fiye/tree"
	"github.com/sevlyar/go-daemon"
)

func main() {
	cntxt := &daemon.Context{
		PidFileName: "dirt.pid",
		PidFilePerm: 0644,
		LogFileName: "dirt.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	p := "/"
	fmt.Printf("# Using path: %s\n", p)

	for _, sleepFor := range []int{0, 1, 3, 9, 27, 81, 243} {
		fmt.Printf("# Benchmark: %d minute(s) delay between tests\n", sleepFor)
		runTest(p, sleepFor, true)
		runTest(p, sleepFor, false)
	}
}

func runTest(path string, sleepFor int, isComprehensive bool) {
	var (
		s1 = tree.FileTree{}
		s2 = tree.FileTree{}
	)

	if isComprehensive {
		fmt.Println("Running COMPREHENSIVE scans")
	} else {
		fmt.Println("Running SHALLOW scans")
	}

	os.Remove("./s1.gob")
	os.Remove("./diff.gob")

	s1 = tree.WalkGenerateTree(path, 0, isComprehensive, nil)
	fmt.Printf("Total tree size is: %d, number of files in tree is: %d\n", s1.Size, s1.NumFilesTotal)
	fmt.Printf("  * Walk 1 took: %dms\n", s1.TimeTaken.Milliseconds())

	err := s1.WriteBinary("./s1.gob")
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Minute * time.Duration(sleepFor))

	s2 = tree.WalkGenerateTree(path, 0, isComprehensive, nil)
	fmt.Printf("  * Walk 2 took: %dms\n", s2.TimeTaken.Milliseconds())

	t := time.Now()
	dff := diff.CompareTrees(&s1, &s2)

	err = dff.WriteBinary("./diff.gob")
	if err != nil {
		panic(err)
	}

	fmt.Printf("  * Compare trees took: %dms\n\n", time.Since(t).Milliseconds())

	// fmt.Printf("  * Trees Diff (b - a) Summary:\n 	NewerPath: %s\n 	Number of Different Files: %d\n 	Number of Different Subtrees (non-recursive): %d\n 	Difference in 'Last Modified' Time: %dms\n 	Difference in Size: %d bytes\n	Difference in num files: %d\n	Difference in time taken: %dms\n\n", dff.NewerPath, len(dff.FilesDiff), len(dff.SubTreesDiff), dff.LastModifiedDiff.Milliseconds(), dff.SizeDiff, dff.NumFilesTotalDiff, dff.TimeTakenDiff.Milliseconds())
}

func printErrorStrings(errs []error) string {
	errStrings := make([]string, len(errs))
	for i, e := range errs {
		errStrings[i] = e.Error()
	}

	return strings.Join(errStrings, "\n")
}
