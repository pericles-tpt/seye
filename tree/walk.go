package tree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/joomcode/errorx"
)

var (
	Here         = time.Local
	UnixYear2048 = time.Date(2025, 1, 1, 0, 0, 0, 0, Here).Unix()
	AllTrees     = []FileTree{}
	ignoredDirs  = []string{"proc", "Fiye"}
	chosenHash   = SHA256
	numFiles     = 0
)

func Walk(path string, depth int, isComprehensive bool) (tree FileTree, errsBelow []string, size int64, numFilesR int64, latestModified *time.Time) {
	tree = FileTree{
		BasePath: path,
	}
	tree.Comprehensive = isComprehensive

	ents, err := os.ReadDir(path)
	tree.ErrStrings = []string{}
	if err != nil {
		tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to open `tree.BasePath`").Error())
	}

	var (
		newestModified     *time.Time = nil
		startWalk                     = time.Now()
		numFilesInSubtrees int64      = 0
	)
	for _, e := range ents {
		fullPath := fmt.Sprintf("%s/%s", tree.BasePath, e.Name())

		if e.IsDir() {
			if contains(ignoredDirs, e.Name()) {
				continue
			}

			subTree, err, subSize, subNumFiles, currModified := Walk(fullPath, depth+1, isComprehensive)
			if len(err) > 0 {
				tree.ErrStrings = append(tree.ErrStrings, err...)
			}
			tree.Size += subSize
			numFilesInSubtrees += subNumFiles

			tree.SubTrees = append(tree.SubTrees, subTree)
			newestModified = GetNewestTime(newestModified, currModified)
		} else if e.Type().IsRegular() {
			fStat, err := e.Info()
			nf := File{
				Err:  "",
				Hash: nil,
			}
			if err != nil {
				tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to stat file").Error())
			} else {
				if isComprehensive {
					nf.Hash = &Hash{}

					fTmp, err := os.OpenFile(fullPath, os.O_RDONLY, 0400)
					if err != nil {
						tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to open file to get `Hash` for 'Comprehensive' scan").Error())
					} else if fStat.Size() > 0 {
						var b bytes.Buffer

						n, err := b.ReadFrom(fTmp)
						if err != nil {
							tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to read WHOLE file").Error())
						}

						if n <= int64(chosenHash) {
							nf.Hash.Type = NONE
							copy(nf.Hash.Bytes[:], b.Bytes())
						} else {
							nf.Hash.Type = chosenHash
							nf.Hash.Bytes = generateSHAHash(b.Bytes())
						}

					}

				}

				nf.Name = fStat.Name()
				nf.Size = fStat.Size()
				nf.LastModified = fStat.ModTime()

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

	return tree, tree.ErrStrings, tree.Size, tree.NumFilesTotal, newestModified
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

func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

func generateSHAHash(bs []byte) [sha256.Size]byte {
	return sha256.Sum256(bs)
}

// func getNEqualTimeSubtrees(tree FileTree, n, percentDiffAllowed int) []string {

// }
