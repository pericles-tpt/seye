package tree

import (
	"path"
	"strings"
	"time"

	"github.com/Fiye/utility"
)

type BubbleUpProps struct {
	Size          int64
	NewestModtime time.Time
	NumFiles      int64
}

/*
Used with "Iterative" tree walk algorithms to "construct" the tree from leaves
to the root. It also allows properties to be "bubbled up" (like recursion), by
using maps and `BubbleUpProps`
*/
func constructTreeFromIterativeQ(buildQ *[]FileTree) FileTree {
	if len(*buildQ) == 0 {
		return FileTree{}
	}

	var (
		// `childTrees` uses to assign subtrees to their parent
		childTrees = map[string][]FileTree{}
		// `childProps` used to propogate recursive properties (e.g. size) from
		// a node to the root
		childProps   = map[string]BubbleUpProps{}
		highestDepth = (*buildQ)[len(*buildQ)-1].Depth
	)

	for len(*buildQ) > 0 {
		t := popBack1D(buildQ)

		// For "trees" above the deepest level, assign their subtrees from the
		// level below
		if t.Depth < highestDepth && childTrees[t.BasePath] != nil && len(childTrees[t.BasePath]) > 0 {
			t.SubTrees = childTrees[t.BasePath]
		}
		delete(childTrees, t.BasePath)

		// Update any properties specified in `BubbleUpProps` from the child(ren)
		// below
		var (
			bup BubbleUpProps
			ok  bool
		)
		t.SizeBelow = t.SizeDirect
		t.LastModifiedBelow = t.LastModifiedDirect
		t.NumFilesBelow = t.NumFilesDirect
		if bup, ok = childProps[t.BasePath]; ok {
			t.SizeBelow = t.SizeDirect + bup.Size
			t.LastModifiedBelow = utility.GetNewestTime(bup.NewestModtime, t.LastModifiedBelow)
			t.NumFilesBelow = t.NumFilesDirect + bup.NumFiles
		}
		delete(childProps, t.BasePath)

		// Sets properties in maps here to apply to the parent node
		if t.Depth > 0 {
			// Key for both maps, so parent nodes can acquire values specified here
			parentDir := path.Dir(strings.TrimSuffix(t.BasePath, "/"))

			// Create/Update an entry in `childProps` for this node's parent to update
			// its properties from its child
			if bup, ok = childProps[parentDir]; ok {
				childProps[parentDir] = BubbleUpProps{
					NewestModtime: t.LastModifiedBelow,
					Size:          t.SizeBelow,
					NumFiles:      t.NumFilesBelow,
				}
			} else {
				childProps[parentDir] = BubbleUpProps{
					NewestModtime: t.LastModifiedBelow,
					Size:          t.SizeBelow,
					NumFiles:      t.NumFilesBelow,
				}
			}

			// Add this tree to the list of `Subtrees` for the parent to get
			childTrees[parentDir] = prepend(childTrees[parentDir], t)
		} else {
			return t
		}
	}

	return FileTree{}
}

func prepend(trees []FileTree, new FileTree) []FileTree {
	return append([]FileTree{new}, trees...)
}
