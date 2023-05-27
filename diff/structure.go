package diff

import (
	"time"

	"github.com/Fiye/file"
)

type DiffType int64

const (
	changed DiffType = iota
	same
	renamed
	removed
	added
)

/*
The "root" ds for a scan diff, contains all detected
file/tree differences, their paths (as map keys)
and an array containing all new file hash values
*/
type ScanDiff struct {
	AllHash []byte // Only populated at depth == 0
	Trees   map[string]TreeDiff
	Files   map[string]FileDiff
}

func (s *ScanDiff) Empty() bool {
	return len(s.AllHash) == 0 &&
		len(s.Files) == 0 &&
		len(s.Trees) == 0
}

/*
Contains the differences between two `FileTree`s
*/
type TreeDiff struct {
	DiffCompleted time.Time
	Comprehensive bool
	Type          DiffType

	// Non-recursive data
	OriginalPath     string
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

	AllHash       []byte // Only populated at depth == 0
	AllHashOffset int64
}

func (t *TreeDiff) isEmpty() bool {
	empty := TreeDiff{}
	return t.Comprehensive == empty.Comprehensive &&
		len(t.AllHash) == 0 &&
		t.AllHashOffset == empty.AllHashOffset &&
		t.DepthDiff == empty.DepthDiff &&
		t.DiffCompleted == empty.DiffCompleted &&
		len(t.ErrStringsDiff) == 0 &&
		t.LastModifiedDiff == empty.LastModifiedDiff &&
		t.LastVisitedDiff == empty.LastVisitedDiff &&
		t.NewerPath == empty.NewerPath &&
		t.NumFilesTotalDiff == empty.NumFilesTotalDiff &&
		t.SizeDiff == empty.SizeDiff &&
		t.TimeTakenDiff == empty.TimeTakenDiff &&
		t.Type == empty.Type
}

/*
Contains the differences between two `File`s
*/
type FileDiff struct {
	NewerName        string
	NewerErr         string
	Type             DiffType
	HashDiff         file.HashLocation
	SizeDiff         int64
	LastModifiedDiff time.Duration
}

func (f *FileDiff) isEmpty() bool {
	empty := FileDiff{}
	return f.Type == empty.Type &&
		f.HashDiff == empty.HashDiff &&
		f.NewerErr == empty.NewerErr &&
		f.NewerName == empty.NewerName &&
		f.SizeDiff == empty.SizeDiff &&
		f.LastModifiedDiff == empty.LastModifiedDiff
}
