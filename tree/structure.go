package tree

import (
	"crypto/sha256"
	"time"
)

const (
	NumSampleBytes = 1000
)

type HashType int

const (
	SHA256 HashType = sha256.Size
	NONE   HashType = -1
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
	ErrStrings    []string
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
	File: Encapsulates size information for a file in the FS

	NOTE: `ByteSample` is only populated for the INITAL and "Comprehensive"
	scans
*/
type File struct {
	Name         string
	Hash         *Hash
	Size         int64
	Err          string
	LastModified time.Time
}

type Hash struct {
	Type  HashType
	Bytes [sha256.Size]byte
}
