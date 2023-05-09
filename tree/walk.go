package tree

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	Here         = time.Local
	UnixYear2048 = time.Date(2025, 1, 1, 0, 0, 0, 0, Here).Unix()
	AllTrees     = []FileTree{}
)

func (tree *FileTree) Walk(depth int) (size int64, latestModified *time.Time) {

	ents, err := os.ReadDir(tree.BasePath)
	tree.Err = err

	var (
		newestModified *time.Time = nil
		startWalk                 = time.Now()
	)
	for _, e := range ents {
		if e.IsDir() {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			subTree := FileTree{
				BasePath: fmt.Sprintf("%s/%s", tree.BasePath, e.Name()),
			}
			subSize, currModified := subTree.Walk(depth + 1)
			tree.Size += subSize
			tree.SubTrees = append(tree.SubTrees, subTree)

			newestModified = GetNewestTime(newestModified, currModified)
		} else if e.Type().IsRegular() {
			f, err := e.Info()
			if err != nil {
				tree.Files = append(tree.Files, File{"", 0, err})
			} else {
				tree.Files = append(tree.Files, File{f.Name(), f.Size(), err})
				tree.Size += f.Size()

				currModified := f.ModTime()
				newestModified = GetNewestTime(newestModified, &currModified)
			}
		}
	}
	tree.TimeTaken = time.Since(startWalk)

	tree.LastModified = newestModified
	tree.LastVisited = time.Now()
	tree.Depth = depth
	tree.Priority = CalculatePriority(*tree)

	AllTrees = append(AllTrees, *tree)

	return tree.Size, newestModified
}

func GetNewestTime(curr *time.Time, new *time.Time) *time.Time {
	if curr == nil || (new != nil && curr.Before(*new)) {
		return new
	}
	return curr
}

// TODO: Delete this?
func PrintTree(t *FileTree) {
	if t == nil {
		return
	}
	fmt.Printf("On tree with path: %s, size at this level is: %d\n", t.BasePath, t.Size)

	for _, f := range t.Files {
		fmt.Printf("	Has file %s of size: %d\n", f.Name, f.Size)
	}

	for _, t := range t.SubTrees {
		PrintTree(&t)
	}
}

func CalculatePriority(t FileTree) int64 {
	// FilesDiff?
	// DirsDiff?

	// older -> good
	// lastVisited := 10 * float64((UnixYear2048-t.LastVisited.Unix())/UnixYear2048)
	// // newer -> good
	// var lastModified float64 = 0
	// if t.LastModified != nil {
	// 	lastModified = 0 //100 * float64(UnixYear2048/(UnixYear2048-t.LastModified.Unix()))
	// }
	// shorter -> good

	// larger -> good
	var size int64 = 1
	var timeTaken float64
	if t.Size > 0 {
		size = t.Size
		timeTaken = 10 * (float64(time.Hour.Nanoseconds()) / float64(t.TimeTaken.Nanoseconds()))
	}
	// deeper -> good
	deeper := t.Depth

	return int64(float64(size*int64(deeper)) * (float64(1) / float64(timeTaken))) // * int64((float64(lastVisited) + float64(lastModified))
}
