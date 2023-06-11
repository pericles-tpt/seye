package records

import (
	"fmt"

	"github.com/pericles-tpt/seye/utility"
)

func GetScansFull(path string) *ScanRecords {
	ret, ok := recs.Scans[path]
	if !ok {
		return nil
	}
	return &ret
}

func GetAllScansFull() *map[string]ScanRecords {
	return &recs.Scans
}

func GetScansDiff(path string) *DiffRecords {
	ret, ok := recs.Diffs[path]
	if !ok {
		return nil
	}
	return &ret
}

func GetAllScansDiff() *map[string]DiffRecords {
	return &recs.Diffs
}

/*
Get the filename of the LAST completed scan (for either a 'diff' or 'full' scan)
*/
func GetLastScanFilename(rootPath string, isDiff bool) string {
	path := fmt.Sprintf("%s_", utility.HashFilePath(rootPath))
	lastNum := 0
	if isDiff {
		diffs, ok := recs.Diffs[rootPath]
		if ok {
			lastNum = len(diffs.Records) - 1
		}
		path += fmt.Sprintf("%d.diff", lastNum)
	} else {
		scans, ok := recs.Scans[rootPath]
		if ok {
			lastNum = scans.CurrScanNum - 1
		}
		path += fmt.Sprintf("%d.tree", lastNum)
	}
	return path
}

/*
Get the filename for the NEXT scan (for either a 'diff' or 'full' scan)
*/
func GetNewScanFilename(rootPath string, isDiff bool) string {
	path := fmt.Sprintf("%s_", utility.HashFilePath(rootPath))
	newNum := 0
	if isDiff {
		existingDiffs, ok := recs.Diffs[rootPath]
		if ok {
			newNum = len(existingDiffs.Records)
		}
		path += fmt.Sprintf("%d.diff", newNum)
	} else {
		scans, ok := recs.Scans[rootPath]
		if ok {
			newNum = scans.CurrScanNum
		}
		path += fmt.Sprintf("%d.tree", newNum)
		recs.incrementCurrScanNum(rootPath)
	}
	return path
}

/*
Get the filename for a scan at an index (for either a 'diff' or 'full' scan)
*/
func GetScanFilename(rootPath string, index int, isDiff bool) string {
	path := fmt.Sprintf("%s_", utility.HashFilePath(rootPath))
	if isDiff {
		path += fmt.Sprintf("%d.diff", index)
	} else {
		path += fmt.Sprintf("%d.tree", index)
	}
	return path
}

func (s *AllRecords) incrementCurrScanNum(path string) {
	curr, ok := recs.Scans[path]
	if ok {
		curr.CurrScanNum++
		recs.Scans[path] = curr
	}
	recs.Flush()
}
