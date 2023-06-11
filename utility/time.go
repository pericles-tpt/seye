package utility

import (
	"time"
)

var (
	GoSpecialTime, _ = time.Parse("2006-01-02 15:05:05.999999999 -0700 MST", "1970-01-01 00:00:00.000000000 +1000 AEST")
)

func GetNewestTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}
