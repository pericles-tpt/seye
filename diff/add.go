package diff

import (
	"path"

	"github.com/Fiye/file"
	"github.com/Fiye/tree"
)

/*
Adds a `TreeDiff` onto a `FileTree`, returning the resultant `FileTree`

NOTE: Assumes the `FileTree` and `TreeDiff` have the same root
*/
func WalkAddDiff(t *tree.FileTree, d *ScanDiff, newAllHash *[]byte, addedTrees []TreeDiff, addedFiles []FileDiff) bool {
	// At depth 0, get `addedTrees` and `addedFile` to pass to deeper recursions
	if t.Depth == 0 {
		for _, d := range d.Trees {
			if d.Type == added {
				addedTrees = append(addedTrees, d)
			}
		}

		for _, f := range d.Files {
			if f.Type == added {
				addedFiles = append(addedFiles, f)
			}
		}
	}

	// Check if we can add any new trees or files here
	for _, at := range addedTrees {
		if t.BasePath == path.Dir(at.NewerPath) {
			addDiffToTree(t, &at, newAllHash)
			d.Trees[at.NewerPath] = TreeDiff{}
		}
	}
	for _, af := range addedFiles {
		if t.BasePath == path.Dir(af.NewerName) {
			nf := file.File{
				Hash: file.InitialiseHashLocation(nil, nil, nil),
			}
			_, diffEmpty := addDiffToFile(&nf, &af, &d.AllHash, newAllHash)
			if !diffEmpty {
				t.Files = append(t.Files, nf)
			}
			d.Files[af.NewerName] = FileDiff{}
		}
	}

	// `DiffMaps` will contain EITHER exact matches or partial matches for files in the tree...
	diff, ok := d.Trees[t.BasePath]
	removeTree := false
	if ok {
		removeTree = addDiffToTree(t, &diff, newAllHash)
		d.Trees[t.BasePath] = TreeDiff{}
	}
	if removeTree {
		return true
	}

	if t != nil {
		newFiles := []file.File{}
		for _, f := range t.Files {
			removeFile := false
			fDiff, ok := d.Files[f.Name]
			if ok {
				removeFile, _ = addDiffToFile(&f, &fDiff, &d.AllHash, newAllHash)
				d.Files[f.Name] = FileDiff{}
			}

			if !removeFile {
				newFiles = append(newFiles, f)
			}
		}
		t.Files = newFiles

		newSubTrees := []tree.FileTree{}
		for _, st := range t.SubTrees {
			removeTree = false
			removeTree = WalkAddDiff(&st, d, newAllHash, addedTrees, addedFiles)
			if !removeTree {
				newSubTrees = append(newSubTrees, st)
			}
		}
		t.SubTrees = newSubTrees
	}

	return false
}

func addDiffToTree(t *tree.FileTree, d *TreeDiff, newAllHash *[]byte) bool {
	if d.isEmpty() {
		return false
	}

	switch d.Type {
	case changed:
		t.Comprehensive = d.Comprehensive
		t.BasePath = d.NewerPath
		t.Depth += d.DepthDiff
		t.ErrStrings = append(t.ErrStrings, d.ErrStringsDiff...)
		t.LastVisited = t.LastVisited.Add(d.LastVisitedDiff)
		t.LastModified = t.LastModified.Add(d.LastModifiedDiff)
		t.Size += d.SizeDiff
		t.NumFilesTotal += d.NumFilesTotalDiff
		t.AllHashOffset = d.AllHashOffset
		t.TimeTaken += d.TimeTakenDiff
		if t.Depth > 0 {
			t.AllHash = d.AllHash
		}
	case renamed:
		t.BasePath = d.NewerPath
	case removed:
		return true
	default:
	}

	return false
}

func addDiffToFile(f *file.File, d *FileDiff, oldAllHash, newAllHash *[]byte) (removeFile bool, diffEmpty bool) {
	if d.isEmpty() {
		return false, true
	}

	switch d.Type {
	case changed:
		f.Name = d.NewerName
		f.Err = d.NewerErr
		f.LastModified = f.LastModified.Add(d.LastModifiedDiff)
		f.Size += d.SizeDiff
		f.Hash = file.InitialiseHashLocation(nil, nil, nil)
		if d.HashDiff.HashOffset > -1 {
			f.Hash = addNewHash(d.HashDiff, oldAllHash, newAllHash)
		}
	case renamed:
		f.Name = d.NewerName
	case removed:
		return true, false
	default:
	}

	return false, false
}

func addNewHash(addFromLocation file.HashLocation, fromAllHash, toAllHash *[]byte) file.HashLocation {
	newHashOffset := len(*toAllHash) - 1
	*toAllHash = append(*toAllHash, (*fromAllHash)[addFromLocation.HashOffset:addFromLocation.HashOffset+addFromLocation.HashLength]...)
	return file.HashLocation{
		HashOffset: newHashOffset,
		HashLength: addFromLocation.HashLength,
		Type:       addFromLocation.Type,
	}
}

func addHashAtOffset(offset int, length int, hashType file.HashType, hashBytes []byte, AllHash *[]byte) file.HashLocation {
	for i := 0; i < length; i++ {
		(*AllHash)[offset+i] = hashBytes[i]
	}
	return file.HashLocation{
		HashOffset: offset,
		HashLength: length,
		Type:       hashType,
	}
}
