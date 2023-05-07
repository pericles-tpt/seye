package tree

import (
	"fmt"
	"os"
)

func Walk(tree *FileTree) {
	ents, err := os.ReadDir(tree.BasePath)
	tree.Err = err

	for _, e := range ents {
		if e.IsDir() {
			subTree := FileTree{
				BasePath: fmt.Sprintf("%s/%s", tree.BasePath, e.Name()),
			}
			Walk(&subTree)
			tree.SubTrees = append(tree.SubTrees, subTree)
		} else {
			f, err := e.Info()
			if err != nil {
				tree.Files = append(tree.Files, File{"", 0, err})
			} else {
				tree.Files = append(tree.Files, File{f.Name(), f.Size(), err})
				tree.Size += int(f.Size())
			}

		}
	}
}

// TODO: Delete this?
func PrintTree(t *FileTree) {
	if t == nil {
		return
	}
	fmt.Printf("On tree with path: %s, size at this level is: %d\n", t.BasePath, t.Size)

	for _, f := range t.Files {
		fmt.Printf("	Has file %s of size: %d\n", f.Name, f.Size)
	}

	for _, t := range t.SubTrees {
		PrintTree(&t)
	}
}
