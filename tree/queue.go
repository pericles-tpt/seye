package tree

// push/pop, front/back operations for 1D arrays
func pushFront1D(f []FileTree, e FileTree) []FileTree {
	return append([]FileTree{e}, f[:]...)
}

func popFront1D(f []FileTree) ([]FileTree, FileTree) {
	var (
		arr  []FileTree
		elem FileTree
	)
	if len(f) > 0 {
		return f[1:], f[0]
	}
	return arr, elem
}

func pushBack1D(f []FileTree, e FileTree) []FileTree {
	return append(f, e)
}

func popBack1D(f *[]FileTree) FileTree {
	if len(*f) > 0 {
		lastIdx := len(*f) - 1
		lastTree := (*f)[lastIdx]
		*f = (*f)[:lastIdx]
		return lastTree
	}
	return FileTree{}
}

// push/pop, front/back operations for 2D arrays
func pushFront2D(f []FileTree, e FileTree) []FileTree {
	return append([]FileTree{e}, f[:]...)
}

func popFront2D(f [][]FileTree, depth int) ([][]FileTree, FileTree) {
	var (
		arr  [][]FileTree
		elem FileTree
	)
	if len(f[depth]) > 0 {
		firstElem := f[depth][0]
		f[depth] = f[depth][1:]
		return f, firstElem
	}
	return arr, elem
}

func pushBack2D(f [][]FileTree, e FileTree, depth int) [][]FileTree {
	f[depth] = append(f[depth], e)
	return f
}
