package file

import "time"

type HashType int

const (
	SHA256 HashType = iota
	// TODO: MD5? Smaller, faster, less secure...
)

/*
Encapsulates size information for a file in the FS
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
	HashOffset int
	HashLength int
	// TODO: When we try to implement MT we can have a map[string][]int as a global
	//		 which contains the path(s) -> AllHashIndex offsets for each thread, we
	//		 can use this value to index it. The value can be set during 'Walk'
	// HashOffsetIndex int
}

func InitialiseHashLocation(Offset *int, Type *HashType, Length *int) HashLocation {
	ret := HashLocation{
		HashOffset: -1,
	}

	if Offset != nil {
		ret.HashOffset = *Offset
	}
	if Type != nil {
		ret.Type = *Type
	}
	if Length != nil {
		ret.HashLength = *Length
	}

	return ret
}
