package tree

import (
	"time"

	"github.com/Fiye/file"
)

/*
	FileTree: A recursive structure, representing the full filetree structure at a point in time
*/
type FileTree struct {
	// Non-recursive data
	Comprehensive bool
	BasePath      string
	Files         []file.File
	ErrStrings    []string
	LastVisited   time.Time
	TimeTaken     time.Duration
	Depth         int
	LastModified  time.Time

	// Recursive data
	SubTrees      []FileTree
	Size          int64
	NumFilesTotal int64

	AllHash       []byte // Only populated at depth == 0
	AllHashOffset int64
}
