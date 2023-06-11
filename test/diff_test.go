package test

import (
	"os"
	"testing"

	"github.com/Fiye/diff"
	"github.com/Fiye/tree"
	"github.com/Fiye/utility"
	"github.com/davecgh/go-spew/spew"
)

/*
	TODO: There are some weird issues here:
	- For some reason running two of the MT walks in one test, leads to some elements in the tree being tested when they're unpopulated -> test failure
	- If we do a "Recursive", then "MT Iterative" Walk it's fine (in either order)
	- We can't run all tests when we're using 1 recursive and 1 "iterative MT" walk, but we can run them individually...
	Must be an issue with testing MT functions? Idk what...
*/

// 1. Diff no changes
func TestDiffNoChange(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}

	treeA := tree.WalkGenerateTreeRecursive(cwd+"/testDir", 0, false, nil)
	treeB := tree.WalkTreeIterativeFile(cwd+"/testDir", 0, false, nil)

	diff := diff.CompareTrees(treeA, treeB)
	if !diff.Empty() {
		spew.Dump(treeA)
		spew.Dump(treeB)
		t.Error("changed found between same `treeA` and `treeB` when there should be none")
	}
}

// 2. Diff added file at start (alpha)
func TestDiffAddFileStart(t *testing.T) {

	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}
	newFilePath := cwd + "/testDir/0"

	originalTree := tree.WalkGenerateTreeRecursive(cwd+"/testDir", 0, false, nil)

	err = os.WriteFile(newFilePath, []byte("Test"), 0400)
	if err != nil {
		t.Error("failed to create file for `TestDiffAddFileStart`")
	}
	defer os.Remove(newFilePath)

	fileAddedTree := tree.WalkTreeIterativeFile(cwd+"/testDir", 0, false, nil)

	d := diff.CompareTrees(originalTree, fileAddedTree)
	var (
		expDiff = diff.ScanDiff{
			AllHash: []byte{},
			Trees: map[string]diff.TreeDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir": {
					DiffCompleted: d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].DiffCompleted,
					Comprehensive: false,
					NewerPath:     "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir",
					FilesDiff: []diff.FileDiff{
						diff.FileDiff{
							NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/0",
							Type:      4,
							HashDiff: utility.HashLocation{
								HashOffset: -1,
							},
							SizeDiff:         4,
							LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/0"].LastModifiedDiff,
						},
					},
					LastVisitedDiff:         d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].LastVisitedDiff,
					TimeTakenDiff:           d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].TimeTakenDiff,
					LastModifiedDiffDirect:  d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].LastModifiedDiffDirect,
					SizeDiffDirect:          4,
					NumFilesTotalDiffDirect: 1,
				},
			},
			Files: map[string]diff.FileDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/0": {
					NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/0",
					Type:      4,
					HashDiff: utility.HashLocation{
						HashOffset: -1,
					},
					SizeDiff:         4,
					LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/0"].LastModifiedDiff,
				},
			},
		}
	)
	if !d.Equals(expDiff) {
		t.Error("changed found between same `treeA` and `treeB` when there should be none")
	}
}

// 3. Diff added file at end (alpha)
func TestDiffAddFileEnd(t *testing.T) {

	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}
	originalTree := tree.WalkGenerateTreeRecursive(cwd+"/testDir", 0, false, nil)

	newFilePath := cwd + "/testDir/D1"
	err = os.WriteFile(newFilePath, []byte("Test"), 0400)
	if err != nil {
		t.Error("failed to create file for `TestDiffAddFileStart`")
	}
	defer os.Remove(newFilePath)

	fileAddedTree := tree.WalkTreeIterativeFile(cwd+"/testDir", 0, false, nil)

	d := diff.CompareTrees(originalTree, fileAddedTree)
	var (
		expDiff = diff.ScanDiff{
			AllHash: []byte{},
			Trees: map[string]diff.TreeDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir": {
					DiffCompleted: d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].DiffCompleted,
					Comprehensive: false,
					NewerPath:     "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir",
					FilesDiff: []diff.FileDiff{
						diff.FileDiff{
							NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/D1",
							Type:      4,
							HashDiff: utility.HashLocation{
								HashOffset: -1,
							},
							SizeDiff:         4,
							LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/D1"].LastModifiedDiff,
						},
					},
					LastVisitedDiff:         d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].LastVisitedDiff,
					TimeTakenDiff:           d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].TimeTakenDiff,
					LastModifiedDiffDirect:  d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].LastModifiedDiffDirect,
					SizeDiffDirect:          4,
					NumFilesTotalDiffDirect: 1,
				},
			},
			Files: map[string]diff.FileDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/D1": {
					NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/D1",
					Type:      4,
					HashDiff: utility.HashLocation{
						HashOffset: -1,
					},
					SizeDiff:         4,
					LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/D1"].LastModifiedDiff,
				},
			},
		}
	)
	if !d.Equals(expDiff) {
		t.Error("changed found between same `treeA` and `treeB` when there should be none")
	}
}

// 4. Diff added file in middle at ROOT (alpha)
func TestDiffAddFileMiddleRoot(t *testing.T) {

	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}
	originalTree := tree.WalkGenerateTreeRecursive(cwd+"/testDir", 0, false, nil)

	newFilePath := cwd + "/testDir/B12"
	err = os.WriteFile(newFilePath, []byte("Test"), 0400)
	if err != nil {
		t.Error("failed to create file for `TestDiffAddFileStart`")
	}
	defer os.Remove(newFilePath)

	fileAddedTree := tree.WalkTreeIterativeFile(cwd+"/testDir", 0, false, nil)

	d := diff.CompareTrees(originalTree, fileAddedTree)
	var (
		expDiff = diff.ScanDiff{
			AllHash: []byte{},
			Trees: map[string]diff.TreeDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir": {
					DiffCompleted: d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].DiffCompleted,
					Comprehensive: false,
					NewerPath:     "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir",
					FilesDiff: []diff.FileDiff{
						diff.FileDiff{
							NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B12",
							Type:      4,
							HashDiff: utility.HashLocation{
								HashOffset: -1,
							},
							SizeDiff:         4,
							LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B12"].LastModifiedDiff,
						},
					},
					LastVisitedDiff:         d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].LastVisitedDiff,
					TimeTakenDiff:           d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].TimeTakenDiff,
					LastModifiedDiffDirect:  d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir"].LastModifiedDiffDirect,
					SizeDiffDirect:          4,
					NumFilesTotalDiffDirect: 1,
				},
			},
			Files: map[string]diff.FileDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B12": {
					NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B12",
					Type:      4,
					HashDiff: utility.HashLocation{
						HashOffset: -1,
					},
					SizeDiff:         4,
					LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B12"].LastModifiedDiff,
				},
			},
		}
	)
	if !d.Equals(expDiff) {
		t.Error("changed found between same `treeA` and `treeB` when there should be none")
	}
}

// 5. Diff added file in middle at depth=1 (alpha)
func TestDiffAddFileMiddleDeeper(t *testing.T) {

	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}
	originalTree := tree.WalkGenerateTreeRecursive(cwd+"/testDir", 0, false, nil)

	newFilePath := cwd + "/testDir/A/B1"
	err = os.WriteFile(newFilePath, []byte("Test"), 0400)
	if err != nil {
		t.Error("failed to create file for `TestDiffAddFileMiddleDeeper`")
	}
	defer os.Remove(newFilePath)

	fileAddedTree := tree.WalkTreeIterativeFile(cwd+"/testDir", 0, false, nil)

	d := diff.CompareTrees(originalTree, fileAddedTree)
	var (
		expDiff = diff.ScanDiff{
			AllHash: []byte{},
			Trees: map[string]diff.TreeDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A": {
					DiffCompleted: d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A"].DiffCompleted,
					Comprehensive: false,
					NewerPath:     "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A",
					FilesDiff: []diff.FileDiff{
						{
							NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/B1",
							Type:      4,
							HashDiff: utility.HashLocation{
								HashOffset: -1,
							},
							SizeDiff:         4,
							LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/B1"].LastModifiedDiff,
						},
					},
					LastVisitedDiff:         d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A"].LastVisitedDiff,
					TimeTakenDiff:           d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A"].TimeTakenDiff,
					LastModifiedDiffDirect:  d.Trees["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A"].LastModifiedDiffDirect,
					SizeDiffDirect:          4,
					NumFilesTotalDiffDirect: 1,
				},
			},
			Files: map[string]diff.FileDiff{
				"/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/B1": {
					NewerName: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/B1",
					Type:      4,
					HashDiff: utility.HashLocation{
						HashOffset: -1,
					},
					SizeDiff:         4,
					LastModifiedDiff: d.Files["/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/B1"].LastModifiedDiff,
				},
			},
		}
	)

	if !d.Equals(expDiff) {
		t.Error("changed found between same `treeA` and `treeB` when there should be none")
	}
}

// 7. Diff added dir at start (alpha)
// 8. Diff added dir at end (alpha)
// 9. Diff added dir in middle (alpha)
// 10. Diff added dir in middle at root (alpha)
// 11. Diff added dir in middle at depth=3 (alpha)
// 12. Diff removed file
// 13. Diff remove dir
// 14. Diff renamed file
// 15. Diff renamed dir
// 16. Diff changed file, no size diff, comprehensive (alpha)
// 17. Diff changed file, no size diff, shallow (alpha)
// 18. Diff changed file, increased size, comprehensive (alpha)
// 19. Diff changed file, increased size, shallow (alpha)
// 20. Diff changed file, decreased size, comprehensive (alpha)
// 21. Diff changed file, decreased size, shallow (alpha)

// 20. Test diff, "Comprehensive" AND non-"Comprehensive"

// 22. Check for, s1 and s2 (that are different), s1 + diff(s1, s2) == s2
func TestAddDiffToMatchScan(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}

	originalTree := tree.WalkGenerateTreeRecursive(cwd+"/testDir", 0, false, nil)

	newFilePath := cwd + "/testDir/B12"
	err = os.WriteFile(newFilePath, []byte("Test"), 0400)
	if err != nil {
		t.Error("failed to create file for `TestDiffAddFileStart`")
	}
	defer os.Remove(newFilePath)

	fileAddedTree := tree.WalkTreeIterativeFile(cwd+"/testDir", 0, false, nil)

	d := diff.CompareTrees(originalTree, fileAddedTree)

	originalPlusDiff := originalTree.DeepCopy()

	_ = diff.WalkAddTreeDiff(&originalPlusDiff, &d, &originalPlusDiff.AllHash, []diff.TreeDiff{}, []diff.FileDiff{})
	err = originalPlusDiff.Equal(*fileAddedTree)
	if err != nil {
		t.Error("changed found between same `originalPlusDiff` and `fileAddedTree` when there should be none: ", err)
	}
}
