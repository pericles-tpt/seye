package utility

import (
	"crypto/md5"
	"fmt"
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

func HashFilePath(input string) string {
	hashed := md5.Sum([]byte(input))
	legalFileName := ""
	for _, v := range hashed {
		legalFileName += fmt.Sprintf("%x", v)
	}
	return legalFileName
}
