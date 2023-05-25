package diff

import (
	"encoding/gob"
	"os"

	"github.com/joomcode/errorx"
)

func (d *TreeDiff) WriteBinary(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to opem/create file for writing FileTree data")
	}
	defer f.Close()

	ge := gob.NewEncoder(f)
	return ge.Encode(&d)
}

func (d *DiffMaps) WriteBinary(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to opem/create file for writing FileTree data")
	}
	defer f.Close()

	ge := gob.NewEncoder(f)
	return ge.Encode(&d)
}

func ReadBinary(path string) (TreeDiff, error) {
	treeDiff := TreeDiff{}
	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return treeDiff, errorx.Decorate(err, "failed to open file for readgin FileTree data")
	}
	defer f.Close()

	gd := gob.NewDecoder(f)
	err = gd.Decode(&treeDiff)

	return treeDiff, err
}
