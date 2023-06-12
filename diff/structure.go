package diff

import (
	"time"

	"github.com/pericles-tpt/seye/utility"
)

type DiffType int64

const (
	modified DiffType = iota
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

func (a *ScanDiff) Equals(b ScanDiff) bool {
	// Hashes could be added in any order, we can only check the length here
	if len(a.AllHash) != len(b.AllHash) {
		return false
	}

	if len(a.Files) != len(b.Files) {
		return false
	}
	aKeys := make([]string, len(a.Files))
	i := 0
	for k, _ := range a.Files {
		aKeys[i] = k
	}
	allFileDiffsEqual := true
	for _, k := range aKeys {
		if v, ok := b.Files[k]; ok {
			allFileDiffsEqual = allFileDiffsEqual && v.Equals(a.Files[k])
			delete(b.Files, k)
		} else {
			return false
		}
	}

	if len(a.Trees) != len(b.Trees) {
		return false
	}
	aKeys = make([]string, len(a.Trees))
	i = 0
	for k, _ := range a.Trees {
		aKeys[i] = k
	}
	allTreeDiffsEqual := true
	for _, k := range aKeys {
		if v, ok := b.Trees[k]; ok {
			allTreeDiffsEqual = allTreeDiffsEqual && v.Equals(a.Trees[k])
			delete(b.Trees, k)
		} else {
			return false
		}
	}

	return allFileDiffsEqual && allTreeDiffsEqual
}

/*
Adds the changes in `new` to diff `s`. So that we can do, for example:

	s1 + diff(s1, s2) + diff(s2, s3) == s3

TODO: Currently not working
*/
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
Captures the differences between two `FileTree`s
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

	LastModifiedDiffDirect  time.Duration
	SizeDiffDirect          int64
	NumFilesTotalDiffDirect int64

	// Recursive data
	SubTreesDiff        []TreeDiff
	SubTreesDiffIndices []int

	AllHash       []byte // Only populated at depth == 0
	AllHashOffset int64
}

func (t *TreeDiff) Equals(b TreeDiff) bool {
	if len(t.ErrStringsDiff) != len(b.ErrStringsDiff) {
		return false
	}
	allErrStringsEqual := true
	for i, e := range t.ErrStringsDiff {
		allErrStringsEqual = allErrStringsEqual && (e == b.ErrStringsDiff[i])
	}

	if len(t.FilesDiff) != len(b.FilesDiff) {
		return false
	}
	allFilesDiffEqual := true
	for i, f := range t.FilesDiff {
		allFilesDiffEqual = allFilesDiffEqual && f.Equals(b.FilesDiff[i])
	}

	if len(t.FilesDiffIndices) != len(b.FilesDiffIndices) {
		return false
	}
	allFilesDiffIndicesEqual := true
	for i, fdi := range t.FilesDiffIndices {
		allFilesDiffIndicesEqual = allFilesDiffIndicesEqual && (fdi == b.FilesDiffIndices[i])
	}

	if len(t.SubTreesDiff) != len(b.SubTreesDiff) {
		return false
	}
	allSubTreesDiffEqual := true
	for i, std := range t.SubTreesDiff {
		allSubTreesDiffEqual = allSubTreesDiffEqual && std.Equals(b.SubTreesDiff[i])
	}

	if len(t.SubTreesDiffIndices) != len(b.SubTreesDiffIndices) {
		return false
	}
	allSubTreesDiffIndicesEqual := true
	for i, stdi := range t.SubTreesDiffIndices {
		allSubTreesDiffIndicesEqual = allSubTreesDiffIndicesEqual && stdi == b.SubTreesDiffIndices[i]
	}

	return t.Comprehensive == b.Comprehensive &&
		t.Type == b.Type &&
		t.OriginalPath == b.OriginalPath &&
		t.NewerPath == b.NewerPath &&
		t.DepthDiff == b.DepthDiff &&
		t.LastVisitedDiff == b.LastVisitedDiff &&
		t.TimeTakenDiff == b.TimeTakenDiff &&
		t.LastModifiedDiffDirect == b.LastModifiedDiffDirect &&
		t.SizeDiffDirect == b.SizeDiffDirect &&
		t.NumFilesTotalDiffDirect == b.NumFilesTotalDiffDirect &&
		t.AllHashOffset == b.AllHashOffset &&
		allErrStringsEqual &&
		allFilesDiffEqual &&
		allFilesDiffIndicesEqual &&
		allSubTreesDiffEqual &&
		allSubTreesDiffIndicesEqual

}

func (t *TreeDiff) Empty() bool {
	empty := TreeDiff{}
	return t.Comprehensive == empty.Comprehensive &&
		len(t.AllHash) == 0 &&
		t.AllHashOffset == empty.AllHashOffset &&
		t.DepthDiff == empty.DepthDiff &&
		time.Time.Equal(t.DiffCompleted, empty.DiffCompleted) &&
		len(t.ErrStringsDiff) == 0 &&
		t.LastModifiedDiffDirect == empty.LastModifiedDiffDirect &&
		t.LastVisitedDiff == empty.LastVisitedDiff &&
		t.NewerPath == empty.NewerPath &&
		t.NumFilesTotalDiffDirect == empty.NumFilesTotalDiffDirect &&
		t.SizeDiffDirect == empty.SizeDiffDirect &&
		t.TimeTakenDiff == empty.TimeTakenDiff &&
		t.Type == empty.Type
}

/*
Utility function used by `AddDiff` to achieve `t += new`

TODO: Currently not working
*/
func (t *TreeDiff) addDiff(new *TreeDiff) {
	if new.Empty() {
		return
	}

	switch new.Type {
	case modified:
		t.Comprehensive = new.Comprehensive
		t.NewerPath = new.NewerPath
		t.DepthDiff += new.DepthDiff
		t.ErrStringsDiff = append(t.ErrStringsDiff, new.ErrStringsDiff...)
		t.LastVisitedDiff += new.LastVisitedDiff
		t.LastModifiedDiffDirect = new.LastModifiedDiffDirect
		t.SizeDiffDirect += new.SizeDiffDirect
		t.NumFilesTotalDiffDirect += new.NumFilesTotalDiffDirect
		t.AllHashOffset = new.AllHashOffset
		t.SizeDiffDirect += new.SizeDiffDirect
	case renamed:
		t.NewerPath = new.NewerPath
	case removed:
		t.Type = new.Type
	default:
	}
}

/*
Contains the differences between two `File`s
*/
type FileDiff struct {
	NewerName        string
	NewerErr         string
	Type             DiffType
	HashDiff         utility.HashLocation
	SizeDiff         int64
	LastModifiedDiff time.Duration
}

func (f *FileDiff) Empty() bool {
	empty := FileDiff{}
	return f.Type == empty.Type &&
		f.HashDiff == empty.HashDiff &&
		f.NewerErr == empty.NewerErr &&
		f.NewerName == empty.NewerName &&
		f.SizeDiff == empty.SizeDiff &&
		f.LastModifiedDiff == empty.LastModifiedDiff
}

func (f *FileDiff) Equals(b FileDiff) bool {
	return f.HashDiff.HashLength == b.HashDiff.HashLength &&
		f.HashDiff.Type == b.HashDiff.Type &&
		f.LastModifiedDiff == b.LastModifiedDiff &&
		f.NewerErr == b.NewerErr &&
		f.NewerName == b.NewerName &&
		f.SizeDiff == b.SizeDiff &&
		f.Type == b.Type
}

/*
Utility function used by `AddDiff` to achieve `f += new`

TODO: Currently not working
*/
func (f *FileDiff) addDiff(new *FileDiff, thisAllHash *[]byte, allHashNew *[]byte) {
	if new.Empty() {
		return
	}

	switch new.Type {
	case modified:
		f.NewerName = new.NewerName
		f.NewerErr = new.NewerErr
		f.LastModifiedDiff = new.LastModifiedDiff
		f.SizeDiff += new.SizeDiff
		new.HashDiff = utility.InitialiseHashLocation()
		if new.HashDiff.HashOffset > -1 {
			// Put the new hash in the old location for `f` in `allHash`
			f.HashDiff = utility.AddHashAtOffset(f.HashDiff.HashOffset, f.HashDiff.HashLength, f.HashDiff.Type, (*allHashNew)[new.HashDiff.HashOffset:new.HashDiff.HashOffset+new.HashDiff.HashLength], thisAllHash)
		}
	case renamed:
		f.NewerName = new.NewerName
	case removed:
		f.Type = new.Type
	default:
	}
}
