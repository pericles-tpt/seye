package tree

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Fiye/file"
	"github.com/Fiye/stats"
	"github.com/Fiye/utility"
	"github.com/joomcode/errorx"
)

var (
	Here         = time.Local
	UnixYear2048 = time.Date(2025, 1, 1, 0, 0, 0, 0, Here).Unix()
	AllTrees     = []FileTree{}
	ignoredDirs  = []string{"proc", "Fiye"}
	chosenHash   = sha256.Size
	copyBuffer   = make([]byte, 1048576)

	filesAddedForHashing = 0
	noHashOffset         = -1

	AllHashBytes = []byte{}
)

func WalkGenerateTree(path string, depth int, isComprehensive bool, walkStats *stats.WalkStats) (tree FileTree) {
	if depth == 0 {
		filesAddedForHashing = 0
		AllHashBytes = []byte{}
	}

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
		// TODO: Setting newestModified to `time.Time{}` might cause issues if a file is incidentally created at the time...
		newestModified           = time.Time{}
		startWalk                = time.Now()
		numFilesInSubtrees int64 = 0
	)
	for _, e := range ents {
		fullPath := fmt.Sprintf("%s/%s", tree.BasePath, e.Name())
		if strings.HasSuffix(tree.BasePath, "/") {
			fullPath = fmt.Sprintf("%s%s", tree.BasePath, e.Name())
		}

		if e.IsDir() {
			if utility.Contains(ignoredDirs, e.Name()) {
				continue
			}

			subTree := WalkGenerateTree(fullPath, depth+1, isComprehensive, walkStats)
			if len(subTree.ErrStrings) > 0 {
				tree.ErrStrings = append(tree.ErrStrings, subTree.ErrStrings...)
			}
			tree.Size += subTree.Size
			numFilesInSubtrees += subTree.NumFilesTotal

			tree.SubTrees = append(tree.SubTrees, subTree)
			newestModified = utility.GetNewestTime(newestModified, subTree.LastModified)
		} else if e.Type().IsRegular() {
			fStat, err := e.Info()
			nf := file.File{
				Hash: file.InitialiseHashLocation(nil, nil, nil),
			}
			if err != nil {
				tree.ErrStrings = append(tree.ErrStrings, errorx.Decorate(err, "failed to stat file").Error())
			} else {
				nf = getFileData(fullPath, fStat, isComprehensive, &tree, walkStats)

				if walkStats != nil {
					walkStats.UpdateLargestFiles(fullPath, fStat)
				}

				newestModified = utility.GetNewestTime(newestModified, nf.LastModified)
			}

			tree.Files = append(tree.Files, nf)
		}
	}
	tree.TimeTaken = time.Since(startWalk)
	tree.NumFilesTotal = int64(len(tree.Files)) + numFilesInSubtrees
	tree.LastModified = newestModified
	tree.LastVisited = time.Now()
	tree.Depth = depth

	if tree.Depth == 0 && len(AllHashBytes) > 0 {
		tree.AllHash = AllHashBytes
	}

	return tree
}

func getFileData(fullPath string, fStat os.FileInfo, isComprehensive bool, tree *FileTree, walkStats *stats.WalkStats) file.File {
	nf := file.File{
		Hash: file.InitialiseHashLocation(nil, nil, nil),
	}

	if isComprehensive {
		fTmp, err := os.OpenFile(fullPath, os.O_RDONLY, 0400)
		if err != nil {
			(*tree).ErrStrings = append((*tree).ErrStrings, errorx.Decorate(err, "failed to open file to get `Hash` for 'Comprehensive' scan").Error())
		} else if fStat.Size() > 0 {
			h := sha256.New()
			_, err := io.CopyBuffer(h, fTmp, copyBuffer)
			if err != nil {
				(*tree).ErrStrings = append((*tree).ErrStrings, errorx.Decorate(err, "failed to read WHOLE file").Error())
			} else {
				newHashOffset := len(AllHashBytes)
				hashedBytes := h.Sum(nil)
				AllHashBytes = append(AllHashBytes, hashedBytes[:]...)
				nf.Hash.Type = file.SHA256
				nf.Hash.HashOffset = newHashOffset
				nf.Hash.HashLength = len(hashedBytes)

				if walkStats != nil {
					walkStats.UpdateDuplicates(hashedBytes[:], fStat.Size(), fullPath)
				}

			}
		}
		defer fTmp.Close()
	}

	nf.Name = fullPath
	nf.Size = fStat.Size()
	nf.LastModified = fStat.ModTime()

	(*tree).Size += fStat.Size()

	return nf
}
