package tree

type FileTree struct {
	BasePath string
	Files    []File
	SubTrees []FileTree
	Size     int
	Err      error
}

type File struct {
	Name string
	Size int64
	Err  error
}

// type Snapshot struct {
// 	Dirs      []Dir
// 	FileCount int
// 	TotalSize int
// }

// type Dir struct {
// 	Path  string
// 	Files []File
// }

// type File struct {
// 	Name string
// 	Size *int64
// 	Err  error
// }
