package scan

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Fiye/config"
	"github.com/Fiye/diff"
	"github.com/Fiye/tree"
	"github.com/Fiye/utility"
	"github.com/joomcode/errorx"
)

func GetScansFull(path string) *ScanRecords {
	ret, ok := scr.Scans[path]
	if !ok {
		return nil
	}
	return &ret
}

func GetAllScansFull() *map[string]ScanRecords {
	return &scr.Scans
}

func GetScansDiff(path string) *DiffRecords {
	ret, ok := scr.Diffs[path]
	if !ok {
		return nil
	}
	return &ret
}

func GetAllScansDiff() *map[string]DiffRecords {
	return &scr.Diffs
}

func (s *ScansRecord) IncrementCurrScanNum(path string) {
	curr, ok := scr.Scans[path]
	if ok {
		curr.CurrScanNum++
		scr.Scans[path] = curr
	}
	scr.Flush()
}

/*
Fiye only keeps the first and last full scans so this removes the second scan (if exists) then adds this new one
*/
func AddScansFull(t tree.FileTree) error {
	// Check for existing scans for this tree
	scanRootPath := t.BasePath
	existingScans, ok := scr.Scans[scanRootPath]
	if !ok {
		scr.Scans[scanRootPath] = ScanRecords{}
	} else if len(existingScans.Records) == 2 {
		os.Remove(config.GetScansOutputDir() + GetLastScanFilename(scanRootPath, false))
		existingScans.Records = existingScans.Records[:1]
		scr.Scans[scanRootPath] = existingScans
	}

	// Write the new tree to a file
	err := t.WriteBinary(config.GetScansOutputDir() + GetNewScanFilename(scanRootPath, false))
	if err != nil {
		return errorx.Decorate(err, "failed to write FileTree to local file")
	}

	// Create and append the new scan record
	newRecord := Record{
		t.Comprehensive,
		time.Now(),
	}
	tmp := scr.Scans[scanRootPath]
	tmp.Records = append(tmp.Records, newRecord)
	scr.Scans[scanRootPath] = tmp

	// Flush the changes to file
	err = scr.Flush()
	if err != nil {
		return errorx.Decorate(err, "failed to flush new `ScansRecord` data after adding new scan")
	}
	return nil
}

func RevertScansFull(t tree.FileTree) error {
	// Check for existing scans for this tree
	scanRootPath := t.BasePath
	existingScans, ok := scr.Scans[scanRootPath]
	if !ok {
		return nil
	}

	// Last scan completed in last 10s, assume it's the failed one, revert actions above
	if len(existingScans.Records) > 0 {
		lastScan := existingScans.Records[len(existingScans.Records)-1]
		if time.Since(lastScan.TimeCompleted) < time.Duration(10*time.Second) {
			os.Remove(config.GetScansOutputDir() + GetLastScanFilename(t.BasePath, false))
			existingScans.Records = existingScans.Records[:len(existingScans.Records)-1]
			scr.Scans[scanRootPath] = existingScans
		}
		return scr.Flush()
	}

	return nil
}

func AddScansDiff(rootPath string, isComprehensive bool, d diff.ScanDiff) error {
	// Write the new tree to a file
	err := d.WriteBinary(config.GetScansOutputDir() + GetNewScanFilename(rootPath, true))
	if err != nil {
		return errorx.Decorate(err, "failed to write ScanDiff to local file")
	}

	// Create and append the new scan record
	newRecord := Record{
		isComprehensive,
		time.Now(),
	}
	tmp := scr.Diffs[rootPath]
	tmp.Records = append(tmp.Records, newRecord)
	scr.Diffs[rootPath] = tmp

	// Flush the changes to file
	err = scr.Flush()
	if err != nil {
		return errorx.Decorate(err, "failed to flush new `ScansRecord` data after adding new scan")
	}

	return nil
}

func RevertScansDiff(rootPath string, isComprehensive bool, d diff.ScanDiff) error {
	// Check for existing scans for this tree
	existingDiffs, ok := scr.Diffs[rootPath]
	if !ok {
		return nil
	}

	// Last scan completed in last 10s, assume it's the failed one, revert actions above
	if len(existingDiffs.Records) > 0 {
		lastScan := existingDiffs.Records[len(existingDiffs.Records)-1]
		if time.Since(lastScan.TimeCompleted) < time.Duration(10*time.Second) {
			os.Remove(config.GetScansOutputDir() + GetLastScanFilename(rootPath, false))
			existingDiffs.Records = existingDiffs.Records[:len(existingDiffs.Records)-1]
			scr.Diffs[rootPath] = existingDiffs
		}
		return scr.Flush()
	}

	return nil
}

func GetLastScanFilename(rootPath string, isDiff bool) string {
	path := fmt.Sprintf("%s_", utility.HashFilePath(rootPath))
	lastNum := 0
	if isDiff {
		diffs, ok := scr.Diffs[rootPath]
		if ok {
			lastNum = len(diffs.Records) - 1
		}
		path += fmt.Sprintf("%d.diff", lastNum)
	} else {
		scans, ok := scr.Scans[rootPath]
		if ok {
			lastNum = scans.CurrScanNum - 1
		}
		path += fmt.Sprintf("%d.tree", lastNum)
	}
	return path
}

func GetNewScanFilename(rootPath string, isDiff bool) string {
	path := fmt.Sprintf("%s_", utility.HashFilePath(rootPath))
	newNum := 0
	if isDiff {
		existingDiffs, ok := scr.Diffs[rootPath]
		if ok {
			newNum = len(existingDiffs.Records)
		}
		path += fmt.Sprintf("%d.diff", newNum)
	} else {
		scans, ok := scr.Scans[rootPath]
		if ok {
			newNum = scans.CurrScanNum
		}
		path += fmt.Sprintf("%d.tree", newNum)
		scr.IncrementCurrScanNum(rootPath)
	}
	return path
}

func GetScanFilename(rootPath string, index int, isDiff bool) string {
	path := fmt.Sprintf("%s_", utility.HashFilePath(rootPath))
	if isDiff {
		path += fmt.Sprintf("%d.diff", index)
	} else {
		path += fmt.Sprintf("%d.tree", index)
	}
	return path
}

func AddDiffsForPath(path string, firstIdx, lastIdx int) (diff.ScanDiff, error) {
	ret := diff.ScanDiff{
		AllHash: []byte{},
		Trees:   map[string]diff.TreeDiff{},
		Files:   map[string]diff.FileDiff{},
	}

	for i := firstIdx; i <= lastIdx; i++ {
		// Load diff from disk
		diffPath := config.GetScansOutputDir() + GetScanFilename(path, i, true)
		fmt.Println("Reading diff with path: ", diffPath)
		diff, err := diff.ReadBinary(diffPath)
		if err != nil {
			return ret, errorx.Decorate(err, "failed to read diff from file at index %d", i)
		}

		fmt.Printf("Changes in diff %d for path '%s'\n", i, path)
		fmt.Println(len(diff.Files))
		fmt.Println(len(diff.Trees))

		fmt.Printf("Loaded diff with hash of size %d\n", len(diff.AllHash))

		// Add diff
		ret.AddDiff(diff, &ret.AllHash)
	}

	return ret, nil
}

func PrintLargestDiffs(limit int, sf diff.ScanDiff) {
	diffArray := make([]diff.TreeDiff, len(sf.Trees))
	i := 0
	for _, v := range sf.Trees {
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

	// Only keep deepest directories?
	deepestDirs := []diff.TreeDiff{}
	for _, v := range diffArray {
		for _, v2 := range diffArray {
			if v2.NewerPath != v.NewerPath && strings.HasPrefix(v2.NewerPath, v.NewerPath) {
				v.SizeDiff -= v2.SizeDiff
			}
		}
		deepestDirs = append(deepestDirs, v)
	}

	sort.SliceStable(deepestDirs, func(i, j int) bool {
		return deepestDirs[i].SizeDiff > deepestDirs[j].SizeDiff
	})

	fmt.Println("Biggest disk usage INCREASED")
	for i := 0; i < limit && i < len(deepestDirs); i++ {
		fmt.Printf("'%s' +%d bytes\n", deepestDirs[i].NewerPath, deepestDirs[i].SizeDiff)
	}

	fmt.Println("\nBiggest disk usage DECREASES")
	for i := 0; i < limit && i < len(deepestDirs); i++ {
		fmt.Printf("'%s' %d bytes\n", deepestDirs[len(deepestDirs)-1-i].NewerPath, deepestDirs[len(deepestDirs)-1-i].SizeDiff)
	}
}
