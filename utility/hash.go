package utility

import (
	"bytes"
	"crypto/md5"
	"fmt"
)

type HashType int

const (
	SHA256 HashType = iota
	// TODO: MD5? Smaller, faster, less secure...
)

/*
File hashes are stored in a contiguous []byte, this structure
lets us access a location in that array for parallels r/w
*/
type HashLocation struct {
	Type       HashType
	HashOffset int
	HashLength int
}

/*
Initialises a `HashLocations` with `HashOffset` = -1 as the "default" `HashLocation`
*/
func InitialiseHashLocation() HashLocation {
	return HashLocation{
		HashOffset: -1,
	}
}

/*
Checks if hashes in two locations have the same bytes, at the same relative indices
*/
func HashesEqual(a, b HashLocation, allHashesA, allHashesB *[]byte) bool {
	if a.HashOffset == -1 && b.HashOffset == -1 {
		return true
	} else if a.HashOffset > -1 && b.HashOffset > -1 {
		bytesA := (*allHashesA)[a.HashOffset : a.HashOffset+a.HashLength]
		bytesB := (*allHashesB)[b.HashOffset : b.HashOffset+b.HashLength]
		return (a.Type == b.Type) && bytes.Equal(bytesA, bytesB)
	}

	// -> one has a hash, the other doesn't -> the file has changed
	return false
}

/*
Adds a hash's bytes, to a larger array of bytes
*/
func AddHashAtOffset(offset int, length int, hashType HashType, hashBytes []byte, AllHash *[]byte) HashLocation {
	if (offset + length - 1) > len(*AllHash)-1 {
		return InitialiseHashLocation()
	}

	for i := 0; i < length; i++ {
		(*AllHash)[offset+i] = hashBytes[i]
	}
	return HashLocation{
		HashOffset: offset,
		HashLength: length,
		Type:       hashType,
	}
}

/*
Copies a hash from an old *[]byte to a new *[]byte
*/
func CopyHashToNewArray(addFromLocation HashLocation, fromAllHash, toAllHash *[]byte) HashLocation {
	newHashOffset := len(*toAllHash) - 1
	*toAllHash = append(*toAllHash, (*fromAllHash)[addFromLocation.HashOffset:addFromLocation.HashOffset+addFromLocation.HashLength]...)
	return HashLocation{
		HashOffset: newHashOffset,
		HashLength: addFromLocation.HashLength,
		Type:       addFromLocation.Type,
	}
}

/*
Hashes a filepath (using MD5), then prints each byte to a string in
hexadecimal so that the filename doesn't contain illegal characters
*/
func HashFilePath(input string) string {
	hashed := md5.New().Sum([]byte(input))
	legalFileName := ""
	for _, v := range hashed {
		legalFileName += fmt.Sprintf("%x", v)
	}
	return legalFileName
}
