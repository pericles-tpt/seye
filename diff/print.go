package diff

import (
	"fmt"
	"sort"
	"strings"
)

var (
	diffTypeToString = map[DiffType]string{
		modified: "modified",
		same:     "same",
		renamed:  "renamed",
		removed:  "removed",
		added:    "added",
	}
)

func PrintLargestDiffs(limit int, sf ScanDiff) {
	diffArray := make([]FileDiff, len(sf.Files))
	i := 0
	for _, v := range sf.Files {
		diffArray[i] = v
		i++
	}

	totalSizeIncrease := 0
	for _, v := range sf.Files {
		totalSizeIncrease += int(v.SizeDiff)
	}
	changeDirection := "increase"
	if totalSizeIncrease < 0 {
		changeDirection = "decrease"
	}
	fmt.Printf("Observed an overall %d byte %s to files\n\n", totalSizeIncrease, changeDirection)
	deepestDirs := diffArray

	sort.SliceStable(deepestDirs, func(i, j int) bool {
		return deepestDirs[i].SizeDiff > deepestDirs[j].SizeDiff
	})

	fmt.Println("Biggest disk usage INCREASES")
	for i := 0; i < limit && i < len(deepestDirs); i++ {
		if deepestDirs[i].SizeDiff > 0 {
			fmt.Printf("'%s' +%d bytes (%s)\n", deepestDirs[i].NewerName, deepestDirs[i].SizeDiff, strings.ToUpper(diffTypeToString[deepestDirs[i].Type]))
		}
	}

	fmt.Println("\nBiggest disk usage DECREASES")
	for i := 0; i < limit && i < len(deepestDirs); i++ {
		if deepestDirs[len(deepestDirs)-1-i].SizeDiff <= 0 {
			fmt.Printf("'%s' %d bytes (%s)\n", deepestDirs[len(deepestDirs)-1-i].NewerName, deepestDirs[len(deepestDirs)-1-i].SizeDiff, strings.ToUpper(diffTypeToString[deepestDirs[len(deepestDirs)-1-i].Type]))
		}
	}
}
