package diff

import (
	"bytes"
	"time"

	"github.com/Fiye/tree"
)

func diffFiles(a, b []tree.File) ([]int, []FileDiff) {
	var (
		differentFiles = []FileDiff{}

		changedAFiles = []int{}
		changesFoundB = map[int]struct{}{}
	)

	// Iterate through `a` looking for exact matches with `b`, renames or changed files
	for i, fa := range a {
		var (
			fileUnchanged = -1
			fileRenamed   = -1
			fileChanged   = -1
		)

		for j, fb := range b {
			var (
				nameSame = fa.Name == fb.Name
				sizeSame = fa.Size == fb.Size
				modSame  = fa.LastModified == fb.LastModified
			)

			if nameSame && sizeSame && modSame {
				fileUnchanged = j

				// If hashes are available compare them too
				hashDiff := diffHash(fa.Hash, fb.Hash)

				if !(fa.Hash == nil && fb.Hash == nil) {
					if (fa.Hash != nil && fb.Hash == nil) || (fa.Hash == nil && fb.Hash != nil) ||
						((*fa.Hash).Type != (*fa.Hash).Type) ||
						!bytes.Equal((*hashDiff).Bytes[:], make([]byte, len((*hashDiff).Bytes))) {
						fileUnchanged = -1
					}
				}
			} else if !nameSame && sizeSame && modSame {
				fileRenamed = j

				// If hashes are available compare them too
				hashDiff := diffHash(fa.Hash, fb.Hash)

				if !(fa.Hash == nil && fb.Hash == nil) {
					if (fa.Hash != nil && fb.Hash == nil) || (fa.Hash == nil && fb.Hash != nil) ||
						((*fa.Hash).Type != (*fa.Hash).Type) ||
						!bytes.Equal((*hashDiff).Bytes[:], make([]byte, len((*hashDiff).Bytes))) {
						fileRenamed = -1
					}
				}
			} else if nameSame && !modSame {
				fileChanged = j
			}

			if fileUnchanged >= 0 {
				break
			}
		}

		// At this point we know the following changes:
		//	- 'unchanged': `continue`, we return an array of indices where there are differences, the indices left out are unchanged
		//  - 'renamed': add the changed 'Name' to `differentFiles`
		//	- 'changed': add the changed size?, hashDiff? and lastModifiedDiff to `differentFiles`
		if fileUnchanged >= 0 {
			changesFoundB[fileUnchanged] = struct{}{}
			continue
		} else if fileRenamed >= 0 {
			differentFiles = append(differentFiles, FileDiff{
				NewerName: b[fileRenamed].Name,
				HashDiff:  b[fileRenamed].Hash,
			})

			if b[fileRenamed].Hash != nil {
				differentFiles[len(differentFiles)-1].HashDiff.Bytes = [32]byte{}
			}

			changesFoundB[fileRenamed] = struct{}{}
		} else if fileChanged >= 0 {
			newer := b[fileChanged]
			older := fa

			// Can't serialise nil... empty value could -> same OR both nil...
			differentFiles = append(differentFiles, FileDiff{
				NewerName:        newer.Name,
				SizeDiff:         newer.Size - older.Size,
				NewerErr:         newer.Err,
				HashDiff:         diffHash(older.Hash, newer.Hash),
				LastModifiedDiff: newer.LastModified.Sub(older.LastModified),
			})

			changesFoundB[fileChanged] = struct{}{}
		} else { // -> file removed
			// -ve time -> file removed, also -ve size (maybe)
			differentFiles = append(differentFiles, FileDiff{
				NewerName:        fa.Name,
				SizeDiff:         -fa.Size,
				NewerErr:         fa.Err,
				HashDiff:         diffHash(fa.Hash, &tree.Hash{Type: fa.Hash.Type}),
				LastModifiedDiff: time.Time{}.Sub(fa.LastModified),
			})
		}
		changedAFiles = append(changedAFiles, i)
	}

	// Iterate through `b` again to find the added files (i.e. not matched with anything in a), put them at the end of `differentFiles`
	for i, fb := range b {
		_, ok := changesFoundB[i]
		if !ok { // Index wasn't matched so add to the end of different files
			differentFiles = append(differentFiles, FileDiff{
				NewerName:        fb.Name,
				NewerErr:         fb.Err,
				HashDiff:         diffHash(&tree.Hash{Type: fb.Hash.Type}, fb.Hash),
				SizeDiff:         fb.Size,
				LastModifiedDiff: fb.LastModified.Sub(time.Time{}),
			})
		}
	}

	return changedAFiles, differentFiles
}

func diffTrees(a, b []tree.FileTree, isComprehensive bool) ([]int, []TreeDiff) {
	var (
		differentTrees = []TreeDiff{}

		changedATrees = []int{}
		changesFoundB = map[int]struct{}{}

		// These maps for `changedFiles` between subtrees are for caching to avoid unnecessary diffFiles() calls
		changedFileIndices = map[int]map[int][]int{}
		differentFiles     = map[int]map[int][]FileDiff{}
	)

	// Iterate through `a` looking for exact matches with `b`, renames or changed files
	for i, ta := range a {
		var (
			treeUnchanged = -1
			treeRenamed   = -1
			treeChanged   = -1
		)

		for j, tb := range b {
			var (
				nameSame = ta.BasePath == tb.BasePath
				sizeSame = ta.Size == tb.Size
				modSame  = ta.LastModified == tb.LastModified
			)

			if nameSame && sizeSame && modSame {
				treeUnchanged = j

				if isComprehensive {
					if changedFileIndices[i] == nil || differentFiles[i] == nil {
						changedFileIndices[i] = map[int][]int{}
						differentFiles[i] = map[int][]FileDiff{}
					}
					if changedFileIndices[i][j] == nil || differentFiles[i][j] == nil {
						changedFileIndices[i][j], differentFiles[i][j] = diffFiles(ta.Files, tb.Files)
					}

					if len(differentFiles[i][j]) > 0 {
						treeUnchanged = -1
					}
				}
			} else if !nameSame && sizeSame && modSame {
				treeRenamed = j

				if isComprehensive {
					if changedFileIndices[i] == nil || differentFiles[i] == nil {
						changedFileIndices[i] = map[int][]int{}
						differentFiles[i] = map[int][]FileDiff{}
					}
					if changedFileIndices[i][j] == nil || differentFiles[i][j] == nil {
						changedFileIndices[i][j], differentFiles[i][j] = diffFiles(ta.Files, tb.Files)
					}

					if len(differentFiles[i][j]) > 0 {
						treeRenamed = -1
					}
				}
			} else if nameSame && !modSame {
				treeChanged = j
			}

			if treeUnchanged >= 0 {
				break
			}
		}

		// At this point we know the following changes:
		//	- 'unchanged': `continue`, we return an array of indices where there are differences, the indices left out are unchanged
		//  - 'renamed': add the changed 'Name' to `differentFiles`
		//	- 'changed': add the changed size?, hashDiff? and lastModifiedDiff to `differentFiles`
		if treeUnchanged >= 0 { // -> nil so do nothing here
			changesFoundB[treeUnchanged] = struct{}{}
			continue
		} else if treeRenamed >= 0 {
			differentTrees = append(differentTrees, TreeDiff{
				NewerPath: b[treeRenamed].BasePath,
			})

			changesFoundB[treeRenamed] = struct{}{}
		} else if treeChanged >= 0 {
			newer := b[treeChanged]
			older := ta

			if changedFileIndices[i] == nil || differentFiles[i] == nil {
				changedFileIndices[i] = map[int][]int{}
				differentFiles[i] = map[int][]FileDiff{}
			}
			if changedFileIndices[i][treeChanged] == nil || differentFiles[i][treeChanged] == nil {
				changedFileIndices[i][treeChanged], differentFiles[i][treeChanged] = diffFiles(older.Files, newer.Files)
			}
			stDiffIdx, stDiff := diffTrees(older.SubTrees, newer.SubTrees, isComprehensive)

			alm := time.Time{}
			blm := time.Time{}
			if older.LastModified != nil {
				alm = *older.LastModified
			}
			if newer.LastModified != nil {
				blm = *newer.LastModified
			}

			differentTrees = append(differentTrees, TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: newer.Comprehensive,

				NewerPath:        newer.BasePath,
				FilesDiff:        differentFiles[i][treeChanged],
				FilesDiffIndices: changedFileIndices[i][treeChanged],
				LastVisitedDiff:  newer.LastVisited.Sub(older.LastVisited),
				TimeTakenDiff:    newer.TimeTaken - older.TimeTaken,
				LastModifiedDiff: blm.Sub(alm),
				DepthDiff:        newer.Depth - older.Depth,
				ErrStringsDiff:   getStringArrayDiff(older.ErrStrings, newer.ErrStrings),

				SubTreesDiff:        stDiff,
				SubTreesDiffIndices: stDiffIdx,
				SizeDiff:            newer.Size - older.Size,
				NumFilesTotalDiff:   newer.NumFilesTotal - older.NumFilesTotal,
			})

			changesFoundB[treeChanged] = struct{}{}
		} else { // -> tree removed
			lm := time.Time{}
			if ta.LastModified != nil {
				lm = *ta.LastModified
			}

			stDiffIdx, stDiff := diffTrees(ta.SubTrees, []tree.FileTree{}, isComprehensive)
			fDiffIdx, fDiff := diffFiles(ta.Files, []tree.File{})

			// -ve time -> file removed, also -ve size (maybe)
			differentTrees = append(differentTrees, TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: ta.Comprehensive,

				NewerPath:        ta.BasePath,
				FilesDiff:        fDiff,
				FilesDiffIndices: fDiffIdx,
				LastVisitedDiff:  time.Time{}.Sub(ta.LastVisited),
				TimeTakenDiff:    -ta.TimeTaken,
				LastModifiedDiff: time.Time{}.Sub(lm),
				DepthDiff:        -ta.Depth,
				ErrStringsDiff:   []string{},

				SubTreesDiff:        stDiff,
				SubTreesDiffIndices: stDiffIdx,
				SizeDiff:            -ta.Size,
				NumFilesTotalDiff:   -ta.NumFilesTotal,
			})
		}
		changedATrees = append(changedATrees, i)
	}

	// Iterate through `b` again to find the added trees (i.e. not matched with anything in a), put them at the end of `differentFiles`
	for i, tb := range b {
		_, ok := changesFoundB[i]
		if !ok { // Index wasn't matched so add to the end of different files
			lm := time.Time{}
			if tb.LastModified != nil {
				lm = *tb.LastModified
			}

			stDiffIdx, stDiff := diffTrees([]tree.FileTree{}, tb.SubTrees, isComprehensive)
			fDiffIdx, fDiff := diffFiles([]tree.File{}, tb.Files)

			differentTrees = append(differentTrees, TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: tb.Comprehensive,

				NewerPath:        tb.BasePath,
				FilesDiff:        fDiff,
				FilesDiffIndices: fDiffIdx,
				LastVisitedDiff:  tb.LastVisited.Sub(time.Time{}),
				TimeTakenDiff:    tb.TimeTaken,
				LastModifiedDiff: lm.Sub(time.Time{}),
				DepthDiff:        tb.Depth,
				ErrStringsDiff:   tb.ErrStrings,

				SubTreesDiff:        stDiff,
				SubTreesDiffIndices: stDiffIdx,
				SizeDiff:            tb.Size,
				NumFilesTotalDiff:   tb.NumFilesTotal,
			})
		}
	}

	return changedATrees, differentTrees
}

func diffHash(a, b *tree.Hash) *tree.Hash {
	// If hashes are comparable, return their byte difference, otherwise just return `b`
	if (a != nil && b != nil) && (*a).Type == (*b).Type {
		for i, aByte := range (*a).Bytes {
			(*b).Bytes[i] -= aByte
		}
	}
	return b
}

func getStringArrayDiff(a, b []string) []string {
	aMap := map[string]int{}
	for _, s := range a {
		aMap[s] = -1
	}

	for _, s := range b {
		_, ok := aMap[s]
		if ok {
			aMap[s] += 1
		} else {
			aMap[s] = 1
		}
	}

	ret := []string{}
	for s, n := range aMap {
		if n > 0 {
			for i := 0; i < n; i++ {
				ret = append(ret, s)
			}
		}
	}

	return ret
}
