package utility

import (
	"time"
)

func GetNewestTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}

func Contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}
