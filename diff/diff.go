package diff

import (
	"bytes"
	"time"

	"github.com/Fiye/file"
	"github.com/Fiye/tree"
)

var (
	AllHashBytesA []byte
	AllHashBytesB []byte
	AllHashDiff   []byte
)

func diffFiles(a, b []file.File, allHashesA, allHashesB *[]byte, diffMap *DiffMaps) ([]int, []FileDiff) {
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
				modSame  = fa.LastModified == fb.LastModified
				// This might seem unnecessary, but untrusted users could fake the lastModified time
				hashesSame = hashesEqual(fa.Hash, fb.Hash, allHashesA, allHashesB)
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

		// At this point we know the following changes:
		//	- 'unchanged': `continue`, we return an array of indices where there are differences, the indices left out are unchanged
		//  - 'renamed': add the changed 'Name' to `differentFiles`
		//	- 'changed': add the changed size?, hashDiff? and lastModifiedDiff to `differentFiles`
		if fileUnchanged >= 0 {
			changesFoundB[fileUnchanged] = struct{}{}
			continue
		} else if fileRenamed >= 0 {
			diffMap.Files[fa.Name] = FileDiff{
				NewerName: b[fileRenamed].Name,
				Type:      renamed,
			}

			differentFiles = append(differentFiles, diffMap.Files[fa.Name])

			changesFoundB[fileRenamed] = struct{}{}
		} else if fileChanged >= 0 {
			newer := b[fileChanged]
			older := fa
			lastAllHashByte := len(AllHashDiff) - 1

			// Can't serialise nil... empty value could -> same OR both nil...
			fDiff := FileDiff{
				Type:             changed,
				NewerName:        newer.Name,
				SizeDiff:         newer.Size - older.Size,
				NewerErr:         newer.Err,
				LastModifiedDiff: newer.LastModified.Sub(older.LastModified),
				HashDiff:         file.InitialiseHashLocation(nil, nil, nil),
			}

			if newer.Hash.HashOffset > -1 {
				fDiff.HashDiff = file.InitialiseHashLocation(&lastAllHashByte, &newer.Hash.Type, &newer.Hash.HashLength)
				AllHashDiff = append(AllHashDiff, (*allHashesB)[newer.Hash.HashOffset:newer.Hash.HashOffset+newer.Hash.HashLength]...)
			}

			diffMap.Files[fa.Name] = fDiff
			differentFiles = append(differentFiles, fDiff)

			changesFoundB[fileChanged] = struct{}{}
		} else { // -> file removed
			// -ve time -> file removed, also -ve size (maybe)
			diffMap.Files[fa.Name] = FileDiff{
				NewerName:        fa.Name,
				Type:             removed,
				SizeDiff:         -fa.Size,
				NewerErr:         fa.Err,
				HashDiff:         file.InitialiseHashLocation(nil, &fa.Hash.Type, nil),
				LastModifiedDiff: time.Time{}.Sub(fa.LastModified),
			}
			differentFiles = append(differentFiles, diffMap.Files[fa.Name])
		}
		changedAFiles = append(changedAFiles, i)
	}

	// Iterate through `b` again to find the added files (i.e. not matched with anything in a), put them at the end of `differentFiles`
	for i, fb := range b {
		lastAllHashByte := len(AllHashDiff) - 1
		_, ok := changesFoundB[i]
		if !ok { // Index wasn't matched so add to the end of different files
			fDiff := FileDiff{
				Type:             added,
				NewerName:        fb.Name,
				NewerErr:         fb.Err,
				SizeDiff:         fb.Size,
				LastModifiedDiff: fb.LastModified.Sub(time.Time{}),
				HashDiff:         file.InitialiseHashLocation(nil, nil, nil),
			}

			if fb.Hash.HashOffset > -1 {
				fDiff.HashDiff = file.InitialiseHashLocation(&lastAllHashByte, &fb.Hash.Type, &fb.Hash.HashLength)
				AllHashDiff = append(AllHashDiff, (*allHashesB)[fb.Hash.HashOffset:fb.Hash.HashOffset+fb.Hash.HashLength]...)
			}

			diffMap.Files[fb.Name] = fDiff
			differentFiles = append(differentFiles, fDiff)
		}
	}

	return changedAFiles, differentFiles
}

func diffTrees(a, b []tree.FileTree, aHashes, bHashes *[]byte, isComprehensive bool, diffMap *DiffMaps) ([]int, DiffMaps) {
	// TODO: Review this condition, supposed to reset AllHashDiff
	if len(a) > 0 && a[0].Depth == 0 {
		AllHashDiff = []byte{}
	}

	var (
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
						changedFileIndices[i][j], differentFiles[i][j] = diffFiles(ta.Files, tb.Files, aHashes, bHashes, diffMap)
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
						changedFileIndices[i][j], differentFiles[i][j] = diffFiles(ta.Files, tb.Files, aHashes, bHashes, diffMap)
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
			diffMap.Trees[ta.BasePath] = TreeDiff{
				NewerPath: b[treeRenamed].BasePath,
				Type:      renamed,
			}

			changesFoundB[treeRenamed] = struct{}{}
		} else if treeChanged >= 0 {
			newer := b[treeChanged]
			older := ta

			if changedFileIndices[i] == nil || differentFiles[i] == nil {
				changedFileIndices[i] = map[int][]int{}
				differentFiles[i] = map[int][]FileDiff{}
			}
			if changedFileIndices[i][treeChanged] == nil || differentFiles[i][treeChanged] == nil {
				changedFileIndices[i][treeChanged], differentFiles[i][treeChanged] = diffFiles(older.Files, newer.Files, aHashes, bHashes, diffMap)
			}
			stDiffIdx, _ := diffTrees(older.SubTrees, newer.SubTrees, aHashes, bHashes, isComprehensive, diffMap)

			alm := time.Time{}
			blm := time.Time{}
			if (older.LastModified != time.Time{}) {
				alm = older.LastModified
			}
			if (newer.LastModified != time.Time{}) {
				blm = newer.LastModified
			}

			diffMap.Trees[ta.BasePath] = TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: newer.Comprehensive,
				Type:          changed,

				NewerPath:        newer.BasePath,
				FilesDiff:        differentFiles[i][treeChanged],
				FilesDiffIndices: changedFileIndices[i][treeChanged],
				LastVisitedDiff:  newer.LastVisited.Sub(older.LastVisited),
				TimeTakenDiff:    newer.TimeTaken - older.TimeTaken,
				LastModifiedDiff: blm.Sub(alm),
				DepthDiff:        newer.Depth - older.Depth,
				ErrStringsDiff:   getStringArrayDiff(older.ErrStrings, newer.ErrStrings),

				SubTreesDiffIndices: stDiffIdx,
				SizeDiff:            newer.Size - older.Size,
				NumFilesTotalDiff:   newer.NumFilesTotal - older.NumFilesTotal,
			}

			changesFoundB[treeChanged] = struct{}{}
		} else { // -> tree removed
			lm := time.Time{}
			if (ta.LastModified != time.Time{}) {
				lm = ta.LastModified
			}

			// -ve time -> file removed, also -ve size (maybe)
			diffMap.Trees[ta.BasePath] = TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: ta.Comprehensive,
				Type:          removed,

				NewerPath:        ta.BasePath,
				LastVisitedDiff:  time.Time{}.Sub(ta.LastVisited),
				TimeTakenDiff:    -ta.TimeTaken,
				LastModifiedDiff: time.Time{}.Sub(lm),
				DepthDiff:        -ta.Depth,
				ErrStringsDiff:   []string{},

				SizeDiff:          -ta.Size,
				NumFilesTotalDiff: -ta.NumFilesTotal,
			}
		}
		changedATrees = append(changedATrees, i)
	}

	// Iterate through `b` again to find the added trees (i.e. not matched with anything in a), put them at the end of `differentFiles`
	for i, tb := range b {
		_, ok := changesFoundB[i]
		if !ok { // Index wasn't matched so add to the end of different files
			lm := time.Time{}
			if (tb.LastModified != time.Time{}) {
				lm = tb.LastModified
			}

			stDiffIdx, _ := diffTrees([]tree.FileTree{}, tb.SubTrees, &([]byte{}), bHashes, isComprehensive, diffMap)
			fDiffIdx, fDiff := diffFiles([]file.File{}, tb.Files, &([]byte{}), bHashes, diffMap)

			diffMap.Trees[tb.BasePath] = TreeDiff{
				DiffCompleted: time.Now(),
				Comprehensive: tb.Comprehensive,
				Type:          added,

				NewerPath:        tb.BasePath,
				FilesDiff:        fDiff,
				FilesDiffIndices: fDiffIdx,
				LastVisitedDiff:  tb.LastVisited.Sub(time.Time{}),
				TimeTakenDiff:    tb.TimeTaken,
				LastModifiedDiff: lm.Sub(time.Time{}),
				DepthDiff:        tb.Depth,
				ErrStringsDiff:   tb.ErrStrings,

				SubTreesDiffIndices: stDiffIdx,
				SizeDiff:            tb.Size,
				NumFilesTotalDiff:   tb.NumFilesTotal,
			}
		}
	}

	if (len(a) == 1 && a[0].Depth == 0) && (len(b) == 1 && b[0].Depth == 0) {
		diffMap.AllHash = AllHashDiff
	}

	return changedATrees, *diffMap
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

/*
NOTE: Assumes `HashLocation.HashOffset` is NEVER nil (which should be true)
*/
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
