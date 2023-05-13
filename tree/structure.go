package tree

import "time"

const (
	NumSampleBytes = 1000
)

/*
	Overview:
	* Initial scan:
		- The FS is traversed and its properties are recorded in a `FileTree` (on disk)
	* Subsequent "shallow" scans:
		- Same process as initial scan (temporarily, not on disk)
		- THEN a "diff" is conducted between the intial scan and the last scan to determine (ignore `FileByteSample`)
		  what changed, that diff is stored ON DISK
		- NOTE: To determine changes in files from the initial state we need to do (n-1) `FileTree`
				comparison e.g. Initial Full Scan + diff1 + diff2 + ... + diffn
	* "Comprehensive" scan:
		- Same process as initial scan (temporarily, not on disk)
		- Then diff the initial scan and this scan (READING `FileByteSample`)
		- NOTE: "Comprehensive" scans are done as requested by the user, they're more expensive
				than "shallow" scans and aren't saved to disk (likely use too much space?)
*/

/*
	FileTree: A recursive structure, representing the full filetree structure at a point in time
*/
type FileTree struct {
	// Non-recursive data
	Comprehensive bool
	BasePath      string
	Files         []File
	Err           []error
	LastVisited   time.Time
	TimeTaken     time.Duration
	Depth         int
	LastModified  *time.Time

	// Recursive data
	SubTrees      []FileTree
	Size          int64
	NumFilesTotal int64
}

/*
	FileTreeDiff: Conveys the change in a `FileTree` at a point in time

	NOTE: For `Comprehensive` scans the diff will also include information about
	whether the `ByteSample` of two files differs (useful for files with equal size
	but data differs)
*/
type FileTreeDiff struct {
	DiffCompleted time.Time
	Comprehensive bool

	// Non-recursive data
	NewerPath        string
	FilesDiff        []File
	LastVisitedDiff  time.Duration
	TimeTakenDiff    time.Duration
	LastModifiedDiff time.Duration

	// Recursive data
	SubTreesDiff      []FileTreeDiff
	SizeDiff          int64
	NumFilesTotalDiff int64
}

/*
	File: Encapsulates size information for a file in the FS

	NOTE: `ByteSample` is only populated for the INITAL and "Comprehensive"
	scans
*/
type File struct {
	Name       string
	ByteSample *ByteSample
	Size       int64
	Err        error
}

/*
	ByteSample: A sample of bytes of a file from a 	`MiddleOffset`

	MiddleOffset = ((float)FileSize / 2.0) -> Ceil() -> cast(int64)
*/
type ByteSample struct {
	MiddleOffset int64
	Bytes        []byte
}
