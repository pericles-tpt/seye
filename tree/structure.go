package tree

import "time"

type FileTree struct {
	BasePath    string
	Files       []File
	SubTrees    []FileTree
	Err         error
	LastVisited time.Time
	TimeTaken   time.Duration
	Depth       int

	Size         int64
	LastModified *time.Time
	Priority     int64
}

type FileTreeDiff struct {
	NewerPath        string
	FilesDiff        []File
	SubTreesDiff     []FileTreeDiff
	NewerErr         error
	LastVisitedDiff  time.Duration
	TimeTakenDiff    time.Duration
	DepthDiff        int
	SizeDiff         int64
	LastModifiedDiff time.Duration
	PriorityDiff     int64
}

type File struct {
	Name string
	Size int64
	Err  error
}

// type Snapshot struct {
// 	Dirs      []Dir
// 	FileCount int
// 	TotalSize int
// }

// type Dir struct {
// 	Path  string
// 	Files []File
// }

// type File struct {
// 	Name string
// 	Size *int64
// 	Err  error
// }
