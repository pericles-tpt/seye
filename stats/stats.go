package stats

import "io/fs"

func (w *WalkStats) UpdateLargestFiles(filePath string, fInfo fs.FileInfo) {
	if w.LargestFiles == nil {
		return
	} else if len(*w.LargestFiles) == 0 {
		*w.LargestFiles = append(*w.LargestFiles, BasicFile{Path: filePath, Size: fInfo.Size()})
		return
	}

	// Start from the bottom find the first place to insert it (if we can)
	var (
		thisFileSize     = fInfo.Size()
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
			newLargeFile           = BasicFile{
				Path: filePath,
				Size: fInfo.Size(),
			}
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
