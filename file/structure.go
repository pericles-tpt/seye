package file

import "time"

type HashType int

const (
	NONE HashType = iota
	SHA256
	RAW
)

/*
	File: Encapsulates size information for a file in the FS

	NOTE: `ByteSample` is only populated for the INITAL and "Comprehensive"
	scans
*/
type File struct {
	Name         string
	Hash         HashLocation
	Size         int64
	Err          string
	LastModified time.Time
}

type HashLocation struct {
	Type       HashType
	HashOffset *int
	HashLength int
}
