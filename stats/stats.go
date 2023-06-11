package stats

import (
	"path"
	"sort"
)

/*
Add a `BasicFile`, in decreasing size order, to the array of largest files up
to `largestFilesLimit` (largest at index 0)
*/
func (w *WalkStats) UpdateLargestFiles(file BasicFile) {
	if w.LargestFiles == nil {
		return
	} else if len(*w.LargestFiles) == 0 {
		*w.LargestFiles = append(*w.LargestFiles, file)
		return
	}

	// Start from the bottom find the first place to insert it (if we can)
	var (
		thisFileSize     = file.Size
		smallerFileIndex = -1
		largerFileIndex  = -1
	)

	for i := range *w.LargestFiles {
		currIndex := len(*w.LargestFiles) - 1 - i
		if currIndex < 0 {
			break
		}

		currFile := (*w.LargestFiles)[currIndex]
		if thisFileSize > currFile.Size {
			smallerFileIndex = currIndex
		} else if thisFileSize <= currFile.Size {
			largerFileIndex = currIndex
		}

		var (
			largerFileFound        = largerFileIndex > -1
			smallerFileFound       = smallerFileIndex > -1
			withinLargesFilesLimit = len((*w.LargestFiles)) < largestFilesLimit
			newLargeFile           = file
		)

		if largerFileFound && !smallerFileFound && withinLargesFilesLimit {
			(*w.LargestFiles) = append((*w.LargestFiles), newLargeFile)
			break
		} else if largerFileFound && smallerFileIndex > -1 {
			(*w.LargestFiles)[smallerFileIndex] = newLargeFile
			break
		} else if smallerFileFound && currIndex == 0 {
			// Prepend to the start
			newLargestFiles := []BasicFile{newLargeFile}
			(*w.LargestFiles) = append(newLargestFiles, (*w.LargestFiles)...)
			if len((*w.LargestFiles)) > largestFilesLimit {
				*w.LargestFiles = (*w.LargestFiles)[:largestFilesLimit-1]
			}
			break
		}
	}
}

/*
Add a `BasicFile` record to the `DuplicateMap`
*/
func (w *WalkStats) UpdateDuplicates(fileHashBytes []byte, fileSize int64, filePath string) {
	if w.DuplicateMap == nil {
		return
	}

	_, ok := (*w.DuplicateMap)[string(fileHashBytes)]
	if !ok {
		(*w.DuplicateMap)[string(fileHashBytes)] = []BasicFile{{Path: filePath, Size: fileSize}}
	} else {
		(*w.DuplicateMap)[string(fileHashBytes)] = append((*w.DuplicateMap)[string(fileHashBytes)], BasicFile{
			Path: filePath,
			Size: fileSize,
		})
	}
}

/*
Sort the stored duplicates and return them in sorted order (0 -> n, largest -> smallest)
duplicate size = num_duplicates * duplicate_size
*/
func (w *WalkStats) GetLargestDuplicates(limit int) [][]BasicFile {
	var orderedDuplicates [][]BasicFile

	// Remove anything that isn't a duplicate
	for _, v := range *w.DuplicateMap {
		if len(v) > 1 {
			v[0].Path = path.Base(v[0].Path)
			orderedDuplicates = append(orderedDuplicates, v)
		}
	}

	sort.SliceStable(orderedDuplicates[:], func(i, j int) bool {
		aSize := int64(len(orderedDuplicates[i])) * (orderedDuplicates[i][0].Size)
		bSize := int64(len(orderedDuplicates[j])) * (orderedDuplicates[j][0].Size)
		return aSize > bSize
	})

	return orderedDuplicates
}
