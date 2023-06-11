package diff

import (
	"github.com/pericles-tpt/seye/tree"
)

/*
Compare two trees and store their differences in an output `ScanDiff`
*/
func CompareTrees(a, b *tree.FileTree) ScanDiff {
	if a == nil && b == nil {
		return ScanDiff{}
	}

	ret := ScanDiff{
		AllHash: []byte{},
		Trees:   map[string]TreeDiff{},
		Files:   map[string]FileDiff{},
	}
	if a == nil {
		_, ret = diffTrees([]tree.FileTree{}, []tree.FileTree{*b}, &([]byte{}), &b.AllHash, nil, false, &ret)
	} else if b == nil {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{}, &a.AllHash, &([]byte{}), nil, false, &ret)
	} else {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{*b}, &a.AllHash, &b.AllHash, nil, (*a).Comprehensive && (*b).Comprehensive, &ret)
	}

	return ret
}
