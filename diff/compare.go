package diff

import (
	"time"

	"github.com/Fiye/tree"
)

/*
	Determine all the differences between two trees and store them in an output tree.FileTree
*/
func CompareTrees(a, b *tree.FileTree) TreeDiff {
	if a == nil && b == nil {
		return TreeDiff{}
	}

	ret := []TreeDiff{}
	if a == nil {
		_, ret = diffTrees([]tree.FileTree{}, []tree.FileTree{*b}, false)
	} else if b == nil {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{}, false)
	} else {
		_, ret = diffTrees([]tree.FileTree{*a}, []tree.FileTree{*b}, (*a).Comprehensive && (*b).Comprehensive)
	}

	return ret[0]
}

func TreeDiffEmpty(d TreeDiff) bool {
	zeroDuration := time.Time{}.Sub(time.Time{})

	return d.DiffCompleted == time.Time{} &&
		!d.Comprehensive &&
		d.NewerPath == "" &&
		d.DepthDiff == 0 &&
		d.ErrStringsDiff == nil &&
		d.FilesDiff == nil &&
		d.FilesDiffIndices == nil &&
		d.LastVisitedDiff == zeroDuration &&
		d.TimeTakenDiff == zeroDuration &&
		d.LastModifiedDiff == zeroDuration &&
		d.SubTreesDiff == nil &&
		d.SubTreesDiffIndices == nil &&
		d.SizeDiff == 0 &&
		d.NumFilesTotalDiff == 0
}
