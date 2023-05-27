package stats

const (
	largestFilesLimit = 100
)

type WalkStats struct {
	LargestFiles *[]BasicFile
	DuplicateMap *map[string][]BasicFile
}

type BasicFile struct {
	Path string
	Size int64
}
