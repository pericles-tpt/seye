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

// TODO: Probably not working
func (s *ScanDiff) AddDiff(new ScanDiff, thisAllHash *[]byte) {
	for k, v := range new.Files {
		existing, ok := s.Files[k]
		if !ok {
			s.Files[k] = v
		} else {
			existing.addDiff(&v, thisAllHash, &new.AllHash)
			s.Files[k] = existing
		}
	}

	for k, v := range new.Trees {
		existing, ok := s.Trees[k]
		if !ok {
			s.Trees[k] = v
		} else {
			existing.addDiff(&v)
			s.Trees[k] = existing
		}
	}
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

// TODO: Probably not working
func (t *TreeDiff) addDiff(new *TreeDiff) {
	if new.isEmpty() {
		return
	}

	switch new.Type {
	case changed:
		t.Comprehensive = new.Comprehensive
		t.NewerPath = new.NewerPath
		t.DepthDiff += new.DepthDiff
		t.ErrStringsDiff = append(t.ErrStringsDiff, new.ErrStringsDiff...)
		t.LastVisitedDiff += new.LastVisitedDiff
		t.LastModifiedDiff = new.LastModifiedDiff
		t.SizeDiff += new.SizeDiff
		t.NumFilesTotalDiff += new.NumFilesTotalDiff
		t.AllHashOffset = new.AllHashOffset
		t.SizeDiff += new.SizeDiff
	case renamed:
		t.NewerPath = new.NewerPath
	case removed:
		t.Type = new.Type
	default:
	}

	return
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

// TODO: Probably not working
func (f *FileDiff) addDiff(new *FileDiff, thisAllHash *[]byte, allHashNew *[]byte) {
	if new.isEmpty() {
		return
	}

	switch new.Type {
	case changed:
		f.NewerName = new.NewerName
		f.NewerErr = new.NewerErr
		f.LastModifiedDiff = new.LastModifiedDiff
		f.SizeDiff += new.SizeDiff
		new.HashDiff = file.InitialiseHashLocation(nil, nil, nil)
		if new.HashDiff.HashOffset > -1 {
			// Put the new hash in the old location for `f` in `allHash`
			f.HashDiff = addHashAtOffset(f.HashDiff.HashOffset, f.HashDiff.HashLength, f.HashDiff.Type, (*allHashNew)[new.HashDiff.HashOffset:new.HashDiff.HashOffset+new.HashDiff.HashLength], thisAllHash)
		}
	case renamed:
		f.NewerName = new.NewerName
	case removed:
		f.Type = new.Type
	default:
	}
}
