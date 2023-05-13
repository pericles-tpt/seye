package tree

import "time"

/*
	Determine all the differences between two trees and store them in an output FileTree
*/
func CompareTrees(a, b *FileTree) FileTreeDiff {
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
			if !DiffEqual(res, FileTreeDiff{}) {
				ret.SubTreesDiff = append(ret.SubTreesDiff, res)
			}
		}

		ret.DiffCompleted = time.Now()
		return ret
	} else if b == nil {
		ret := FileTreeDiff{
			NewerPath:         a.BasePath,
			FilesDiff:         []File{},
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
			if !DiffEqual(res, FileTreeDiff{}) {
				ret.SubTreesDiff = append(ret.SubTreesDiff, res)
			}
		}

		ret.DiffCompleted = time.Now()
		return ret
	}

	retDiff := FileTreeDiff{}

	// 1. Basic checks
	// Ignored (highly variable)
	//	- LastVisited
	//  - TimeTaken
	//	- Priority
	simpleEqual := BasicTreesEqual(*a, *b)
	if simpleEqual {
		if a.BasePath == b.BasePath {
			return retDiff
		}
		retDiff.NewerPath = b.BasePath
		return retDiff
	}

	// 0. Need to determine how to compare files i.e. whether the trees are BOTH "Comprehensive"
	retDiff.Comprehensive = (*a).Comprehensive && (*b).Comprehensive

	// 2. Compare files at tree roots
	differentFiles := []File{}
	aFilesMap := map[string]*File{}
	for _, f := range a.Files {
		aFilesMap[f.Name] = &f
	}

	// 2a. Find all EXACT matching files
	notMatchedIndices := []int{}
	for i, f := range b.Files {
		aFile, ok := aFilesMap[f.Name]
		if !ok || (ok && !FilesEqual(f, *aFile, retDiff.Comprehensive)) {
			notMatchedIndices = append(notMatchedIndices, i)
		} else {
			aFilesMap[f.Name] = nil
		}
	}

	// 2b. Now look for renamed, changed or removed files
	for _, i := range notMatchedIndices {
		var (
			currFile          = b.Files[i]
			exactMatchFoundAt = ""
			sameNameFound     = (aFilesMap[currFile.Name] != nil)
		)

		for name, file := range aFilesMap {
			if file == nil {
				continue
			}

			if FilesEqual(b.Files[i], *file, retDiff.Comprehensive) {
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

	// 3. Compare subtrees
	differentTrees := []FileTreeDiff{}
	aFileTreesMap := map[string]*FileTree{}
	for _, f := range a.SubTrees {
		aFileTreesMap[f.BasePath] = &f
	}

	// 3a. Find all EXACT matching trees
	notMatchedTreeIndices := []int{}
	for i, f := range b.SubTrees {
		aTree, ok := aFileTreesMap[f.BasePath]
		if !ok || (ok && !BasicTreesEqual(f, *aTree)) {
			notMatchedTreeIndices = append(notMatchedTreeIndices, i)
		} else {
			aFileTreesMap[f.BasePath] = nil
		}
	}

	// 3b. Now look for renamed, changed or removed trees
	for _, i := range notMatchedTreeIndices {
		var (
			currTree          = b.SubTrees[i]
			exactMatchFoundAt = ""
			samePathFound     = (aFileTreesMap[currTree.BasePath] != nil)
		)

		for path, tree := range aFileTreesMap {
			if tree == nil {
				continue
			}

			if BasicTreesEqual(b.SubTrees[i], *tree) {
				exactMatchFoundAt = path
				continue
			}
		}

		if exactMatchFoundAt == "" && !samePathFound { // -> added
			res := CompareTrees(nil, &currTree)
			if !DiffEqual(res, FileTreeDiff{}) {
				differentTrees = append(differentTrees, res)
			}
		} else if exactMatchFoundAt != "" { // -> renamed
			res := CompareTrees(aFileTreesMap[exactMatchFoundAt], &currTree)
			if !DiffEqual(res, FileTreeDiff{}) {
				differentTrees = append(differentTrees, res)
			}
			aFileTreesMap[exactMatchFoundAt] = nil
		} else if samePathFound { // -> changed
			res := CompareTrees(aFileTreesMap[currTree.BasePath], &currTree)
			if !DiffEqual(res, FileTreeDiff{}) {
				differentTrees = append(differentTrees, res)
			}
			aFileTreesMap[exactMatchFoundAt] = nil
		}
	}

	// 3c. Anything left in aFileTreesMap has been REMOVED
	for _, tree := range aFileTreesMap {
		res := CompareTrees(tree, nil)
		if !DiffEqual(res, FileTreeDiff{}) {
			differentTrees = append(differentTrees, res)
		}

	}

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

func BasicTreesEqual(a, b FileTree) bool {
	return a.Depth == b.Depth &&
		len(a.Files) == len(b.Files) &&
		a.LastModified == b.LastModified &&
		a.Size == b.Size &&
		len(a.SubTrees) == len(b.SubTrees)
}

func FilesEqual(a, b File, isComprehensive bool) bool {
	if isComprehensive {
		return a.Size == b.Size && a.ByteSample == b.ByteSample
	} else {
		return a.Size == b.Size
	}
}

/*
	Checks everything except for basepath
*/
func TreesEqual(a, b FileTree) bool {
	isEqual := a.Depth == b.Depth
	isEqual = isEqual && (a.LastModified == b.LastModified)
	// Ignore `Priority`, `LastVisited` and `TimeTaken`, very sensitive
	isEqual = isEqual && (a.Size == b.Size)
	isEqual = isEqual && filesEqual(a.Files, b.Files)

	isEqual = len(a.SubTrees) == len(b.SubTrees)
	if !isEqual {
		return isEqual
	}
	for i, t := range a.SubTrees {
		isEqual = isEqual && TreesEqual(t, b.SubTrees[i])
	}
	return isEqual
}

func filesEqual(a, b []File) bool {
	if len(a) != len(b) {
		return false
	}

	isEqual := true
	for i, f := range a {
		isEqual = isEqual && (f == b[i])
	}
	return isEqual
}

/*
	NOTE: Not very thorough
*/
func DiffEqual(a, b FileTreeDiff) bool {
	return len(a.FilesDiff) == len(b.FilesDiff) &&
		a.LastModifiedDiff == b.LastModifiedDiff &&
		a.LastVisitedDiff == b.LastVisitedDiff &&
		a.NewerPath == b.NewerPath &&
		a.SizeDiff == b.SizeDiff &&
		len(a.SubTreesDiff) == len(b.SubTreesDiff) &&
		a.TimeTakenDiff == b.TimeTakenDiff
}
