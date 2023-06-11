package tree

import (
	"errors"
	"time"

	"github.com/Fiye/utility"
)

/*
A recursive structure, representing the full FileTree structure at a point in time
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

	LastModifiedDirect time.Time
	SizeDirect         int64
	NumFilesDirect     int64

	// Recursive data
	LastModifiedBelow time.Time
	SizeBelow         int64
	NumFilesBelow     int64
	SubTrees          []FileTree

	AllHash       []byte // Only populated at depth == 0
	AllHashOffset int64
}

/*
Encapsulates size information for a file in the FS
*/
type File struct {
	Name         string
	Hash         utility.HashLocation
	Size         int64
	Err          string
	LastModified time.Time
}

/*
Not a complete equality check, doesn't check hashed bytes of both files
*/
func (a *File) Equal(b File) bool {
	return a.Err == b.Err &&
		time.Time.Equal(a.LastModified, b.LastModified) &&
		a.Name == b.Name && a.Size == b.Size
}

func (a *FileTree) Equal(b FileTree) error {
	allFilesEqual := true
	if len(a.Files) != len(b.Files) {
		return errors.New("number of files in each tree aren't equal")
	}
	for i, f := range a.Files {
		allFilesEqual = allFilesEqual && f.Equal(b.Files[i])
	}

	allErrsEqual := true
	if len(a.ErrStrings) != len(b.ErrStrings) {
		return errors.New("number of errors in each tree aren't equal")
	}
	for i, e := range a.ErrStrings {
		allErrsEqual = allErrsEqual && (e == b.ErrStrings[i])
	}

	allSubTreesEqual := true
	if len(a.SubTrees) != len(b.SubTrees) {
		return errors.New("number of subtrees in each tree aren't equal")
	}
	for i, st := range a.SubTrees {
		allSubTreesEqual = allSubTreesEqual && (st.Equal(b.SubTrees[i]) == nil)
	}

	allHashEqual := true
	if len(a.AllHash) != len(b.AllHash) {
		return errors.New("size of hashes in each tree aren't equal")
	}
	for i, aByte := range a.AllHash {
		allHashEqual = allHashEqual && (aByte == b.AllHash[i])
	}

	if a.Comprehensive != b.Comprehensive {
		return errors.New("trees don't have the same `Comprehensive` value")
	}

	if a.BasePath != b.BasePath {
		return errors.New("trees don't have the same `BasePath` value")
	}

	if !allFilesEqual {
		return errors.New("trees don't have the same `Files` value")
	}

	if !allErrsEqual {
		return errors.New("trees don't have the same `ErrStrings`")
	}

	if a.Depth != b.Depth {
		return errors.New("trees don't have the same `Depth`")
	}

	if a.LastModifiedDirect != b.LastModifiedDirect {
		return errors.New("trees don't have the same `LastModifiedDirect`")
	}

	if a.SizeDirect != b.SizeDirect {
		return errors.New("trees don't have the same `SizeDirect`")
	}

	if a.NumFilesDirect != b.NumFilesDirect {
		return errors.New("trees don't have the same `NumFilesDirect`")
	}

	if a.LastModifiedBelow != b.LastModifiedBelow {
		return errors.New("trees don't have the same `LastModifiedBelow`")
	}

	if a.SizeBelow != b.SizeBelow {
		return errors.New("trees don't have the same `SizeBelow`")
	}

	if a.NumFilesBelow != b.NumFilesBelow {
		return errors.New("trees don't have the same `NumFilesBelow`")
	}

	if !allSubTreesEqual {
		return errors.New("trees don't have the same `SubTrees`")
	}

	if !allHashEqual {
		return errors.New("trees don't have the same AllHashes")
	}

	if a.AllHashOffset != b.AllHashOffset {
		return errors.New("trees don't have the same `AllHashOffset`")
	}

	return nil
	// Ignored:
	// - LastVisited
	// - TimeTaken
}
