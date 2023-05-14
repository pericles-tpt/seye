package diff

import (
	"bytes"
	"time"

	"github.com/Fiye/tree"
)

/*
	Determine all the differences between two trees and store them in an output tree.FileTree
*/
func CompareTrees(a, b *tree.FileTree) FileTreeDiff {
	if a == nil && b == nil {
		return FileTreeDiff{}
	}

	if a == nil {
		ret := FileTreeDiff{
			NewerPath:         b.BasePath,
			FilesDiff:         b.Files,
			SubTreesDiff:      []FileTreeDiff{},
			LastVisitedDiff:   b.LastVisited.Sub(time.Time{}),
			LastModifiedDiff:  time.Time{}.Sub(time.Time{}),
			TimeTakenDiff:     b.TimeTaken,
			SizeDiff:          b.Size,
			NumFilesTotalDiff: b.NumFilesTotal,
		}

		if b.LastModified != nil {
			ret.LastModifiedDiff = (*b.LastModified).Sub(time.Time{})
		}

		for _, t := range b.SubTrees {
			res := CompareTrees(nil, &t)
			if !DiffEqual(res, FileTreeDiff{}, false) {
				ret.SubTreesDiff = append(ret.SubTreesDiff, res)
			}
		}

		ret.DiffCompleted = time.Now()
		return ret
	} else if b == nil {
		ret := FileTreeDiff{
			NewerPath:         a.BasePath,
			FilesDiff:         []tree.File{},
			SubTreesDiff:      []FileTreeDiff{},
			LastVisitedDiff:   time.Time{}.Sub(a.LastVisited),
			TimeTakenDiff:     -a.TimeTaken,
			SizeDiff:          -a.Size,
			LastModifiedDiff:  time.Time{}.Sub(time.Time{}),
			NumFilesTotalDiff: -a.NumFilesTotal,
		}

		if a.LastModified != nil {
			ret.LastModifiedDiff = (*a.LastModified).Sub(time.Time{})
		}

		for _, t := range a.SubTrees {
			res := CompareTrees(&t, nil)
			if !DiffEqual(res, FileTreeDiff{}, false) {
				ret.SubTreesDiff = append(ret.SubTreesDiff, res)
			}
		}

		ret.DiffCompleted = time.Now()
		return ret
	}

	retDiff := FileTreeDiff{}

	// 0. Need to determine how to compare files i.e. whether the trees are BOTH "Comprehensive"
	retDiff.Comprehensive = (*a).Comprehensive && (*b).Comprehensive

	// TODO: Review file diff logic
	differentFiles := getDiffFiles((*a).Files, (*b).Files, retDiff.Comprehensive)

	// TODO: 3b and 3c below are probably wrong, causing large size of diff...
	// 3. Compare subtrees
	differentTrees := getDiffTrees((*a).SubTrees, (*b).SubTrees, retDiff.Comprehensive)

	retDiff.NewerPath = b.BasePath
	retDiff.FilesDiff = differentFiles
	retDiff.SubTreesDiff = differentTrees
	retDiff.LastVisitedDiff = (*b).LastVisited.Sub((*a).LastVisited)
	retDiff.TimeTakenDiff = b.TimeTaken - a.TimeTaken
	retDiff.SizeDiff = b.Size - a.Size
	retDiff.LastModifiedDiff = time.Time{}.Sub(time.Time{})
	retDiff.NumFilesTotalDiff = b.NumFilesTotal - a.NumFilesTotal

	alm := time.Time{}
	blm := time.Time{}
	if b.LastModified != nil {
		blm = *b.LastModified
	}
	if a.LastModified != nil {
		alm = *a.LastModified
	}
	retDiff.LastModifiedDiff = (blm).Sub(alm)
	retDiff.DiffCompleted = time.Now()

	return retDiff
}

func TreesEqual(a, b tree.FileTree, isComprehensive bool) bool {
	if isComprehensive {
		return DiffEqual(CompareTrees(&a, &b), FileTreeDiff{}, false)
	}
	return a.Depth == b.Depth &&
		len(a.Files) == len(b.Files) &&
		a.LastModified == b.LastModified &&
		a.Size == b.Size &&
		a.NumFilesTotal == b.NumFilesTotal &&
		len(a.SubTrees) == len(b.SubTrees) &&
		a.Comprehensive == b.Comprehensive
}

func FilesEqual(a, b tree.File, isComprehensive bool) bool {
	if isComprehensive {
		return a.Size == b.Size && hashesOrRawEqual(a.Hash, b.Hash)
	}
	return a.Size == b.Size
}

/*
	NOTE: Not very thorough
*/
func DiffEqual(a, b FileTreeDiff, isComprehensive bool) bool {
	return len(a.FilesDiff) == len(b.FilesDiff) &&
		a.LastModifiedDiff == b.LastModifiedDiff &&
		a.LastVisitedDiff == b.LastVisitedDiff &&
		a.NewerPath == b.NewerPath &&
		a.SizeDiff == b.SizeDiff &&
		len(a.SubTreesDiff) == len(b.SubTreesDiff) &&
		a.TimeTakenDiff == b.TimeTakenDiff
}

func hashesOrRawEqual(a, b *tree.Hash) bool {
	if a.Type != b.Type {
		return false
	}
	return (a == nil && b == nil) || a != nil && b != nil && bytes.Equal((a).Bytes[:], (b).Bytes[:])
}

func getDiffFiles(a, b []tree.File, isComprehensive bool) []tree.File {
	differentFiles := []tree.File{}

	aFilesMap := map[string]*tree.File{}
	for _, f := range a {
		aFilesMap[f.Name] = &f
	}

	// 2a. Find all EXACT matching files
	notMatchedIndices := []int{}
	for i, f := range b {
		aFile, ok := aFilesMap[f.Name]
		if !ok || !FilesEqual(f, *aFile, isComprehensive) {
			notMatchedIndices = append(notMatchedIndices, i)
		} else {
			aFilesMap[f.Name] = nil
		}
	}

	// 2b. Now look for renamed, changed or removed files
	for _, i := range notMatchedIndices {
		var (
			currFile          = b[i]
			exactMatchFoundAt = ""
			sameNameFound     = (aFilesMap[currFile.Name] != nil)
		)

		for name, file := range aFilesMap {
			if file == nil {
				continue
			}

			if FilesEqual(b[i], *file, isComprehensive) {
				exactMatchFoundAt = name
				continue
			}
		}

		if exactMatchFoundAt == "" && !sameNameFound { // -> added
			differentFiles = append(differentFiles, currFile)
		} else if exactMatchFoundAt != "" { // -> renamed
			currFile.Name = exactMatchFoundAt
			currFile.Size = 0
			differentFiles = append(differentFiles, currFile)
			aFilesMap[exactMatchFoundAt] = nil
		} else if sameNameFound { // -> changed
			currFile.Size -= aFilesMap[currFile.Name].Size
			differentFiles = append(differentFiles, currFile)
			aFilesMap[exactMatchFoundAt] = nil
		}
	}

	// 2c. Anything left in aFilesMap has been REMOVED
	for _, file := range aFilesMap {
		if file != nil {
			file.Size = -file.Size

			differentFiles = append(differentFiles, *file)
		}
	}

	return differentFiles
}

func getDiffTrees(a, b []tree.FileTree, isComprehensive bool) []FileTreeDiff {
	moreTrees := len(a)
	if len(b) > len(a) {
		moreTrees = len(b)
	}
	differentTrees := make([]FileTreeDiff, moreTrees)
	aFileTreesMap := map[string]*tree.FileTree{}
	for _, f := range a {
		aFileTreesMap[f.BasePath] = &f
	}

	// 3a. Find all EXACT matching trees
	notMatchedTreeIndices := []int{}
	for i, f := range b {
		aTree, ok := aFileTreesMap[f.BasePath]
		if !ok || (ok && !TreesEqual(f, *aTree, isComprehensive)) {
			notMatchedTreeIndices = append(notMatchedTreeIndices, i)
		} else {
			aFileTreesMap[f.BasePath] = nil
			differentTrees[i] = FileTreeDiff{}
		}
	}

	// 3b. Now look for renamed, changed or removed trees
	for _, i := range notMatchedTreeIndices {
		var (
			currTree          = b[i]
			exactMatchFoundAt = ""
			samePathFound     = (aFileTreesMap[currTree.BasePath] != nil)
		)

		for path, tree := range aFileTreesMap {
			if tree == nil {
				continue
			}

			if TreesEqual(b[i], *tree, isComprehensive) {
				exactMatchFoundAt = path
				break
			}
		}

		if exactMatchFoundAt == "" && !samePathFound { // -> added
			res := CompareTrees(nil, &currTree)
			differentTrees[i] = res
		} else if exactMatchFoundAt != "" { // -> renamed
			res := CompareTrees(aFileTreesMap[exactMatchFoundAt], &currTree)
			differentTrees[i] = res
			aFileTreesMap[exactMatchFoundAt] = nil
		} else if samePathFound { // -> changed
			res := CompareTrees(aFileTreesMap[currTree.BasePath], &currTree)
			differentTrees[i] = res
			aFileTreesMap[exactMatchFoundAt] = nil
		}
	}

	// 3c. Anything left in aFileTreesMap has been REMOVED
	for _, tree := range aFileTreesMap {
		res := CompareTrees(tree, nil)
		differentTrees = append(differentTrees, res)
	}

	return differentTrees
}
