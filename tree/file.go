package tree

import (
	"encoding/gob"
	"encoding/json"
	"os"

	"github.com/joomcode/errorx"
)

// TODO: Should revisit these, JSON (although easier to read) isn't as efficient as a binary file format (review `gob` library for recursive data structures)
func WriteBinary(tree any, path string) error {
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
		return tree, errorx.Decorate(err, "failed to open file for readgin FileTree data")
	}
	defer f.Close()

	gd := gob.NewDecoder(f)
	err = gd.Decode(tree)

	return tree, err
}

func WriteJSON(tree any, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errorx.Decorate(err, "failed to opem/create file for writing FileTree data")
	}
	defer f.Close()

	je := json.NewEncoder(f)
	return je.Encode(tree)
}

func ReadJSON(path string) (FileTree, error) {
	tree := FileTree{}
	f, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		return tree, errorx.Decorate(err, "failed to open file for readgin FileTree data")
	}
	defer f.Close()

	jd := json.NewDecoder(f)
	err = jd.Decode(tree)

	return tree, err
}
