package tree

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"strings"

	"github.com/joomcode/errorx"
)

/*
Using `gob` to do a basic deep copy of a tree

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
		return errorx.Decorate(err, "failed to open/create file for writing FileTree data")
	}
	defer f.Close()

	ge := gob.NewEncoder(f)
	return ge.Encode(tree)
}

func ReadBinary(path string) (FileTree, error) {
	var tree FileTree
	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return tree, errorx.Decorate(err, "failed to open file for reading FileTree data")
	}
	defer f.Close()

	gd := gob.NewDecoder(f)
	err = gd.Decode(&tree)
	return tree, err
}

func getFullPath(dir, basepath string) string {
	fullPath := fmt.Sprintf("%s/%s", dir, basepath)
	if strings.HasSuffix(dir, "/") {
		fullPath = fmt.Sprintf("%s%s", dir, basepath)
	}
	return fullPath
}
