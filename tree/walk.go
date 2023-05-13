package tree

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/joomcode/errorx"
)

var (
	Here         = time.Local
	UnixYear2048 = time.Date(2025, 1, 1, 0, 0, 0, 0, Here).Unix()
	AllTrees     = []FileTree{}
)

func (tree *FileTree) Walk(depth int, isComprehensive bool) (errsBelow []error, size int64, numFilesR int64, latestModified *time.Time) {
	tree.Comprehensive = isComprehensive

	ents, err := os.ReadDir(tree.BasePath)
	tree.Err = []error{}
	if err != nil {
		tree.Err = append(tree.Err, errorx.Decorate(err, "failed to open `tree.BasePath`"))
	}

	var (
		newestModified     *time.Time = nil
		startWalk                     = time.Now()
		numFilesInSubtrees int64      = 0
	)
	for _, e := range ents {
		fullPath := fmt.Sprintf("%s/%s", tree.BasePath, e.Name())

		if e.IsDir() {
			subTree := FileTree{
				BasePath: fullPath,
			}
			err, subSize, subNumFiles, currModified := subTree.Walk(depth+1, isComprehensive)
			if len(err) > 0 {
				tree.Err = append(tree.Err, err...)
			}
			tree.Size += subSize
			numFilesInSubtrees += subNumFiles

			tree.SubTrees = append(tree.SubTrees, subTree)
			newestModified = GetNewestTime(newestModified, currModified)
		} else if e.Type().IsRegular() {
			fStat, err := e.Info()
			nf := File{
				Err:        err,
				ByteSample: nil,
			}
			if err != nil {
				tree.Err = append(tree.Err, errorx.Decorate(err, "failed to stat file"))
			} else {
				var sample *ByteSample = nil

				if isComprehensive {
					fTmp, err := os.OpenFile(fullPath, os.O_RDONLY, 0400)
					if err != nil {
						tree.Err = append(tree.Err, errorx.Decorate(err, "failed to open file to get `ByteSample` for 'Comprehensive' scan"))
					} else if fStat.Size() > 0 {
						sample = &ByteSample{
							MiddleOffset: int64(math.Ceil(float64(fStat.Size()) / 2.0)),
							Bytes:        make([]byte, 1000),
						}

						if fStat.Size() <= NumSampleBytes {
							_, err = fTmp.Read(sample.Bytes)
							if err != nil {
								tree.Err = append(tree.Err, errorx.Decorate(err, "failed to read WHOLE file"))
							}
						} else {
							n, err := fTmp.ReadAt(sample.Bytes, sample.MiddleOffset-(NumSampleBytes/2))
							if err != nil && n < NumSampleBytes { // Should only return the error here if we didn't read enough bytes
								tree.Err = append(tree.Err, errorx.Decorate(err, "failed to read PARTIAL file"))
							}
						}
					} else {
						sample = &ByteSample{}
					}

				}

				nf.Name = fStat.Name()
				nf.ByteSample = sample
				nf.Size = fStat.Size()

				tree.Size += fStat.Size()
				currModified := fStat.ModTime()
				newestModified = GetNewestTime(newestModified, &currModified)
			}

			tree.Files = append(tree.Files, nf)
		}
	}
	tree.TimeTaken = time.Since(startWalk)
	tree.NumFilesTotal = int64(len(tree.Files)) + numFilesInSubtrees
	tree.LastModified = newestModified
	tree.LastVisited = time.Now()
	tree.Depth = depth

	return tree.Err, tree.Size, tree.NumFilesTotal, newestModified
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
