package stats

const (
	// TODO: Move this hardcoded value to a config
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
