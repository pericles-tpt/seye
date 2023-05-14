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
type FileTreeDiff struct {
	DiffCompleted time.Time
	Comprehensive bool

	// Non-recursive data
	NewerPath        string
	FilesDiff        []tree.File
	LastVisitedDiff  time.Duration
	TimeTakenDiff    time.Duration
	LastModifiedDiff time.Duration

	// Recursive data
	SubTreesDiff      []FileTreeDiff
	SizeDiff          int64
	NumFilesTotalDiff int64
}
