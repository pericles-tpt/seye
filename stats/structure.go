package stats

const (
	largestFilesLimit = 100
)

type WalkStats struct {
	LargestFiles  *[]LargeFile
	DuplicateMap  *map[string][]string
	NumDuplicates int
}

type LargeFile struct {
	FullName string
	Size     int64
}
