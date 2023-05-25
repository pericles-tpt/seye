package diff

import (
	"github.com/Fiye/tree"
)

/*
	Determine all the differences between two trees and store them in an output Diff
*/
func CompareTrees(a, b *tree.FileTree) DiffMaps {
	if a == nil && b == nil {
		return DiffMaps{}
	}

	ret := DiffMaps{
		AllHash: []byte{},
		Trees:   map[string]TreeDiff{},
		Files:   map[string]FileDiff{},
	}
	if a == nil {
		_, ret = diffTrees([]tree.FileTree{}, []tree.FileTree{*b}, &([]byte{}), &b.AllHash, false, &ret)
	} else if b == nil {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{}, &a.AllHash, &([]byte{}), false, &ret)
	} else {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{*b}, &a.AllHash, &b.AllHash, (*a).Comprehensive && (*b).Comprehensive, &ret)
	}

	return ret
}

func TreeDiffEmpty(d DiffMaps) bool {
	return len(d.AllHash) == 0 &&
		len(d.Files) == 0 &&
		len(d.Trees) == 0
}
