package diff

import (
	"time"

	"github.com/Fiye/tree"
)

/*
	Adds a `TreeDiff` onto a `FileTree`, returning the resultant `FileTree`

	NOTE: Assumes the `FileTree` and `TreeDiff` have the same root
*/
func WalkAddDiff(t tree.FileTree, d TreeDiff) tree.FileTree {
	t.BasePath = d.NewerPath
	t.Comprehensive = d.Comprehensive
	if t.LastModified == nil && d.LastModifiedDiff != (time.Time{}).Sub(time.Time{}) {
		t.LastModified = &time.Time{}
		t.LastModified.Add(d.LastModifiedDiff)
	}
	t.LastVisited.Add(d.LastVisitedDiff)
	t.NumFilesTotal += d.NumFilesTotalDiff
	t.Size += d.SizeDiff
	t.TimeTaken += d.TimeTakenDiff

	t.Files = addFilesDiff(t.Files, d.FilesDiff)
	t.SubTrees = addTreesDiff(t.SubTrees, d.SubTreesDiff)

	return t
}

func addFilesDiff(files []tree.File, diffs []FileDiff) []tree.File {
	ret := []tree.File{}

	// Create a map of existing files
	fileMap := map[string]*tree.File{}
	for _, f := range files {
		fileMap[f.Name] = &f
	}

	/* Apply diffs:
	- Found in `fileMap` -> file was changed, add diff
	- Not found in `fileMap` -> file was removed
	*/
	for _, d := range diffs {
		f, ok := fileMap[d.NewerName]
		if ok {
			f.Hash = addHash(f.Hash, d.HashDiff)
			f.LastModified.Add(d.LastModifiedDiff)
			f.Size += d.SizeDiff
		} else if d.LastModifiedDiff.Seconds() < 0 { // -> Removed file
			// Do nothing
		} else {
			f := &tree.File{}
			f.LastModified = time.Time{}.Add(d.LastModifiedDiff)
			f.Name = d.NewerName
			f.Size = d.SizeDiff
			f.Hash = d.HashDiff
			f.Err = d.NewerErr
		}

		if f != nil {
			ret = append(ret, *f)
		}
	}

	return ret
}

func addTreesDiff(trees []tree.FileTree, diffs []TreeDiff) []tree.FileTree {
	ret := []tree.FileTree{}

	// Create a map of existing trees
	treeMap := map[string]*tree.FileTree{}
	for _, t := range trees {
		treeMap[t.BasePath] = &t
	}

	/* Apply diffs:
	- Found in `treeMap` -> tree was changed, add diff
	- Not found in `treeMap` -> tree was removed
	*/
	for _, d := range diffs {
		t, ok := treeMap[d.NewerPath]
		if ok {
			newTree := WalkAddDiff(*t, d)
			t = &newTree
		} else if d.TimeTakenDiff < 0 { // -> Removed file
			// Do nothing
		} else {
			t = &tree.FileTree{}
			newLastModified := time.Time{}.Add(d.LastModifiedDiff)
			t.BasePath = d.NewerPath
			t.Comprehensive = d.Comprehensive
			t.Depth = d.DepthDiff
			t.ErrStrings = d.ErrStringsDiff
			t.Files = addFilesDiff(make([]tree.File, len(d.FilesDiff)), d.FilesDiff)
			t.LastModified = &newLastModified
			t.LastVisited = time.Time{}.Add(d.LastVisitedDiff)
			t.NumFilesTotal = d.NumFilesTotalDiff
			t.Size = d.SizeDiff
			t.SubTrees = addTreesDiff(make([]tree.FileTree, len(d.SubTreesDiff)), d.SubTreesDiff)
			t.TimeTaken = d.TimeTakenDiff
		}

		if t != nil {
			ret = append(ret, *t)
		}
	}

	return ret
}

func addHash(a, b *tree.Hash) *tree.Hash {
	if (a != nil && b != nil) && (*a).Type == (*b).Type {
		for i, bByte := range (*b).Bytes {
			(*a).Bytes[i] += bByte
		}
	}
	return b
}
