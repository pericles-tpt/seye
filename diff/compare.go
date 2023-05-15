package diff

import (
	"github.com/Fiye/tree"
)

/*
	Determine all the differences between two trees and store them in an output tree.FileTree
*/
func CompareTrees(a, b *tree.FileTree) TreeDiff {
	if a == nil && b == nil {
		return TreeDiff{}
	}

	if a == nil {
		_, diff := diffTrees([]tree.FileTree{}, []tree.FileTree{*b}, false)
		return diff[0]
	} else if b == nil {
		_, diff := diffTrees([]tree.FileTree{*a}, []tree.FileTree{}, false)
		return diff[0]
	}

	_, diff := diffTrees([]tree.FileTree{*a}, []tree.FileTree{*b}, (*a).Comprehensive && (*b).Comprehensive)
	return diff[0]
}
