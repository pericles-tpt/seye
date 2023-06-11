package utility

func Contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

/*
Get all strings that are DUPLICATES of strings in `a` OR only exist in `b`
*/
func AdditionalStringsInB(a, b []string) []string {
	aMap := map[string]int{}
	for _, s := range a {
		aMap[s] = -1
	}

	for _, s := range b {
		_, ok := aMap[s]
		if ok {
			aMap[s] += 1
		} else {
			aMap[s] = 1
		}
	}

	ret := []string{}
	for s, n := range aMap {
		if n > 0 {
			for i := 0; i < n; i++ {
				ret = append(ret, s)
			}
		}
	}

	return ret
}
