package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/Fiye/tree"
)

func main() {
	p := "/"
	fmt.Printf("# Using path: %s\n", p)
	for _, sleepFor := range []int{0, 1, 3, 9, 27, 81, 243} {
		fmt.Printf("# Benchmark: %d minute(s) delay between tests\n", sleepFor)

		s1 := &tree.FileTree{
			BasePath: p,
		}

		a := time.Now()
		_, size, numFiles, _ := s1.Walk(0, true)
		fmt.Printf("Some stats from the INITIAL, COMPLETE scan are:\n")
		// fmt.Printf("Errors are: %s\n", printErrorStrings(errs))
		fmt.Printf("Total tree size is: %d, number of files in tree is: %d\n", size, numFiles)
		fmt.Printf("  * Walk 1 took: %dms\n", time.Since(a).Milliseconds())

		time.Sleep(time.Minute * time.Duration(sleepFor))

		s2 := &tree.FileTree{
			BasePath: p,
		}

		a = time.Now()
		s2.Walk(0, true)
		fmt.Printf("  * Walk 2 took: %dms\n", time.Since(a).Milliseconds())

		// dt := tree.FileTree{}
		a = time.Now()
		diff := tree.CompareTrees(s1, s2)
		fmt.Printf("  * Compare trees took: %dms\n\n", time.Since(a).Milliseconds())

		s1 = nil
		s2 = nil

		fmt.Printf("  * Trees Diff (b - a) Summary:\n 	NewerPath: %s\n 	Number of Different Files: %d\n 	Number of Different Subtrees (non-recursive): %d\n 	Difference in 'Last Modified' Time: %dms\n 	Difference in Size: %d bytes\n	Difference in num files: %d\n	Difference in time taken: %dms\n\n", diff.NewerPath, len(diff.FilesDiff), len(diff.SubTreesDiff), diff.LastModifiedDiff.Milliseconds(), diff.SizeDiff, diff.NumFilesTotalDiff, diff.TimeTakenDiff.Milliseconds())
		// tree.PrintTree(&dt)

		// time.Sleep(time.Second * 30)

		// s2 := FileTree{
		// 	BasePath: "/home/",
		// }

		// Walk(&s2)

		// PrintTree(&s1)

		// CompareTrees(&s1, &s2, nil)
	}
}

func printErrorStrings(errs []error) string {
	errStrings := make([]string, len(errs))
	for i, e := range errs {
		errStrings[i] = e.Error()
	}

	return strings.Join(errStrings, "\n")
}
