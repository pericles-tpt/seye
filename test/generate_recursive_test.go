package test

import (
	"os"
	"testing"
	"time"

	"github.com/Fiye/tree"
	"github.com/Fiye/utility"
)

func TestGenerateRPopulatedDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}

	populatedDir := cwd + "/testDir"
	populatedDirTree := tree.WalkGenerateTreeRecursive(populatedDir, 0, false, nil)
	var (
		timeA, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-10 14:41:08.899366352 +1000 AEST")
		timeB, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-10 14:41:08.899366352 +1000 AEST")
		timeC, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-10 14:41:08.899366352 +1000 AEST")
		timeD, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:52:02.937675614 +1000 AEST")
		timeE, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:53:06.852119727 +1000 AEST")
		timeF, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:53:07.876436108 +1000 AEST")
		timeG, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:53:08.974341879 +1000 AEST")
		timeH, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:52:30.76262004 +1000 AEST")
		timeI, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:52:44.737678331 +1000 AEST")
		timeJ, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:53:08.974341879 +1000 AEST")
		timeK, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2023-06-09 18:53:08.974341879 +1000 AEST")
		expTree  = tree.FileTree{
			Comprehensive:      false,
			BasePath:           "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir",
			LastVisited:        populatedDirTree.LastVisited,
			TimeTaken:          populatedDirTree.TimeTaken,
			LastModifiedDirect: timeA,
			SizeDirect:         79,
			NumFilesDirect:     3,
			LastModifiedBelow:  timeB,
			SizeBelow:          105,
			NumFilesBelow:      7,
			Files: []tree.File{
				{
					Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A1",
					Hash: utility.HashLocation{
						HashOffset: -1,
					},
					Size:         44,
					LastModified: timeC,
				},
				{
					Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B1",
					Hash: utility.HashLocation{
						HashOffset: -1,
					},
					LastModified: timeD,
				},
				{
					Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/C1",
					Hash: utility.HashLocation{
						HashOffset: -1,
					},
					Size:         35,
					LastModified: timeE,
				},
			},
			SubTrees: []tree.FileTree{{
				Comprehensive: false,
				BasePath:      "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A",
				Files: []tree.File{
					{
						Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/A",
						Hash: utility.HashLocation{
							HashOffset: -1,
						},
						Size:         11,
						LastModified: timeF,
					},
					{
						Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/B",
						Hash: utility.HashLocation{
							HashOffset: -1,
						},
						Size:         6,
						LastModified: timeG,
					},
					{
						Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/C",
						Hash: utility.HashLocation{
							HashOffset: -1,
						},
						LastModified: timeH,
					},
					{
						Name: "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/A/D",
						Hash: utility.HashLocation{
							HashOffset: -1,
						},
						Size:         9,
						LastModified: timeI,
					},
				},
				LastVisited:        populatedDirTree.SubTrees[0].LastVisited,
				TimeTaken:          populatedDirTree.SubTrees[0].TimeTaken,
				Depth:              1,
				LastModifiedDirect: timeJ,
				SizeDirect:         26,
				NumFilesDirect:     4,
				LastModifiedBelow:  timeK,
				SizeBelow:          26,
				NumFilesBelow:      4,
			},
				{
					Comprehensive: false,
					BasePath:      "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/B",
					LastVisited:   populatedDirTree.SubTrees[1].LastVisited,
					TimeTaken:     populatedDirTree.SubTrees[1].TimeTaken,
					Depth:         1,
				},
				{
					Comprehensive: false,
					BasePath:      "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/C",
					LastVisited:   populatedDirTree.SubTrees[2].LastVisited,
					TimeTaken:     populatedDirTree.SubTrees[2].TimeTaken,
					Depth:         1,
				},
				{
					Comprehensive: false,
					BasePath:      "/Users/ptelemachou/Code/Work/Go_Projects/src/github.com/pericles-tpt/Fiye/test/testDir/D",
					LastVisited:   populatedDirTree.SubTrees[3].LastVisited,
					TimeTaken:     populatedDirTree.SubTrees[3].TimeTaken,
					Depth:         1,
				},
			},
		}
	)

	notEqualReason := (*populatedDirTree).Equal(expTree)
	if notEqualReason != nil {
		t.Error("tree for populate dir NOT equal to expected `populatedDirTree`, reason: ", notEqualReason)
	}
}

func TestGenerateRNonexistentDir(t *testing.T) {
	invalidPath := "/invalidPath"
	invalidPathTree := tree.WalkGenerateTreeRecursive(invalidPath, 0, false, nil)
	expTree := tree.FileTree{
		BasePath:   invalidPath,
		ErrStrings: []string{"failed to open `tree.BasePath`, cause: open /invalidPath: no such file or directory"},
	}
	notEqualReason := (*invalidPathTree).Equal(expTree)
	if notEqualReason != nil {
		t.Error("tree for invalid path NOT equal to empty `FileTree`, reason: ", notEqualReason)
	}
}

func TestGenerateREmptyDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error("failed to get cwd", err)
	}

	emptyDirPath := cwd + "/testDir/B"
	emptyDirTree := tree.WalkGenerateTreeRecursive(emptyDirPath, 0, false, nil)
	expTree := tree.FileTree{
		BasePath:      emptyDirPath,
		Comprehensive: false,
		LastVisited:   emptyDirTree.LastVisited,
		TimeTaken:     emptyDirTree.TimeTaken,
		Depth:         0,
	}
	notEqualReason := (*emptyDirTree).Equal(expTree)
	if notEqualReason != nil {
		t.Error("tree for invalid path NOT equal to expected `emptyDirTree`, reason: ", notEqualReason)
	}
}
