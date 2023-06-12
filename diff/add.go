package diff

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/joomcode/errorx"
	"github.com/pericles-tpt/seye/config"
	"github.com/pericles-tpt/seye/tree"
	"github.com/pericles-tpt/seye/utility"
)

/*
Adds a `TreeDiff` onto a `FileTree`, returning the resultant `FileTree`

NOTE: Assumes the `FileTree` and `TreeDiff` have the same root
*/
func WalkAddTreeDiff(t *tree.FileTree, d *ScanDiff, newTreeAllHash *[]byte, addedTrees []TreeDiff, addedFiles []FileDiff) (removeThisTree bool) {
	// At depth 0, get `addedTrees` and `addedFiles` to pass to deeper recursions
	if t.Depth == 0 {
		for _, ft := range d.Trees {
			if ft.Type == added {
				addedTrees = append(addedTrees, ft)
			}
		}

		for _, f := range d.Files {
			if f.Type == added {
				addedFiles = append(addedFiles, f)
			}
		}
	}

	// Check if we can add any NEW trees or files, to the current tree `t`
	for _, at := range addedTrees {
		if t.BasePath == path.Dir(at.NewerPath) {
			addDiffToTree(t, &at)
			d.Trees[at.NewerPath] = TreeDiff{}
		}
	}
	for _, af := range addedFiles {
		if t.BasePath == path.Dir(af.NewerName) {
			nf := tree.File{
				Hash: utility.InitialiseHashLocation(),
			}
			_, diffEmpty := addDiffToFile(&nf, &af, &d.AllHash, newTreeAllHash)
			if !diffEmpty {
				t.Files = addFileInAlphaOrder(t.Files, nf)
			}
			d.Files[af.NewerName] = FileDiff{}
		}
	}

	// Check if there are any `TreeDiff`s that apply to the current tree `t`
	diff, ok := d.Trees[t.BasePath]
	removeTree := false
	if ok {
		removeTree = addDiffToTree(t, &diff)
		d.Trees[t.BasePath] = TreeDiff{}
	}
	// If we need to remove the tree, signal the previous level of recursion
	if removeTree {
		return true
	}

	// Go through this tree `t`'s `File`s, apply any diffs, assign the modified files
	filesAfterAddingDiff := []tree.File{}
	for _, f := range t.Files {
		removeFile := false
		fDiff, ok := d.Files[f.Name]
		if ok {
			removeFile, _ = addDiffToFile(&f, &fDiff, &d.AllHash, newTreeAllHash)
			d.Files[f.Name] = FileDiff{}
		}

		if !removeFile {
			filesAfterAddingDiff = append(filesAfterAddingDiff, f)
		}
	}
	t.Files = filesAfterAddingDiff

	// Go through this tree `t`'s Subtrees, apply any diffs, assign the modified trees
	t.SizeDirect = 0
	t.SizeBelow = 0
	t.LastModifiedDirect = time.Time{}
	t.LastModifiedBelow = time.Time{}
	t.NumFilesDirect = int64(len(t.Files))
	t.NumFilesBelow = int64(len(t.Files))
	for _, f := range t.Files {
		t.LastModifiedDirect = utility.GetNewestTime(t.LastModifiedDirect, f.LastModified)
		t.LastModifiedBelow = utility.GetNewestTime(t.LastModifiedBelow, f.LastModified)
		t.SizeDirect += f.Size
	}
	t.SizeBelow = t.SizeDirect

	newSubTrees := []tree.FileTree{}
	for _, st := range t.SubTrees {
		removeTree = false
		removeTree = WalkAddTreeDiff(&st, d, newTreeAllHash, addedTrees, addedFiles)
		if !removeTree {
			newSubTrees = addFileTreeInAlphaOrder(newSubTrees, st)
		}
		t.LastModifiedBelow = utility.GetNewestTime(t.LastModifiedBelow, st.LastModifiedBelow)
		t.NumFilesBelow += st.NumFilesBelow
		t.SizeBelow += st.SizeBelow
	}
	t.SubTrees = newSubTrees

	return false
}

/*
Add a `TreeDiff` to a `FileTree`
*/
func addDiffToTree(t *tree.FileTree, d *TreeDiff) bool {
	if d.Empty() {
		return false
	}

	switch d.Type {
	case renamed:
		t.BasePath = d.NewerPath
	case removed:
		return true
	case modified:
		fallthrough
	case added:
		t.Comprehensive = d.Comprehensive
		t.BasePath = d.NewerPath
		t.Depth += d.DepthDiff
		t.ErrStrings = append(t.ErrStrings, d.ErrStringsDiff...)
		if t.LastVisited.Equal(time.Time{}) {
			t.LastVisited = utility.GoSpecialTime.Add(d.LastVisitedDiff)
		} else {
			t.LastVisited = t.LastVisited.Add(d.LastVisitedDiff)
		}
		if t.LastModifiedDirect.Equal(time.Time{}) {
			t.LastModifiedDirect = utility.GoSpecialTime.Add(d.LastVisitedDiff)
		} else {
			t.LastModifiedDirect = t.LastModifiedDirect.Add(d.LastVisitedDiff)
		}
		t.SizeDirect += d.SizeDiffDirect
		t.NumFilesDirect += d.NumFilesTotalDiffDirect
		t.AllHashOffset = d.AllHashOffset
		t.TimeTaken += d.TimeTakenDiff
		if t.Depth > 0 {
			t.AllHash = d.AllHash
		}
	default:
	}

	return false
}

/*
Adds a `FileDiff` to a `File`, also copies the file's SHA256 hash (if present),
to the `AllHash` in the new tree
*/
func addDiffToFile(f *tree.File, d *FileDiff, diffAllHash, newTreeAllHash *[]byte) (removeFile bool, diffEmpty bool) {
	if d.Empty() {
		return false, true
	}

	switch d.Type {
	case renamed:
		f.Name = d.NewerName
	case removed:
		return true, false
	case modified:
		fallthrough
	case added:
		f.Name = d.NewerName
		f.Err = d.NewerErr
		if f.LastModified.Equal(time.Time{}) {
			f.LastModified = utility.GoSpecialTime.Add(d.LastModifiedDiff)
		} else {
			f.LastModified = f.LastModified.Add(d.LastModifiedDiff)
		}
		f.Size += d.SizeDiff
		f.Hash = utility.InitialiseHashLocation()
		if d.HashDiff.HashOffset > -1 {
			f.Hash = utility.CopyHashToNewArray(d.HashDiff, diffAllHash, newTreeAllHash)
		}
	default:
	}

	return false, false
}

/*
For a given path, and each index between `firstIdx` and `lastIdx` (of recorded diffs)
accumulate the "diff"s into a single "diff"

TODO: I think this is missing some error conditions
*/
func AddDiffsForPath(path string, firstIdx, lastIdx int) (ScanDiff, error) {
	ret := ScanDiff{
		AllHash: []byte{},
		Trees:   map[string]TreeDiff{},
		Files:   map[string]FileDiff{},
	}

	for i := firstIdx; i <= lastIdx; i++ {
		// Load diff from disk
		diffPath := config.GetScansOutputDir() + getScanFilename(path, i, true)
		diff, err := ReadBinary(diffPath)
		if err != nil {
			return ret, errorx.Decorate(err, "failed to read diff from file at index %d", i)
		}

		// Add diff
		ret.AddDiff(diff, &ret.AllHash)
	}

	return ret, nil
}

// NOTE: Copied/modified from `GetScanFilename` in `records` to avoid an "import cycle" for now
func getScanFilename(rootPath string, index int, isDiff bool) string {
	return fmt.Sprintf("%s_%d.diff", utility.HashFilePath(rootPath), index)
}

func addFileInAlphaOrder(existing []tree.File, new tree.File) []tree.File {
	i := 0
	f := tree.File{}
	foundPostion := false
	for i, f = range existing {
		cmp := strings.Compare(new.Name, f.Name)
		if cmp < 0 {
			foundPostion = true
			break
		}
	}

	if !foundPostion {
		return append(existing, new)
	}

	existingCopy := make([]tree.File, len(existing))
	copy(existingCopy, existing)

	var (
		beforeNewElem = existing[:i]
		afterNewElem  = existingCopy[i:]
	)

	ret := append(beforeNewElem, []tree.File{new}...)
	ret = append(ret, afterNewElem...)

	return ret
}

func addFileTreeInAlphaOrder(existing []tree.FileTree, new tree.FileTree) []tree.FileTree {
	i := 0
	t := tree.FileTree{}
	foundPostion := false
	for i, t = range existing {
		cmp := strings.Compare(new.BasePath, t.BasePath)
		if cmp < 0 {
			foundPostion = true
			break
		}
	}

	if !foundPostion {
		return append(existing, new)
	}

	existingCopy := make([]tree.FileTree, len(existing))
	copy(existingCopy, existing)

	var (
		beforeNewElem = existing[:i]
		afterNewElem  = existingCopy[i:]
	)
	ret := append(beforeNewElem, []tree.FileTree{new}...)
	ret = append(ret, afterNewElem...)

	return ret
}
