package diff

import (
	"bytes"

	"github.com/Fiye/file"
	"github.com/Fiye/tree"
)

/*
Determine all the differences between two trees and store them in an output Diff
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
		_, ret = diffTrees([]tree.FileTree{}, []tree.FileTree{*b}, &([]byte{}), &b.AllHash, false, &ret)
	} else if b == nil {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{}, &a.AllHash, &([]byte{}), false, &ret)
	} else {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{*b}, &a.AllHash, &b.AllHash, (*a).Comprehensive && (*b).Comprehensive, &ret)
	}

	return ret
}

func hashesEqual(a, b file.HashLocation, allHashesA, allHashesB *[]byte) bool {
	if a.HashOffset == -1 && b.HashOffset == -1 {
		return true
	} else if a.HashOffset > -1 && b.HashOffset > -1 {
		bytesA := (*allHashesA)[a.HashOffset : a.HashOffset+a.HashLength]
		bytesB := (*allHashesB)[b.HashOffset : b.HashOffset+b.HashLength]
		return (a.Type == b.Type) && bytes.Equal(bytesA, bytesB)
	}

	// -> one has a hash, the other doesn't -> the file has changed
	return false
}
