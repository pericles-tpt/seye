package records

import (
	"os"
	"time"

	"github.com/joomcode/errorx"
	"github.com/pericles-tpt/seye/config"
	"github.com/pericles-tpt/seye/diff"
	"github.com/pericles-tpt/seye/tree"
)

/*
Record that a new diff has been generated (i.e. .diff file)
*/
func AddFullScanRecord(t tree.FileTree) error {
	// Check for existing scans for this tree
	scanRootPath := t.BasePath
	existingScans, ok := recs.Scans[scanRootPath]
	if !ok {
		recs.Scans[scanRootPath] = ScanRecords{}
	} else if len(existingScans.Records) == 2 {
		os.Remove(config.GetScansOutputDir() + GetLastScanFilename(scanRootPath, false))
		existingScans.Records = existingScans.Records[:1]
		recs.Scans[scanRootPath] = existingScans
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
	tmp := recs.Scans[scanRootPath]
	tmp.Records = append(tmp.Records, newRecord)
	recs.Scans[scanRootPath] = tmp

	err = recs.Flush()
	if err != nil {
		return errorx.Decorate(err, "failed to flush new `ScansRecord` data after adding new scan")
	}
	return nil
}

/*
Revert the record of the LAST generated diff (i.e. .diff file)
*/
func RevertFullScanRecord(t tree.FileTree) error {
	// Check for existing scans for this tree
	scanRootPath := t.BasePath
	existingScans, ok := recs.Scans[scanRootPath]
	if !ok {
		return nil
	}

	// Last scan completed in last 10s, assume it's the failed one, revert actions above
	if len(existingScans.Records) > 0 {
		lastScan := existingScans.Records[len(existingScans.Records)-1]
		if time.Since(lastScan.TimeCompleted) < time.Duration(10*time.Second) {
			os.Remove(config.GetScansOutputDir() + GetLastScanFilename(t.BasePath, false))
			existingScans.Records = existingScans.Records[:len(existingScans.Records)-1]
			recs.Scans[scanRootPath] = existingScans
		}
		return recs.Flush()
	}

	return nil
}

/*
Record that a new scan has been generated (i.e. .tree file)
*/
func AddDiffScanRecord(rootPath string, isComprehensive bool, d diff.ScanDiff) error {
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
	tmp := recs.Diffs[rootPath]
	tmp.Records = append(tmp.Records, newRecord)
	recs.Diffs[rootPath] = tmp

	err = recs.Flush()
	if err != nil {
		return errorx.Decorate(err, "failed to flush new `ScansRecord` data after adding new scan")
	}

	return nil
}

/*
Revert the record of the LAST generated scan (i.e. .tree file)
*/
func RevertDiffScanRecord(rootPath string, isComprehensive bool, d diff.ScanDiff) error {
	// Check for existing scans for this tree
	existingDiffs, ok := recs.Diffs[rootPath]
	if !ok {
		return nil
	}

	// Last scan completed in last 10s, assume it's the failed one, revert actions above
	if len(existingDiffs.Records) > 0 {
		lastScan := existingDiffs.Records[len(existingDiffs.Records)-1]
		if time.Since(lastScan.TimeCompleted) < time.Duration(10*time.Second) {
			os.Remove(config.GetScansOutputDir() + GetLastScanFilename(rootPath, false))
			existingDiffs.Records = existingDiffs.Records[:len(existingDiffs.Records)-1]
			recs.Diffs[rootPath] = existingDiffs
		}
		return recs.Flush()
	}

	return nil
}
