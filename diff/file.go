package diff

import (
	"encoding/gob"
	"os"

	"github.com/joomcode/errorx"
)

func (d *ScanDiff) WriteBinary(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to open/create file for writing ScanDiff data")
	}
	defer f.Close()

	ge := gob.NewEncoder(f)
	return ge.Encode(&d)
}

func ReadBinary(path string) (ScanDiff, error) {
	scanDiff := ScanDiff{}
	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return scanDiff, errorx.Decorate(err, "failed to open file for readgin FileTree data")
	}
	defer f.Close()

	gd := gob.NewDecoder(f)
	err = gd.Decode(&scanDiff)

	return scanDiff, err
}
