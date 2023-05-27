package tree

import (
	"bytes"
	"encoding/gob"
	"os"

	"github.com/joomcode/errorx"
)

/*
A basic deep copy implementation, using `gob`

(useful when you want to modify an existing tree by adding a `ScanDiff` to it)
*/
func (tree *FileTree) DeepCopy() FileTree {
	var (
		b       bytes.Buffer
		newTree FileTree
	)
	ge := gob.NewEncoder(&b)
	ge.Encode(tree)

	gd := gob.NewDecoder(&b)
	gd.Decode(&newTree)

	return newTree
}

func (tree *FileTree) WriteBinary(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to opem/create file for writing FileTree data")
	}
	defer f.Close()

	ge := gob.NewEncoder(f)
	return ge.Encode(tree)
}

func ReadBinary(path string) (FileTree, error) {
	tree := FileTree{}
	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return tree, errorx.Decorate(err, "failed to open file for reading FileTree data")
	}
	defer f.Close()

	gd := gob.NewDecoder(f)
	err = gd.Decode(&tree)
	return tree, err
}
