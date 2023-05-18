package diff

import (
	"time"

	"github.com/Fiye/tree"
)

/*
	FileTreeDiff: Conveys the change in a `FileTree` at a point in time

	NOTE: For `Comprehensive` scans the diff will also include information about
	whether the `ByteSample` of two files differs (useful for files with equal size
	but data differs)
*/
type TreeDiff struct {
	DiffCompleted time.Time
	Comprehensive bool

	// Non-recursive data
	NewerPath        string
	DepthDiff        int
	ErrStringsDiff   []string
	FilesDiff        []FileDiff
	FilesDiffIndices []int
	LastVisitedDiff  time.Duration
	TimeTakenDiff    time.Duration
	LastModifiedDiff time.Duration

	// Recursive data
	SubTreesDiff        []TreeDiff
	SubTreesDiffIndices []int
	SizeDiff            int64
	NumFilesTotalDiff   int64

	// TODO: Try improving Read/Write performance with hashes by having a contiguous AllHash structure
	// e.g. AllHash, only exists a Depth==0
	// AllHash *[]byte
}

type FileDiff struct {
	NewerName        string
	NewerErr         string
	HashDiff         *tree.Hash
	SizeDiff         int64
	LastModifiedDiff time.Duration

	// TODO: To work with AllHash (above), each file should know it's offset in AllHash and its hash length (from enum)
	// 		 `Hash`, should have type, length and offset.
	// 			* Type should be NONE, RAW, SHA256.
	//			* Length should be lookup from enum -> length
	//		    * Offset, is retrieved from a global offset, then incremented after each new byte section is added
}
