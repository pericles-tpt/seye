package tree

import (
	"os"
	"time"

	"github.com/joomcode/errorx"
	"github.com/pericles-tpt/seye/stats"
	"github.com/pericles-tpt/seye/utility"
)

type ReadLocation struct {
	WalkStats   *stats.WalkStats
	HashOffset  int
	HashLength  int
	FullPath    string
	Size        int64
	AllHashByte *[]byte
}

/*
Currently the slowest, although still correct, algorithm. Currently singlethreaded.

TODO: Rework for a MT approach, we'd likely need information about the `n`, roughly, equally most expensive, subtrees ahead of time so
we can split off from the main thread at those point. Then rejoin the subtrees from those threads with the main thread.
*/
func WalkGenerateTreeRecursive(path string, depth int, isComprehensive bool, walkStats *stats.WalkStats) (tree *FileTree) {
	AllHashBytes := []byte{}
	if depth == 0 {
		filesAddedForHashing = 0
	}

	tree = &FileTree{BasePath: path}

	ents, err := os.ReadDir(path)
	if err != nil {
		tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to open `tree.BasePath`").Error())
		if depth == 0 {
			return tree
		}
	}

	// This function is ST currently, these are initialised here for MT properties elsewhere
	threadsCopyBuffer = make([][]byte, 1)
	threadsBytesRead = make([]int64, 1)
	threadsFilesRead = make([]int64, 1)
	threadsCopyBuffer[0] = make([]byte, copyBufferLen)

	var (
		startWalk = time.Now()
	)
	for _, e := range ents {
		fullPath := getFullPath(path, e.Name())

		if e.IsDir() {
			if utility.Contains(ignoredDirs, e.Name()) {
				continue
			}

			subTree := WalkGenerateTreeRecursive(fullPath, depth+1, isComprehensive, walkStats)
			if len(subTree.ErrStrings) > 0 {
				tree.ErrStrings = append(tree.ErrStrings, subTree.ErrStrings...)
			}
			tree.SizeBelow += subTree.SizeBelow
			tree.NumFilesBelow += subTree.NumFilesBelow

			tree.SubTrees = append(tree.SubTrees, *subTree)
			tree.LastModifiedBelow = utility.GetNewestTime(tree.LastModifiedBelow, subTree.LastModifiedBelow)
		} else if e.Type().IsRegular() {
			fStat, err := e.Info()
			nf := File{
				Name: fullPath,
				Hash: utility.InitialiseHashLocation(),
				Size: fStat.Size(),
				// TODO: Err is never populated atm...
				LastModified: fStat.ModTime(),
			}
			if err != nil {
				tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to stat file").Error())
			} else {
				if isComprehensive {
					oldAllHashByteLen := len(AllHashBytes)
					AllHashBytes = append(AllHashBytes, make([]byte, chosenHash)...)
					var errStrings []string
					nf.Hash, errStrings = readHashFile(ReadLocation{
						WalkStats:   walkStats,
						FullPath:    fullPath,
						HashOffset:  oldAllHashByteLen,
						HashLength:  chosenHash,
						Size:        nf.Size,
						AllHashByte: &AllHashBytes,
					}, 0)

					if len(errStrings) > 0 {
						tree.ErrStrings = append(tree.ErrStrings, errStrings...)
					}
				}

				if walkStats != nil {
					walkStats.UpdateLargestFiles(stats.BasicFile{Path: fullPath, Size: fStat.Size()})
				}

				tree.LastModifiedDirect = utility.GetNewestTime(tree.LastModifiedDirect, nf.LastModified)
				tree.LastModifiedBelow = utility.GetNewestTime(tree.LastModifiedBelow, nf.LastModified)

				tree.SizeDirect += nf.Size
			}

			tree.Files = append(tree.Files, nf)
		}
	}
	tree.TimeTaken = time.Since(startWalk)

	tree.BasePath = path
	tree.Comprehensive = isComprehensive

	tree.NumFilesDirect = int64(len(tree.Files))
	tree.NumFilesBelow += tree.NumFilesDirect
	tree.SizeBelow += tree.SizeDirect

	tree.LastVisited = time.Now()
	tree.Depth = depth

	if tree.Depth == 0 && len(AllHashBytes) > 0 {
		tree.AllHash = AllHashBytes
	}

	return tree
}
