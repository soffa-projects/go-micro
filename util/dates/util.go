package dates

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func NowPrt() *time.Time {
	value := time.Now().UTC()
	return &value
}
