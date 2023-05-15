package diff

import (
	"time"
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
}

type FileDiff struct {
	NewerName        string
	HashDiff         string
	SizeDiff         int64
	LastModifiedDiff time.Duration
}
