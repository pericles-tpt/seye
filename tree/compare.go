package tree

import "fmt"

func CompareTrees(a, b, diffTree *FileTree) {
	if a == nil {
		if b != nil {
			diffTree = b
			return
		}
		diffTree = nil
		return
	} else if b == nil {
		if a != nil {
			diffTree = a
			return
		}
		diffTree = nil
		return
	}

	if a.BasePath != b.BasePath {
		if diffTree == nil {
			*diffTree = FileTree{BasePath: b.BasePath}
			fmt.Printf("Found trees with different names: %s and %s\n", a.BasePath, b.BasePath)
		}
	}

	differentFiles := []File{}
	var filesSize int64
	for i, f := range b.Files {
		if f != a.Files[i] {
			sizeDiff := f.Size - (a.Files[i].Size)
			filesSize += sizeDiff
			differentFiles = append(differentFiles, File{f.Name, sizeDiff, f.Err})
			fmt.Printf("Found different files, names: %s and %s, size: %d and %d\n", a.Files[i].Name, f.Name, a.Files[i].Size, f.Size)
		}
	}

	differentTrees := []FileTree{}
	for i, t := range b.SubTrees {
		var diffTree *FileTree = nil
		CompareTrees(&a.SubTrees[i], &t, diffTree)
		if diffTree != nil {
			differentTrees = append(differentTrees, *diffTree)
		}
	}

	if len(differentFiles) > 0 {
		diffTree.Files = differentFiles
	}
	if len(differentTrees) > 0 {
		diffTree.SubTrees = differentTrees
	}
}
