package main

func main() {
	s1 := FileTree{
		BasePath: "/home/",
	}

	Walk(&s1)

	// time.Sleep(time.Second * 30)

	// s2 := FileTree{
	// 	BasePath: "/home/",
	// }

	// Walk(&s2)

	// PrintTree(&s1)

	// CompareTrees(&s1, &s2, nil)
}
