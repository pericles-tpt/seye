package diff

import (
	"runtime"
	"time"

	"github.com/pericles-tpt/seye/tree"
	"github.com/pericles-tpt/seye/utility"
)

/*
Find and returns differences (renamed, removed, added or changed) between two File arrays
*/
func diffFiles(a, b []tree.File, allHashesA, allHashesB, allHashDiff *[]byte, sDiff *ScanDiff) ([]int, []FileDiff) {
	var (
		differentFiles = []FileDiff{}
		changedAFiles  = []int{}
		changesFoundB  = map[int]struct{}{}
	)

	// 1. Iterate through a, then b (for each a). By comparing files in `a` to `b`, classify them as 'unchanged', 'renamed' or 'changed'
	for i, fa := range a {
		var (
			fileUnchanged = -1
			fileRenamed   = -1
			fileChanged   = -1
		)

		// 1a. Compare THIS file in `a`, to each file in `b`, try to find one it matches with for the conditions listed in `var`
		for j, fb := range b {
			var (
				nameSame   = fa.Name == fb.Name
				modSame    = time.Time.Equal(fa.LastModified, fb.LastModified)
				hashesSame = utility.HashesEqual(fa.Hash, fb.Hash, allHashesA, allHashesB)
			)

			if hashesSame {
				if !nameSame {
					fileRenamed = j
				} else {
					fileUnchanged = j
				}
			} else if !modSame {
				fileChanged = j
			}

			if fileUnchanged >= 0 {
				break
			}
		}

		// 1b. For THIS file in `a`, we now know IF it has been modified and HOW, add information about this file to `sDiff` IF it's modified
		if fileUnchanged >= 0 {
			changesFoundB[fileUnchanged] = struct{}{}
			continue
		} else if fileRenamed >= 0 {
			sDiff.Files[fa.Name] = FileDiff{
				NewerName: b[fileRenamed].Name,
				Type:      renamed,
			}

			differentFiles = append(differentFiles, sDiff.Files[fa.Name])

			changesFoundB[fileRenamed] = struct{}{}
		} else if fileChanged >= 0 {
			newer := b[fileChanged]
			older := fa
			lastAllHashByte := len(*allHashDiff) - 1

			fDiff := FileDiff{
				Type:             changed,
				NewerName:        newer.Name,
				SizeDiff:         newer.Size - older.Size,
				NewerErr:         newer.Err,
				LastModifiedDiff: newer.LastModified.Sub(older.LastModified),
				HashDiff:         utility.InitialiseHashLocation(),
			}

			if newer.Hash.HashOffset > -1 {
				fDiff.HashDiff = utility.HashLocation{Type: newer.Hash.Type, HashOffset: lastAllHashByte, HashLength: newer.Hash.HashLength}
				*allHashDiff = append(*allHashDiff, (*allHashesB)[newer.Hash.HashOffset:newer.Hash.HashOffset+newer.Hash.HashLength]...)
			}

			sDiff.Files[fa.Name] = fDiff
			differentFiles = append(differentFiles, fDiff)

			changesFoundB[fileChanged] = struct{}{}
		} else { // -> removed
			sDiff.Files[fa.Name] = FileDiff{
				NewerName:        fa.Name,
				Type:             removed,
				SizeDiff:         -fa.Size,
				NewerErr:         fa.Err,
				HashDiff:         utility.InitialiseHashLocation(),
				LastModifiedDiff: utility.GoSpecialTime.Sub(fa.LastModified),
			}
			differentFiles = append(differentFiles, sDiff.Files[fa.Name])
		}

		// Record the index of this modified file, so we know which files in `a`, have been modified
		changedAFiles = append(changedAFiles, i)
	}

	// 2. Loop through `b` again, to find files in `b` but not in `a`, i.e. ADDED files
	for i, fb := range b {
		lastAllHashByte := len(*allHashDiff) - 1
		_, ok := changesFoundB[i]
		// File NOT recorded in `changesFoundB` -> it's an ADDED file
		if !ok {
			fDiff := FileDiff{
				Type:             added,
				NewerName:        fb.Name,
				NewerErr:         fb.Err,
				SizeDiff:         fb.Size,
				LastModifiedDiff: fb.LastModified.Sub(utility.GoSpecialTime),
				HashDiff:         utility.InitialiseHashLocation(),
			}

			if fb.Hash.HashOffset > -1 {
				fDiff.HashDiff = utility.HashLocation{Type: fb.Hash.Type, HashOffset: lastAllHashByte, HashLength: fb.Hash.HashLength}
				*allHashDiff = append(*allHashDiff, (*allHashesB)[fb.Hash.HashOffset:fb.Hash.HashOffset+fb.Hash.HashLength]...)
			}

			sDiff.Files[fb.Name] = fDiff
			differentFiles = append(differentFiles, fDiff)
		}
	}

	return changedAFiles, differentFiles
}

/*
Find and returns differences (renamed, removed, added or changed) between two FileTree arrays
*/
func diffTrees(a, b []tree.FileTree, aHashes, bHashes, allHashDiff *[]byte, isComprehensive bool, sDiff *ScanDiff) ([]int, ScanDiff) {
	if (len(a) > 0 && a[0].Depth == 0) || allHashDiff == nil {
		allHashDiff = &[]byte{}
	}

	var (
		changedATrees = []int{}
		changesFoundB = map[int]struct{}{}

		// Performing `diffFiles` between two subtrees is expensive, so we cache the comparison results here, to avoid unnecessary calls
		changedFileIndices = map[int]map[int][]int{}
		changedFiles       = map[int]map[int][]FileDiff{}
	)

	// 1. Iterate through a, then b (for each a). By comparing `FileTree`s in `a` to `b`, classify them as 'unchanged', 'renamed' or 'changed'
	for i, ta := range a {
		var (
			treeUnchanged = -1
			treeRenamed   = -1
			treeChanged   = -1
		)

		// 1a. Compare THIS FileTree in `a`, to each FileTree in `b`, try to find one it matches with for the conditions listed in `var`
		for j, tb := range b {
			var (
				nameSame = ta.BasePath == tb.BasePath
				sizeSame = ta.SizeDirect == tb.SizeDirect
				modSame  = time.Time.Equal(ta.LastModifiedDirect, tb.LastModifiedDirect)
			)

			if nameSame && sizeSame && modSame {
				treeUnchanged = j

				if changedFileIndices[i] == nil || changedFiles[i] == nil {
					changedFileIndices[i] = map[int][]int{}
					changedFiles[i] = map[int][]FileDiff{}
				}
				if changedFileIndices[i][j] == nil || changedFiles[i][j] == nil {
					changedFileIndices[i][j], changedFiles[i][j] = diffFiles(ta.Files, tb.Files, aHashes, bHashes, allHashDiff, sDiff)
				}

				if len(changedFiles[i][j]) > 0 {
					treeUnchanged = -1
					treeChanged = j
					break
				}
			} else if !nameSame && sizeSame && modSame {
				treeRenamed = j

				if changedFileIndices[i] == nil || changedFiles[i] == nil {
					changedFileIndices[i] = map[int][]int{}
					changedFiles[i] = map[int][]FileDiff{}
				}
				if changedFileIndices[i][j] == nil || changedFiles[i][j] == nil {
					changedFileIndices[i][j], changedFiles[i][j] = diffFiles(ta.Files, tb.Files, aHashes, bHashes, allHashDiff, sDiff)
				}

				if len(changedFiles[i][j]) > 0 {
					treeRenamed = -1
					treeChanged = j
					break
				}
				// TODO: For some reason, when adding a file to a directory, that "shifts" the position of other files/directory down, the
				// LastModified time seems to be changed on MacOS, idk why this happens, needs further investigation
			} else if nameSame && ((runtime.GOOS == "darwin" && !sizeSame) || (runtime.GOOS != "darwin" && !modSame)) {
				treeChanged = j
			}

			if treeUnchanged >= 0 {
				break
			}
		}

		// 1b. For THIS FileTree in `a`, we now know IF it has been modified and HOW, add information about this file to `sDiff` IF it's modified
		if treeUnchanged >= 0 { // -> nil so do nothing here
			changesFoundB[treeUnchanged] = struct{}{}

			// 1bi. At the moment we do not know if a change has occured below this tree, so we need to diff the trees below it to check for changes
			_, _ = diffTrees(ta.SubTrees, b[treeUnchanged].SubTrees, aHashes, bHashes, allHashDiff, isComprehensive, sDiff)
			continue
		} else if treeRenamed >= 0 {
			sDiff.Trees[ta.BasePath] = TreeDiff{
				NewerPath: b[treeRenamed].BasePath,
				Type:      renamed,
			}

			changesFoundB[treeRenamed] = struct{}{}
		} else if treeChanged >= 0 {
			newer := b[treeChanged]
			older := ta

			if changedFileIndices[i] == nil || changedFiles[i] == nil {
				changedFileIndices[i] = map[int][]int{}
				changedFiles[i] = map[int][]FileDiff{}
			}
			if changedFileIndices[i][treeChanged] == nil || changedFiles[i][treeChanged] == nil {
				changedFileIndices[i][treeChanged], changedFiles[i][treeChanged] = diffFiles(older.Files, newer.Files, aHashes, bHashes, allHashDiff, sDiff)
			}
			stDiffIdx, _ := diffTrees(older.SubTrees, newer.SubTrees, aHashes, bHashes, allHashDiff, isComprehensive, sDiff)

			alm := utility.GoSpecialTime
			blm := utility.GoSpecialTime
			if older.LastModifiedDirect != utility.GoSpecialTime {
				alm = older.LastModifiedDirect
			}
			if newer.LastModifiedDirect != utility.GoSpecialTime {
				blm = newer.LastModifiedDirect
			}

			sDiff.Trees[ta.BasePath] = TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: newer.Comprehensive,
				Type:          changed,

				NewerPath:              newer.BasePath,
				FilesDiff:              changedFiles[i][treeChanged],
				FilesDiffIndices:       changedFileIndices[i][treeChanged],
				LastVisitedDiff:        newer.LastVisited.Sub(older.LastVisited),
				TimeTakenDiff:          newer.TimeTaken - older.TimeTaken,
				LastModifiedDiffDirect: blm.Sub(alm),
				DepthDiff:              newer.Depth - older.Depth,
				ErrStringsDiff:         utility.AdditionalStringsInB(older.ErrStrings, newer.ErrStrings),

				SubTreesDiffIndices:     stDiffIdx,
				SizeDiffDirect:          newer.SizeDirect - older.SizeDirect,
				NumFilesTotalDiffDirect: newer.NumFilesDirect - older.NumFilesDirect,
			}

			changesFoundB[treeChanged] = struct{}{}
		} else { // -> tree removed
			lm := utility.GoSpecialTime
			if ta.LastModifiedDirect != utility.GoSpecialTime {
				lm = ta.LastModifiedDirect
			}

			sDiff.Trees[ta.BasePath] = TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: ta.Comprehensive,
				Type:          removed,

				NewerPath:              ta.BasePath,
				LastVisitedDiff:        utility.GoSpecialTime.Sub(ta.LastVisited),
				TimeTakenDiff:          -ta.TimeTaken,
				LastModifiedDiffDirect: utility.GoSpecialTime.Sub(lm),
				DepthDiff:              -ta.Depth,
				ErrStringsDiff:         []string{},

				SizeDiffDirect:          -ta.SizeDirect,
				NumFilesTotalDiffDirect: -ta.NumFilesDirect,
			}
		}

		// Record the index of this modified FileTree, so we know which files in `a`, have been modified
		changedATrees = append(changedATrees, i)
	}

	// 2. Loop through `b` again, to find files in `b` but not in `a`, i.e. ADDED FileTrees
	for i, tb := range b {
		_, ok := changesFoundB[i]
		// File NOT recorded in `changesFoundB` -> it's an ADDED FileTree
		if !ok {
			lm := utility.GoSpecialTime
			if tb.LastModifiedDirect != utility.GoSpecialTime {
				lm = tb.LastModifiedDirect
			}

			stDiffIdx, _ := diffTrees([]tree.FileTree{}, tb.SubTrees, &([]byte{}), bHashes, allHashDiff, isComprehensive, sDiff)
			fDiffIdx, fDiff := diffFiles([]tree.File{}, tb.Files, &([]byte{}), bHashes, allHashDiff, sDiff)

			sDiff.Trees[tb.BasePath] = TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: tb.Comprehensive,
				Type:          added,

				NewerPath:              tb.BasePath,
				FilesDiff:              fDiff,
				FilesDiffIndices:       fDiffIdx,
				LastVisitedDiff:        tb.LastVisited.Sub(utility.GoSpecialTime),
				TimeTakenDiff:          tb.TimeTaken,
				LastModifiedDiffDirect: lm.Sub(utility.GoSpecialTime),
				DepthDiff:              tb.Depth,
				ErrStringsDiff:         tb.ErrStrings,

				SubTreesDiffIndices:     stDiffIdx,
				SizeDiffDirect:          tb.SizeDirect,
				NumFilesTotalDiffDirect: tb.NumFilesDirect,
			}
		}
	}

	if (len(a) == 1 && a[0].Depth == 0) && (len(b) == 1 && b[0].Depth == 0) {
		sDiff.AllHash = *allHashDiff
	}

	return changedATrees, *sDiff
}
